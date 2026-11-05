package solver

import (
	"slices"

	"github.com/tosh00/group_project_2026/internal/model"
)

// kandydat to jedno przypisanie węzła: dzień, slot startowy, sala (numer z problem.sale) i wariant tygodni.
type kandydat struct {
	dzien, start, sala, wariant int
}

// kandydaci wywołuje f dla każdego kandydata z domeny węzła (faza 1). f zwraca false, żeby przerwać.
func (w *wezel) kandydaci(f func(k kandydat) bool) {
	for d := 1; d <= model.LiczbaDni; d++ {
		for s := 1; s <= model.LiczbaSlotow; s++ {
			if !w.moznaStart(d, s) {
				continue
			}
			for _, sala := range w.sale {
				for war := range w.warianty {
					if !f(kandydat{d, s, sala, war}) {
						return
					}
				}
			}
		}
	}
}

// stan to częściowy albo pełny plan: przypisania węzłów i zajętość sal.
// Konflikty (reguły 3, 4, 5 i rozne_dni z 10) liczymy przyrostowo względem przypisanych węzłów.
type stan struct {
	pr        *problem
	przyp     []kandydat
	ok        []bool  // węzeł ma przypisanie
	wSali     [][]int // sala → przypisane do niej węzły
	przypisan int
}

func nowyStan(pr *problem) *stan {
	return &stan{
		pr:    pr,
		przyp: make([]kandydat, len(pr.wezly)),
		ok:    make([]bool, len(pr.wezly)),
		wSali: make([][]int, len(pr.sale)),
	}
}

func (st *stan) przypisz(w int, k kandydat) {
	if st.ok[w] {
		st.usun(w)
	}
	st.przyp[w], st.ok[w] = k, true
	st.wSali[k.sala] = append(st.wSali[k.sala], w)
	st.przypisan++
}

func (st *stan) usun(w int) {
	if !st.ok[w] {
		return
	}
	sala := st.przyp[w].sala
	st.wSali[sala] = slices.DeleteFunc(st.wSali[sala], func(v int) bool { return v == w })
	st.ok[w] = false
	st.przypisan--
}

// nakladaja: ten sam dzień, wspólny slot i wspólny tydzień (data/README.md, sekcja 5).
func (st *stan) nakladaja(a int, ka kandydat, b int, kb kandydat) bool {
	if ka.dzien != kb.dzien {
		return false
	}
	wa, wb := st.pr.wezly[a], st.pr.wezly[b]
	if ka.start > kb.start+wb.dlugosc-1 || kb.start > ka.start+wa.dlugosc-1 {
		return false
	}
	return wa.warianty[ka.wariant].suma.Nakladaja(wb.warianty[kb.wariant].suma)
}

// konflikty wywołuje f dla każdego przypisanego węzła, z którym kandydat k węzła w łamie regułę twardą:
// wspólny prowadzący lub student (3, 5), ta sama sala (4) albo ten sam dzień w zestawie rozne_dni (10).
// Węzeł może pojawić się kilka razy, raz na każdą złamaną regułę. f zwraca false, żeby przerwać.
func (st *stan) konflikty(w int, k kandydat, f func(v int) bool) {
	wz := st.pr.wezly[w]
	for _, v := range wz.sasiedzi {
		if st.ok[v] && st.nakladaja(w, k, v, st.przyp[v]) && !f(v) {
			return
		}
	}
	for _, v := range st.wSali[k.sala] {
		if v != w && st.nakladaja(w, k, v, st.przyp[v]) && !f(v) {
			return
		}
	}
	for _, v := range wz.rozneDni {
		if st.ok[v] && st.przyp[v].dzien == k.dzien && !f(v) {
			return
		}
	}
}

// wolny mówi, czy kandydat nie łamie żadnej reguły względem przypisanych węzłów.
func (st *stan) wolny(w int, k kandydat) bool {
	ok := true
	st.konflikty(w, k, func(int) bool { ok = false; return false })
	return ok
}

// liczbaKonfliktow to liczba złamanych reguł przez kandydata (do trybu awaryjnego), z naruszeniami stałymi węzła.
func (st *stan) liczbaKonfliktow(w int, k kandydat) int {
	n := st.pr.wezly[w].naruszeniaStale
	st.konflikty(w, k, func(int) bool { n++; return true })
	return n
}

// maWolnegoKandydata mówi, czy nieprzypisany węzeł ma choć jednego kandydata bez konfliktów.
func (st *stan) maWolnegoKandydata(w int) bool {
	jest := false
	st.pr.wezly[w].kandydaci(func(k kandydat) bool {
		jest = st.wolny(w, k)
		return !jest
	})
	return jest
}

// przypisania zamienia stan na przypisania plan.json (jedno na zajęcia, w kolejności z danych).
func (st *stan) przypisania() []model.Przypisanie {
	out := make([]model.Przypisanie, 0, len(st.pr.dane.Zajecia))
	for i := range st.pr.dane.Zajecia {
		z := &st.pr.dane.Zajecia[i]
		w := st.pr.wezly[st.pr.wezelZajec[z.ID]]
		k := st.przyp[w.id]
		var tyg model.Tygodnie
		for j, zw := range w.zajecia {
			if zw.ID == z.ID {
				tyg = w.warianty[k.wariant].tygodnie[j]
			}
		}
		out = append(out, model.Przypisanie{
			ZajeciaID: z.ID,
			Dzien:     k.dzien,
			SlotStart: k.start,
			SalaID:    st.pr.sale[k.sala].ID,
			Tygodnie:  tyg,
		})
	}
	return out
}
