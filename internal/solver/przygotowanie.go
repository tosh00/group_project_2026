package solver

import (
	"math/bits"
	"slices"
	"strings"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Faza 1 (docs/plan.md): warianty tygodni, domeny sal i terminów, graf konfliktów.
// Wszystko, co tu liczymy, zależy tylko od danych wejściowych, a nie od planu.

// problem to dane przygotowane dla faz 2 i 3.
type problem struct {
	dane *model.Dane
	ix   *model.Indeks

	sale       []*model.Sala // sale, które występują w jakiejkolwiek domenie; indeks = numer sali
	wezly      []*wezel
	wezelZajec map[string]int // id zajęć → numer węzła
}

// wezel to jednostka planowania: jedne zajęcia albo zestaw ten_sam_termin
// (wspólny dzień, slot i sala, np. Z037 + Z038).
type wezel struct {
	id      int
	zajecia []*model.Zajecia
	dlugosc int // sloty na spotkanie (w zestawie ten_sam_termin najdłuższe)
	osoby   int // największa liczba osób w zestawie

	// sale: numery sal z problem.sale, które spełniają reguły 6 i 7 dla wszystkich zajęć węzła,
	// od najmniejszej pojemności (najlepsze dopasowanie najpierw).
	sale []int
	// starty[d]: maska slotów startowych w dniu d (bit s−1 = slot s), dla których całe spotkanie
	// mieści się w 1–12 i omija wykluczenia prowadzących, Radę Wydziału i blokady grup (reguły 1, 8, 9).
	starty [model.LiczbaDni + 1]uint16
	// warianty tygodni (reguła 2); w zestawie ten_sam_termin iloczyn wariantów członków.
	warianty []wariant

	sasiedzi []int // węzły ze wspólnym prowadzącym albo studentem (reguły 3 i 5)
	rozneDni []int // węzły, które muszą być w innym dniu (reguła 10, rozne_dni)
}

// wariant to wybór tygodni dla każdego zajęcia węzła (tygodnie[i] dla zajecia[i]).
type wariant struct {
	tygodnie []model.Tygodnie
	suma     model.Tygodnie // tygodnie, w których węzeł zajmuje salę, prowadzących i studentów
}

func (w *wezel) nazwa() string {
	ids := make([]string, len(w.zajecia))
	for i, z := range w.zajecia {
		ids[i] = z.ID
	}
	return strings.Join(ids, "+")
}

func (w *wezel) moznaStart(dzien, slot int) bool {
	return slot >= 1 && slot <= model.LiczbaSlotow && w.starty[dzien]&(1<<(slot-1)) != 0
}

func (w *wezel) liczbaStartow() int {
	n := 0
	for d := 1; d <= model.LiczbaDni; d++ {
		n += bits.OnesCount16(w.starty[d])
	}
	return n
}

// rozmiarDomeny to liczba kombinacji (termin, sala, wariant), bez konfliktów z innymi węzłami.
func (w *wezel) rozmiarDomeny() int {
	return w.liczbaStartow() * len(w.sale) * len(w.warianty)
}

// przygotuj buduje problem z danych, które przeszły model.Sprawdz.
// Puste domeny nie są błędem: obsługuje je tryb awaryjny (ADR 0006).
func przygotuj(d *model.Dane) *problem {
	pr := &problem{dane: d, ix: model.NowyIndeks(d), wezelZajec: make(map[string]int)}
	pr.zbudujWezly()
	zablokowane := pr.zablokowaneSloty()
	numerSali := make(map[string]int)
	for _, w := range pr.wezly {
		w.sale = pr.domenaSal(w, numerSali)
		w.starty = pr.domenaStartow(w, zablokowane)
		w.warianty = wariantyWezla(w)
	}
	pr.zbudujGraf()
	return pr
}

func (pr *problem) zbudujWezly() {
	dodaj := func(zs ...*model.Zajecia) {
		w := &wezel{id: len(pr.wezly), zajecia: zs}
		for _, z := range zs {
			w.dlugosc = max(w.dlugosc, z.SlotyNaSpotkanie)
			w.osoby = max(w.osoby, z.LiczbaOsob())
			pr.wezelZajec[z.ID] = w.id
		}
		pr.wezly = append(pr.wezly, w)
	}
	for i := range pr.dane.Zajecia {
		z := &pr.dane.Zajecia[i]
		if _, juz := pr.wezelZajec[z.ID]; juz {
			continue
		}
		if z.TypPowiazania == model.TenSamTermin {
			dodaj(pr.ix.Powiazania[z.Powiazanie]...)
			continue
		}
		dodaj(z)
	}
}

// salaPasuje sprawdza reguły 6 i 7 dla jednych zajęć.
func salaPasuje(z *model.Zajecia, s *model.Sala) bool {
	if len(z.DozwoloneSale) > 0 {
		// Sale z dozwolone_sale wolno użyć nawet wtedy, gdy nie są dostępne do planowania (data/README.md).
		if !slices.Contains(z.DozwoloneSale, s.ID) {
			return false
		}
	} else if !s.DostepnaDoPlanowania || !slices.Contains(z.TypySal, s.Typ) {
		return false
	}
	for _, wyp := range z.WymaganeWyposazenie {
		if !slices.Contains(s.Wyposazenie, wyp) {
			return false
		}
	}
	return s.Pojemnosc >= z.LiczbaOsob() // bez marginesu (ADR 0005)
}

func (pr *problem) domenaSal(w *wezel, numerSali map[string]int) []int {
	var sale []*model.Sala
	for i := range pr.dane.Sale {
		s := &pr.dane.Sale[i]
		ok := true
		for _, z := range w.zajecia {
			ok = ok && salaPasuje(z, s)
		}
		if ok {
			sale = append(sale, s)
		}
	}
	slices.SortStableFunc(sale, func(a, b *model.Sala) int { return a.Pojemnosc - b.Pojemnosc })
	numery := make([]int, len(sale))
	for i, s := range sale {
		n, ok := numerSali[s.ID]
		if !ok {
			n = len(pr.sale)
			numerSali[s.ID] = n
			pr.sale = append(pr.sale, s)
		}
		numery[i] = n
	}
	return numery
}

// zablokowane to sloty niedostępne dla zajęć: maska slotów na dzień (bit s−1 = slot s).
type zablokowane struct {
	prowadzacy map[string]*[model.LiczbaDni + 1]uint16 // dezyderaty i Rada Wydziału (reguła 8)
	grupy      map[string]*[model.LiczbaDni + 1]uint16 // blokady_stale (reguła 9), przypisane grupie z blokady
}

func maskaSlotow(od, do int) uint16 {
	var m uint16
	for s := max(od, 1); s <= min(do, model.LiczbaSlotow); s++ {
		m |= 1 << (s - 1)
	}
	return m
}

func (pr *problem) zablokowaneSloty() zablokowane {
	zb := zablokowane{
		prowadzacy: make(map[string]*[model.LiczbaDni + 1]uint16),
		grupy:      make(map[string]*[model.LiczbaDni + 1]uint16),
	}
	dodaj := func(m map[string]*[model.LiczbaDni + 1]uint16, klucz string, dzien int, sloty uint16) {
		if dzien < 1 || dzien > model.LiczbaDni {
			return
		}
		if m[klucz] == nil {
			m[klucz] = new([model.LiczbaDni + 1]uint16)
		}
		m[klucz][dzien] |= sloty
	}
	for _, t := range pr.dane.DezyderatyTerminy {
		dodaj(zb.prowadzacy, t.Prowadzacy, t.Dzien, maskaSlotow(t.Slot, t.Slot))
	}
	for _, p := range pr.dane.Prowadzacy {
		if !p.CzlonekRadyWydzialu {
			continue
		}
		for _, r := range pr.dane.RadaWydzialu {
			dodaj(zb.prowadzacy, p.ID, r.Dzien, maskaSlotow(r.SlotOd, r.SlotDo))
		}
	}
	for _, b := range pr.dane.BlokadyStale {
		dodaj(zb.grupy, b.Grupa, b.Dzien, maskaSlotow(b.SlotOd, b.SlotDo))
	}
	return zb
}

func (pr *problem) domenaStartow(w *wezel, zb zablokowane) [model.LiczbaDni + 1]uint16 {
	var starty [model.LiczbaDni + 1]uint16
	for dzien := 1; dzien <= model.LiczbaDni; dzien++ {
		var zajete uint16
		for _, z := range w.zajecia {
			for _, p := range z.Prowadzacy {
				if m := zb.prowadzacy[p]; m != nil {
					zajete |= m[dzien]
				}
			}
		}
		for g, m := range zb.grupy {
			if m[dzien] != 0 && pr.kolidujeZGrupa(w, g) {
				zajete |= m[dzien]
			}
		}
		for s := 1; s+w.dlugosc-1 <= model.LiczbaSlotow; s++ {
			if maskaSlotow(s, s+w.dlugosc-1)&zajete == 0 {
				starty[dzien] |= 1 << (s - 1)
			}
		}
	}
	return starty
}

// kolidujeZGrupa: któraś grupa węzła ma wspólnych studentów z grupą g (np. grupą z blokady).
func (pr *problem) kolidujeZGrupa(w *wezel, g string) bool {
	for _, z := range w.zajecia {
		for _, gz := range z.Grupy {
			if pr.ix.GrupyKoliduja(gz, g) {
				return true
			}
		}
	}
	return false
}

// wariantyTygodni zwraca dopuszczalne maski tygodni dla zajęć (reguła 2, ADR 0004).
// Zbiór pusty oznacza dane, dla których żadna maska nie spełnia reguły 2.
func wariantyTygodni(z *model.Zajecia) []model.Tygodnie {
	dozwolone := z.DozwoloneTygodnie()
	n := z.LiczbaSpotkan
	var out []model.Tygodnie
	dodaj := func(t model.Tygodnie) {
		if t.Liczba() == n && !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	switch z.Czestotliwosc {
	case model.CoTydzien:
		dodaj(dozwolone)
	case model.Co2Tygodnie:
		dodaj(dozwolone.Nieparzyste())
		dodaj(dozwolone.Parzyste())
	case model.WybraneTygodnie:
		// Wszystkich podzbiorów może być dużo (Z105: C(14, 9) = 2002), więc na start bierzemy
		// okna kolejnych dozwolonych tygodni. Faza 3 może zmieniać pojedyncze tygodnie.
		lista := dozwolone.Lista()
		for i := 0; i+n <= len(lista); i++ {
			dodaj(model.TygodnieZListy(lista[i : i+n]))
		}
	}
	return out
}

func wariantyWezla(w *wezel) []wariant {
	out := []wariant{{}}
	for _, z := range w.zajecia {
		var nowe []wariant
		for _, wr := range out {
			for _, t := range wariantyTygodni(z) {
				nowe = append(nowe, wariant{
					tygodnie: append(slices.Clone(wr.tygodnie), t),
					suma:     wr.suma | t,
				})
			}
		}
		out = nowe
	}
	return out
}

func (pr *problem) zbudujGraf() {
	for i, a := range pr.wezly {
		for _, b := range pr.wezly[i+1:] {
			if pr.wezlyKoliduja(a, b) {
				a.sasiedzi = append(a.sasiedzi, b.id)
				b.sasiedzi = append(b.sasiedzi, a.id)
			}
		}
	}
	for _, zs := range pr.ix.Powiazania {
		if zs[0].TypPowiazania != model.RozneDni {
			continue
		}
		for _, z1 := range zs {
			w1 := pr.wezly[pr.wezelZajec[z1.ID]]
			for _, z2 := range zs {
				if w2 := pr.wezelZajec[z2.ID]; w2 != w1.id && !slices.Contains(w1.rozneDni, w2) {
					w1.rozneDni = append(w1.rozneDni, w2)
				}
			}
		}
	}
}

func (pr *problem) wezlyKoliduja(a, b *wezel) bool {
	for _, za := range a.zajecia {
		for _, zb := range b.zajecia {
			if model.WspolnyProwadzacy(za, zb) || pr.ix.ZajeciaKolidujaStudenci(za, zb) {
				return true
			}
		}
	}
	return false
}
