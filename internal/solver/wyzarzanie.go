package solver

import (
	"context"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Faza 3 (docs/plan.md): symulowane wyżarzanie.
//
// Ruchy: przesunięcie terminu (czasem razem ze zmianą sali), zmiana sali, zamiana terminów
// dwóch węzłów o tej samej długości, zmiana wariantu tygodni (parzystość, okno tygodni Z105).
// Gdy plan nie ma naruszeń, przyjmujemy tylko ruchy dopuszczalne. W trybie awaryjnym energia to
// koszt + karaNaruszenia × liczba naruszeń, więc wyżarzanie najpierw usuwa naruszenia (ADR 0006).

const (
	karaNaruszenia     = 10_000.0
	probkiTemperatury  = 300
	temperaturaKoncowa = 0.01
	coIlePostep        = 10_000
)

type wyzarzanie struct {
	pr   *problem
	st   *stan
	k    *koszt
	rng  *rand.Rand
	wagi *Wagi

	starty     [][]kandydat // węzeł → lista (dzień, start) z domeny
	poDlugosci map[int][]int
	naruszenia int

	ruch  []zmiana // bieżący ruch
	stare []kandydat
}

type zmiana struct {
	w  int
	kd kandydat
}

func noweWyzarzanie(pr *problem, st *stan, wagi *Wagi, rng *rand.Rand) *wyzarzanie {
	s := &wyzarzanie{pr: pr, st: st, k: nowyKoszt(pr, st, wagi), rng: rng, wagi: wagi, poDlugosci: make(map[int][]int)}
	s.starty = make([][]kandydat, len(pr.wezly))
	for _, w := range pr.wezly {
		for d := 1; d <= model.LiczbaDni; d++ {
			for sl := 1; sl <= model.LiczbaSlotow; sl++ {
				if w.moznaStart(d, sl) {
					s.starty[w.id] = append(s.starty[w.id], kandydat{dzien: d, start: sl})
				}
			}
		}
		s.poDlugosci[w.dlugosc] = append(s.poDlugosci[w.dlugosc], w.id)
	}
	s.naruszenia = s.policzNaruszenia()
	return s
}

// policzNaruszenia od zera: pary w konflikcie (każda raz) i naruszenia stałe.
func (s *wyzarzanie) policzNaruszenia() int {
	n := 0
	for _, w := range s.pr.wezly {
		n += 2 * w.naruszeniaStale
		s.st.konflikty(w.id, s.st.przyp[w.id], func(int) bool { n++; return true })
	}
	return n / 2
}

// naruszeniaRuchu liczy konflikty z udziałem węzłów ruchu; para węzłów ruchu liczona raz.
func (s *wyzarzanie) naruszeniaRuchu() int {
	n := 0
	for i, z := range s.ruch {
		s.st.konflikty(z.w, s.st.przyp[z.w], func(v int) bool {
			if !slices.ContainsFunc(s.ruch[:i], func(y zmiana) bool { return y.w == v }) {
				n++
			}
			return true
		})
	}
	return n
}

// losujRuch wypełnia s.ruch; false, gdy wylosowany ruch nic nie zmienia.
func (s *wyzarzanie) losujRuch() bool {
	s.ruch = s.ruch[:0]
	w := s.rng.IntN(len(s.pr.wezly))
	wz := s.pr.wezly[w]
	kd := s.st.przyp[w]
	los := s.rng.Float64()
	switch {
	case los < 0.15 && len(wz.warianty) > 1:
		kd.wariant = (kd.wariant + 1 + s.rng.IntN(len(wz.warianty)-1)) % len(wz.warianty)
	case los < 0.35 && len(wz.sale) > 1:
		kd.sala = wz.sale[s.rng.IntN(len(wz.sale))]
	case los < 0.60:
		inne := s.poDlugosci[wz.dlugosc]
		v := inne[s.rng.IntN(len(inne))]
		kv := s.st.przyp[v]
		if v == w || (kv.dzien == kd.dzien && kv.start == kd.start) ||
			!wz.moznaStart(kv.dzien, kv.start) || !s.pr.wezly[v].moznaStart(kd.dzien, kd.start) {
			return false
		}
		kd.dzien, kd.start, kv.dzien, kv.start = kv.dzien, kv.start, kd.dzien, kd.start
		s.ruch = append(s.ruch, zmiana{v, kv})
	default:
		t := s.starty[w][s.rng.IntN(len(s.starty[w]))]
		kd.dzien, kd.start = t.dzien, t.start
		if s.rng.Float64() < 0.3 {
			kd.sala = wz.sale[s.rng.IntN(len(wz.sale))]
		}
	}
	if kd == s.st.przyp[w] && len(s.ruch) == 0 {
		return false
	}
	s.ruch = append(s.ruch, zmiana{w, kd})
	return true
}

// wykonaj stosuje s.ruch i zwraca zmianę kosztu i liczby naruszeń.
func (s *wyzarzanie) wykonaj() (dKoszt float64, dNar int) {
	przed := s.naruszeniaRuchu()
	s.stare = s.stare[:0]
	s.k.poczatekRuchu()
	for _, z := range s.ruch {
		s.stare = append(s.stare, s.st.przyp[z.w])
		s.k.zmienWezel(z.w, z.kd)
	}
	dKoszt = s.k.przelicz()
	return dKoszt, s.naruszeniaRuchu() - przed
}

func (s *wyzarzanie) cofnij(dKoszt float64) {
	for i := len(s.ruch) - 1; i >= 0; i-- {
		s.k.zmienWezel(s.ruch[i].w, s.stare[i])
	}
	s.k.cofnijSkladniki(dKoszt)
}

// temperaturaPoczatkowa: średni wzrost kosztu losowego ruchu dopuszczalnego przyjmujemy z p = 1/2.
func (s *wyzarzanie) temperaturaPoczatkowa() float64 {
	var suma float64
	n := 0
	for i := 0; i < probkiTemperatury*10 && n < probkiTemperatury; i++ {
		if !s.losujRuch() {
			continue
		}
		dk, dn := s.wykonaj()
		s.cofnij(dk)
		if dn == 0 && dk > 0 {
			suma += dk
			n++
		}
	}
	if n == 0 {
		return 1
	}
	return suma / float64(n) / math.Ln2
}

type wynikWyzarzania struct {
	przyp      []kandydat
	koszt      float64
	naruszenia int
}

func lepszy(kosztA float64, narA int, kosztB float64, narB int) bool {
	if narA != narB {
		return narA < narB
	}
	return kosztA < kosztB
}

// uruchom wykonuje iteracje ruchów i zwraca najlepszy znaleziony plan.
// Przerwanie przez ctx kończy wyżarzanie i zwraca najlepszy plan do tej pory.
func (s *wyzarzanie) uruchom(ctx context.Context, iteracje int, postep func(iteracja int, koszt float64, naruszenia int)) wynikWyzarzania {
	best := wynikWyzarzania{przyp: slices.Clone(s.st.przyp), koszt: s.k.suma, naruszenia: s.naruszenia}
	t0 := s.temperaturaPoczatkowa()
	tk := min(temperaturaKoncowa, t0)
	for it := 0; it < iteracje; it++ {
		if it%1024 == 0 && ctx.Err() != nil {
			break
		}
		if postep != nil && it%coIlePostep == 0 {
			postep(it, s.k.suma, s.naruszenia)
		}
		if !s.losujRuch() {
			continue
		}
		temp := t0 * math.Pow(tk/t0, float64(it)/float64(iteracje))
		dk, dn := s.wykonaj()
		przyjmij := false
		switch {
		case s.naruszenia == 0 && dn > 0:
			// Plan dopuszczalny: tylko ruchy dopuszczalne.
		default:
			de := dk + karaNaruszenia*float64(dn)
			przyjmij = de <= 0 || s.rng.Float64() < math.Exp(-de/temp)
		}
		if !przyjmij {
			s.cofnij(dk)
			continue
		}
		s.naruszenia += dn
		if lepszy(s.k.suma, s.naruszenia, best.koszt, best.naruszenia) {
			best.koszt, best.naruszenia = s.k.suma, s.naruszenia
			copy(best.przyp, s.st.przyp)
		}
	}
	if postep != nil {
		postep(iteracje, best.koszt, best.naruszenia)
	}
	return best
}
