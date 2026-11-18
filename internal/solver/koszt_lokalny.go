package solver

import (
	"strings"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Kryteria miękkie zależne tylko od przypisania jednego węzła (bez innych zajęć).
// Konstrukcja zachłanna wybiera według nich kandydatów; faza 3 dolicza kryteria globalne (okienka, dni).

const (
	slotPo17     = 10 // sloty 10–12 zaczynają się po 17:00
	slotPrzed9   = 1  // slot 1 = 8:15–9:00
	ostatniRanny = 6  // sloty 1–6 to „rano”, 7–12 „popołudnie” (pora dnia z inne_uwagi)
)

// uwagiProwadzacego to preferencje z dezyderaty_ogolne.inne_uwagi rozpoznane po słowach kluczowych.
type uwagiProwadzacego struct {
	niePo17      bool // „Proszę nie planować zajęć po 17:00.”
	rano         bool // „Wolę zajęcia w godzinach porannych.”
	popoludnie   bool // „Wolę zajęcia w godzinach popołudniowych.”
	saleA        bool // „Proszę o zajęcia w budynku A, jeśli to możliwe.”
	dzienWolny   bool // „Proszę o jeden dzień w tygodniu bez zajęć.”
	przerwaMax2h bool // „Proszę o przerwę między zajęciami nie dłuższą niż 2 godziny.”
}

func rozpoznajUwagi(tekst string) uwagiProwadzacego {
	t := strings.ToLower(tekst)
	return uwagiProwadzacego{
		niePo17:      strings.Contains(t, "po 17"),
		rano:         strings.Contains(t, "porann"),
		popoludnie:   strings.Contains(t, "popołudn"),
		saleA:        strings.Contains(t, "budynku a"),
		dzienWolny:   strings.Contains(t, "dzień w tygodniu bez"),
		przerwaMax2h: strings.Contains(t, "przerw"),
	}
}

func (pr *problem) uwagi(prowadzacy string) uwagiProwadzacego {
	if o := pr.ix.DezyderatyOgolne[prowadzacy]; o != nil {
		return rozpoznajUwagi(o.InneUwagi)
	}
	return uwagiProwadzacego{}
}

// slotyWZakresie liczy sloty spotkania [start, start+dl−1] w przedziale [od, do].
func slotyWZakresie(start, dl, od, do int) int {
	return max(0, min(start+dl-1, do)-max(start, od)+1)
}

// kosztLokalny to suma kryteriów miękkich zależnych tylko od kandydata k węzła w.
func (pr *problem) kosztLokalny(w *wezel, k kandydat, wagi *Wagi) float64 {
	sala := pr.sale[k.sala]
	var koszt float64
	for _, z := range w.zajecia {
		studenci := float64(z.LiczbaStudentow) / 25
		koszt += wagi.Po17 * float64(slotyWZakresie(k.start, z.SlotyNaSpotkanie, slotPo17, model.LiczbaSlotow)) * studenci
		koszt += wagi.Przed9 * float64(slotyWZakresie(k.start, z.SlotyNaSpotkanie, 1, slotPrzed9)) * studenci
		if sala.Pojemnosc > 0 {
			koszt += wagi.ZaDuzaSala * float64(max(0, sala.Pojemnosc-z.LiczbaOsob())) / float64(sala.Pojemnosc)
		}
		for _, p := range z.Prowadzacy {
			u := pr.uwagi(p)
			if u.niePo17 {
				koszt += wagi.NiePo17 * float64(slotyWZakresie(k.start, z.SlotyNaSpotkanie, slotPo17, model.LiczbaSlotow))
			}
			if u.rano {
				koszt += wagi.PoraDnia * float64(slotyWZakresie(k.start, z.SlotyNaSpotkanie, ostatniRanny+1, model.LiczbaSlotow))
			}
			if u.popoludnie {
				koszt += wagi.PoraDnia * float64(slotyWZakresie(k.start, z.SlotyNaSpotkanie, 1, ostatniRanny))
			}
			if u.saleA && strings.HasPrefix(sala.ID, "B") {
				koszt += wagi.SaleA
			}
		}
	}
	return koszt
}
