package solver

import (
	"slices"
	"testing"

	"github.com/tosh00/group_project_2026/internal/model"
)

func wczytajDane(t *testing.T) *model.Dane {
	t.Helper()
	d, err := model.Wczytaj("../../data/dane_planu.json")
	if err != nil {
		t.Fatalf("wczytanie danych: %v", err)
	}
	return d
}

func TestWariantyTygodni(t *testing.T) {
	przypadki := []struct {
		nazwa string
		z     model.Zajecia
		chce  []model.Tygodnie
	}{
		{
			"co tydzień cały semestr",
			model.Zajecia{Czestotliwosc: model.CoTydzien, LiczbaSpotkan: 14, TydzienOd: 1, TydzienDo: 14},
			[]model.Tygodnie{model.WszystkieTygodnie},
		},
		{
			"co tydzień połowa semestru",
			model.Zajecia{Czestotliwosc: model.CoTydzien, LiczbaSpotkan: 9, TydzienOd: 6, TydzienDo: 14},
			[]model.Tygodnie{model.Zakres(6, 14)},
		},
		{
			"co tydzień ze złą liczbą spotkań",
			model.Zajecia{Czestotliwosc: model.CoTydzien, LiczbaSpotkan: 13, TydzienOd: 1, TydzienDo: 14},
			nil,
		},
		{
			"co 2 tygodnie: parzyste i nieparzyste",
			model.Zajecia{Czestotliwosc: model.Co2Tygodnie, LiczbaSpotkan: 7, TydzienOd: 1, TydzienDo: 14},
			[]model.Tygodnie{model.WszystkieTygodnie.Nieparzyste(), model.WszystkieTygodnie.Parzyste()},
		},
		{
			"co 2 tygodnie: tylko jedna parzystość ma dość spotkań",
			model.Zajecia{Czestotliwosc: model.Co2Tygodnie, LiczbaSpotkan: 4, TydzienOd: 1, TydzienDo: 7},
			[]model.Tygodnie{model.Zakres(1, 7).Nieparzyste()},
		},
		{
			"wybrane tygodnie wyznaczone przez dane (jak Z169)",
			model.Zajecia{Czestotliwosc: model.WybraneTygodnie, LiczbaSpotkan: 13, TydzienOd: 1, TydzienDo: 14, TygodnieNiedostepne: []int{6}},
			[]model.Tygodnie{model.WszystkieTygodnie.Bez(model.Tydzien(6))},
		},
		{
			"wybrane tygodnie: okna kolejnych dozwolonych tygodni",
			model.Zajecia{Czestotliwosc: model.WybraneTygodnie, LiczbaSpotkan: 3, TydzienOd: 1, TydzienDo: 5, TygodnieNiedostepne: []int{2}},
			[]model.Tygodnie{
				model.TygodnieZListy([]int{1, 3, 4}),
				model.TygodnieZListy([]int{3, 4, 5}),
			},
		},
	}
	for _, p := range przypadki {
		t.Run(p.nazwa, func(t *testing.T) {
			got := wariantyTygodni(&p.z)
			if !slices.Equal(got, p.chce) {
				t.Errorf("warianty = %v, oczekiwano %v", listy(got), listy(p.chce))
			}
		})
	}
}

func listy(ws []model.Tygodnie) [][]int {
	var out [][]int
	for _, w := range ws {
		out = append(out, w.Lista())
	}
	return out
}

func TestPrzygotujPrawdziweDane(t *testing.T) {
	d := wczytajDane(t)
	pr := przygotuj(d)

	t.Run("każde zajęcia w dokładnie jednym węźle", func(t *testing.T) {
		ile := make(map[string]int)
		for _, w := range pr.wezly {
			for _, z := range w.zajecia {
				ile[z.ID]++
			}
		}
		for _, z := range d.Zajecia {
			if ile[z.ID] != 1 {
				t.Errorf("%s występuje w %d węzłach", z.ID, ile[z.ID])
			}
		}
	})

	t.Run("ten_sam_termin to jeden węzeł", func(t *testing.T) {
		w := pr.wezly[pr.wezelZajec["Z037"]]
		if pr.wezelZajec["Z038"] != w.id || len(w.zajecia) != 2 {
			t.Fatalf("Z037 i Z038 nie są w jednym węźle")
		}
		if got := w.warianty[0].suma; got != model.WszystkieTygodnie {
			t.Errorf("suma tygodni Z037+Z038 = %v, oczekiwano 1–14", got.Lista())
		}
	})

	t.Run("brak pustych domen", func(t *testing.T) {
		for _, w := range pr.wezly {
			if len(w.sale) == 0 || len(w.warianty) == 0 || w.liczbaStartow() == 0 {
				t.Errorf("%s: sale %d, warianty %d, starty %d", w.nazwa(), len(w.sale), len(w.warianty), w.liczbaStartow())
			}
		}
	})

	t.Run("sale jednoznaczne", func(t *testing.T) {
		przypadki := []struct {
			zajecia string
			chce    []string
		}{
			{"Z115", []string{"A234"}},
			{"Z071", []string{"B472"}},
			{"Z075", []string{"B472"}},
			{"Z136", []string{"B153"}},
		}
		for _, p := range przypadki {
			var got []string
			for _, s := range pr.wezly[pr.wezelZajec[p.zajecia]].sale {
				got = append(got, pr.sale[s].ID)
			}
			slices.Sort(got)
			if !slices.Equal(got, p.chce) {
				t.Errorf("%s: sale %v, oczekiwano %v", p.zajecia, got, p.chce)
			}
		}
	})

	t.Run("warianty tygodni", func(t *testing.T) {
		for _, w := range pr.wezly {
			for _, wr := range w.warianty {
				for i, z := range w.zajecia {
					tyg := wr.tygodnie[i]
					if !tyg.Podzbior(z.DozwoloneTygodnie()) || tyg.Liczba() != z.LiczbaSpotkan {
						t.Errorf("%s: zły wariant %v", z.ID, tyg.Lista())
					}
				}
			}
		}
		if n := len(pr.wezly[pr.wezelZajec["Z056"]].warianty); n != 2 {
			t.Errorf("Z056 (co_2_tygodnie): %d wariantów, oczekiwano 2", n)
		}
		if n := len(pr.wezly[pr.wezelZajec["Z105"]].warianty); n < 2 {
			t.Errorf("Z105 (9 z 14): %d wariantów, oczekiwano kilku", n)
		}
	})

	t.Run("sale spełniają reguły 6 i 7", func(t *testing.T) {
		for _, w := range pr.wezly {
			for _, si := range w.sale {
				s := pr.sale[si]
				for _, z := range w.zajecia {
					if s.Pojemnosc < z.LiczbaOsob() {
						t.Errorf("%s w %s: %d miejsc < %d", z.ID, s.ID, s.Pojemnosc, z.LiczbaOsob())
					}
					if len(z.DozwoloneSale) > 0 && !slices.Contains(z.DozwoloneSale, s.ID) {
						t.Errorf("%s w %s spoza dozwolone_sale", z.ID, s.ID)
					}
					if len(z.DozwoloneSale) == 0 && (!s.DostepnaDoPlanowania || !slices.Contains(z.TypySal, s.Typ)) {
						t.Errorf("%s w %s: zły typ lub sala niedostępna", z.ID, s.ID)
					}
					for _, wyp := range z.WymaganeWyposazenie {
						if !slices.Contains(s.Wyposazenie, wyp) {
							t.Errorf("%s w %s: brak %s", z.ID, s.ID, wyp)
						}
					}
				}
			}
		}
	})

	t.Run("starty omijają wykluczenia, Radę i blokady", func(t *testing.T) {
		ix := model.NowyIndeks(d)
		zablokowane := func(z *model.Zajecia, dzien, slot int) bool {
			for _, p := range z.Prowadzacy {
				for _, dt := range d.DezyderatyTerminy {
					if dt.Prowadzacy == p && dt.Dzien == dzien && dt.Slot == slot {
						return true
					}
				}
				if ix.Prowadzacy[p].CzlonekRadyWydzialu {
					for _, r := range d.RadaWydzialu {
						if r.Dzien == dzien && r.SlotOd <= slot && slot <= r.SlotDo {
							return true
						}
					}
				}
			}
			for _, b := range d.BlokadyStale {
				if b.Dzien != dzien || slot < b.SlotOd || slot > b.SlotDo {
					continue
				}
				for _, g := range z.Grupy {
					if ix.GrupyKoliduja(g, b.Grupa) {
						return true
					}
				}
			}
			return false
		}
		for _, w := range pr.wezly {
			for dzien := 1; dzien <= model.LiczbaDni; dzien++ {
				for start := 1; start <= model.LiczbaSlotow; start++ {
					if !w.moznaStart(dzien, start) {
						continue
					}
					if start+w.dlugosc-1 > model.LiczbaSlotow {
						t.Errorf("%s: start %d wychodzi poza dzień", w.nazwa(), start)
					}
					for _, z := range w.zajecia {
						for s := start; s < start+w.dlugosc; s++ {
							if zablokowane(z, dzien, s) {
								t.Errorf("%s: dozwolony start dzień %d slot %d, ale slot %d jest zablokowany", z.ID, dzien, start, s)
							}
						}
					}
				}
			}
		}
	})

	t.Run("graf konfliktów", func(t *testing.T) {
		ix := model.NowyIndeks(d)
		for _, a := range pr.wezly {
			for _, b := range pr.wezly {
				if a.id == b.id {
					continue
				}
				chce := false
				for _, za := range a.zajecia {
					for _, zb := range b.zajecia {
						chce = chce || model.WspolnyProwadzacy(za, zb) || ix.ZajeciaKolidujaStudenci(za, zb)
					}
				}
				if got := slices.Contains(a.sasiedzi, b.id); got != chce {
					t.Errorf("%s–%s: sąsiedzi %v, oczekiwano %v", a.nazwa(), b.nazwa(), got, chce)
				}
			}
		}
	})

	t.Run("rozne_dni", func(t *testing.T) {
		for nazwa, zs := range model.NowyIndeks(d).Powiazania {
			if zs[0].TypPowiazania != model.RozneDni {
				continue
			}
			for _, z1 := range zs {
				w1 := pr.wezly[pr.wezelZajec[z1.ID]]
				for _, z2 := range zs {
					if z1 != z2 && !slices.Contains(w1.rozneDni, pr.wezelZajec[z2.ID]) {
						t.Errorf("%s: brak %s w rozneDni %s", nazwa, z2.ID, z1.ID)
					}
				}
			}
		}
	})
}
