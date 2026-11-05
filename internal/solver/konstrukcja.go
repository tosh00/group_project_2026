package solver

import (
	"cmp"
	"context"
	"math/rand/v2"
	"slices"
)

// Faza 2 (docs/plan.md): konstrukcja zachłanna z forward checking, naprawą ślepych zaułków
// i trybem awaryjnym (ADR 0006).
//
//   - Kolejność węzłów: najmniejsza domena, potem najwyższy stopień w grafie, najdłuższe, największe.
//   - Węzeł dostaje wolnego kandydata o najmniejszym koszcie lokalnym, który nie odbiera
//     ostatniego wolnego kandydata żadnemu nieprzypisanemu węzłowi (forward checking).
//   - Ślepy zaułek (brak wolnego kandydata): naprawa wstawia węzeł tam, gdzie wyrzuca najmniej
//     przypisanych węzłów (z tabu na niedawno wstawione), a wyrzucone wracają do kolejki.
//   - Po wyczerpaniu limitu napraw: tryb awaryjny, czyli kandydat o najmniejszej liczbie naruszeń.

const (
	kandydaciFC = 20 // ilu najtańszych kandydatów sprawdzamy forward checkingiem
	tabuNapraw  = 10 // węzeł wstawiony naprawą przez tyle napraw jest chroniony przed wyrzuceniem
	karaTabu    = 100
	szumKosztu  = 0.5 // losowy składnik kosztu: różne seedy dają różne plany
)

// statystykiKonstrukcji opisują przebieg fazy 2 (do testów i logów).
type statystykiKonstrukcji struct {
	naprawy  int
	awaryjne []int // węzły wstawione w trybie awaryjnym
}

type konstrukcja struct {
	pr    *problem
	st    *stan
	rng   *rand.Rand
	wagi  *Wagi
	ranga []int   // ranga[w]: pozycja węzła w kolejności (mniejsza = wcześniej)
	sali  [][]int // sala → węzły, które mają ją w domenie
	stat  statystykiKonstrukcji

	ostatnioNaprawa []int // numer naprawy, w której węzeł był ostatnio wstawiony naprawą (−1 = nigdy)
}

func (pr *problem) kolejnosc() []int {
	ids := make([]int, len(pr.wezly))
	for i := range ids {
		ids[i] = i
	}
	slices.SortStableFunc(ids, func(a, b int) int {
		wa, wb := pr.wezly[a], pr.wezly[b]
		return cmp.Or(
			cmp.Compare(wa.rozmiarDomeny(), wb.rozmiarDomeny()),
			cmp.Compare(len(wb.sasiedzi), len(wa.sasiedzi)),
			cmp.Compare(wb.dlugosc, wa.dlugosc),
			cmp.Compare(wb.osoby, wa.osoby),
		)
	})
	return ids
}

// konstruuj buduje pełny plan: każdy węzeł ma przypisanie. Błąd tylko przy przerwaniu przez ctx.
func (pr *problem) konstruuj(ctx context.Context, rng *rand.Rand, wagi *Wagi, limitNapraw int) (*stan, statystykiKonstrukcji, error) {
	k := &konstrukcja{
		pr: pr, st: nowyStan(pr), rng: rng, wagi: wagi,
		ranga:           make([]int, len(pr.wezly)),
		sali:            make([][]int, len(pr.sale)),
		ostatnioNaprawa: make([]int, len(pr.wezly)),
	}
	kolejnosc := pr.kolejnosc()
	for i, w := range kolejnosc {
		k.ranga[w] = i
	}
	for _, w := range pr.wezly {
		k.ostatnioNaprawa[w.id] = -1
		for _, s := range w.sale {
			k.sali[s] = append(k.sali[s], w.id)
		}
	}

	for krok := 0; k.st.przypisan < len(pr.wezly); krok++ {
		if krok%32 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, k.stat, err
			}
		}
		w := k.nastepny(kolejnosc)
		if kand := k.wolniKandydaci(w); len(kand) > 0 {
			k.st.przypisz(w, k.wybierzFC(w, kand))
			continue
		}
		if k.stat.naprawy < limitNapraw {
			k.napraw(w)
			continue
		}
		k.awaryjnie(w)
	}
	return k.st, k.stat, nil
}

// nastepny zwraca nieprzypisany węzeł o najmniejszej randze.
func (k *konstrukcja) nastepny(kolejnosc []int) int {
	for _, w := range kolejnosc {
		if !k.st.ok[w] {
			return w
		}
	}
	panic("solver: brak nieprzypisanych węzłów")
}

type ocenionyKandydat struct {
	k     kandydat
	koszt float64
}

func (k *konstrukcja) koszt(w int, kd kandydat) float64 {
	return k.pr.kosztLokalny(k.pr.wezly[w], kd, k.wagi) + k.rng.Float64()*szumKosztu
}

// wolniKandydaci to kandydaci bez konfliktów, od najtańszego.
func (k *konstrukcja) wolniKandydaci(w int) []ocenionyKandydat {
	var out []ocenionyKandydat
	k.pr.wezly[w].kandydaci(func(kd kandydat) bool {
		if k.st.wolny(w, kd) {
			out = append(out, ocenionyKandydat{kd, k.koszt(w, kd)})
		}
		return true
	})
	slices.SortFunc(out, func(a, b ocenionyKandydat) int { return cmp.Compare(a.koszt, b.koszt) })
	return out
}

// dotkniete to nieprzypisane węzły, którym przypisanie w do sali może odebrać kandydatów.
func (k *konstrukcja) dotkniete(w, sala int) []int {
	wz := k.pr.wezly[w]
	var out []int
	dodaj := func(vs []int) {
		for _, v := range vs {
			if v != w && !k.st.ok[v] && !slices.Contains(out, v) {
				out = append(out, v)
			}
		}
	}
	dodaj(wz.sasiedzi)
	dodaj(wz.rozneDni)
	dodaj(k.sali[sala])
	return out
}

// wybierzFC wybiera najtańszego kandydata, po którym każdy dotknięty węzeł, który miał wolnego
// kandydata, nadal go ma. Jeśli żaden z kandydaciFC nie przejdzie, bierze najtańszego.
func (k *konstrukcja) wybierzFC(w int, kand []ocenionyKandydat) kandydat {
	przed := make(map[int]bool)
	for i, ok := range kand {
		if i == kandydaciFC {
			break
		}
		k.st.przypisz(w, ok.k)
		dobry := true
		for _, v := range k.dotkniete(w, ok.k.sala) {
			if k.st.maWolnegoKandydata(v) {
				continue
			}
			miał, znane := przed[v]
			if !znane {
				k.st.usun(w)
				miał = k.st.maWolnegoKandydata(v)
				przed[v] = miał
				k.st.przypisz(w, ok.k)
			}
			if miał {
				dobry = false
				break
			}
		}
		k.st.usun(w)
		if dobry {
			return ok.k
		}
	}
	return kand[0].k
}

// napraw wstawia w tam, gdzie wyrzuca najmniej przypisanych węzłów, i wyrzuca je.
func (k *konstrukcja) napraw(w int) {
	najlepszy, najlepszaOcena := kandydat{}, -1.0
	var doWyrzucenia []int
	k.pr.wezly[w].kandydaci(func(kd kandydat) bool {
		var kolizje []int
		k.st.konflikty(w, kd, func(v int) bool {
			if !slices.Contains(kolizje, v) {
				kolizje = append(kolizje, v)
			}
			return true
		})
		ocena := k.koszt(w, kd) / 1000
		for _, v := range kolizje {
			ocena++
			if o := k.ostatnioNaprawa[v]; o >= 0 && k.stat.naprawy-o < tabuNapraw {
				ocena += karaTabu
			}
		}
		if najlepszaOcena < 0 || ocena < najlepszaOcena {
			najlepszy, najlepszaOcena, doWyrzucenia = kd, ocena, kolizje
		}
		return true
	})
	for _, v := range doWyrzucenia {
		k.st.usun(v)
	}
	k.st.przypisz(w, najlepszy)
	k.ostatnioNaprawa[w] = k.stat.naprawy
	k.stat.naprawy++
}

// awaryjnie wstawia w u kandydata z najmniejszą liczbą naruszeń, potem najmniejszym kosztem (ADR 0006).
func (k *konstrukcja) awaryjnie(w int) {
	najlepszy, najmniej, najtaniej := kandydat{}, -1, 0.0
	k.pr.wezly[w].kandydaci(func(kd kandydat) bool {
		n := k.st.liczbaKonfliktow(w, kd)
		if najmniej >= 0 && n > najmniej {
			return true
		}
		c := k.koszt(w, kd)
		if najmniej < 0 || n < najmniej || c < najtaniej {
			najlepszy, najmniej, najtaniej = kd, n, c
		}
		return true
	})
	k.st.przypisz(w, najlepszy)
	k.stat.awaryjne = append(k.stat.awaryjne, w)
}
