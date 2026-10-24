package model

import (
	"strings"
	"testing"
)

func TestSprawdzPrawdziweDane(t *testing.T) {
	if err := wczytajDane(t).Sprawdz(); err != nil {
		t.Fatalf("prawdziwe dane powinny być spójne: %v", err)
	}
}

// Każdy przypadek psuje prawdziwe dane w jeden sposób i oczekuje błędu z danym fragmentem.
func TestSprawdzWykrywaNiespojnosci(t *testing.T) {
	przypadki := []struct {
		nazwa    string
		zepsuj   func(d *Dane)
		fragment string
	}{
		{"powtórzone id zajęć", func(d *Dane) { d.Zajecia[1].ID = d.Zajecia[0].ID }, "powtórzony identyfikator"},
		{"nieznany prowadzący", func(d *Dane) { d.Zajecia[0].Prowadzacy = []string{"P99"} }, `"P99"`},
		{"nieznana grupa", func(d *Dane) { d.Zajecia[0].Grupy = []string{"XYZ"} }, `"XYZ"`},
		{"nieznana sala", func(d *Dane) { d.Zajecia[0].DozwoloneSale = []string{"C999"} }, `"C999"`},
		{"zła forma", func(d *Dane) { d.Zajecia[0].Forma = "X" }, "forma"},
		{"zła częstotliwość", func(d *Dane) { d.Zajecia[0].Czestotliwosc = "co_miesiac" }, "czestotliwosc"},
		{"zakres tygodni odwrócony", func(d *Dane) { d.Zajecia[0].TydzienOd, d.Zajecia[0].TydzienDo = 10, 5 }, "tydzien_do"},
		{"za długie spotkanie", func(d *Dane) { d.Zajecia[0].SlotyNaSpotkanie = 13 }, "sloty_na_spotkanie"},
		{"liczba spotkań co tydzień", func(d *Dane) { d.Zajecia[0].LiczbaSpotkan = 13 }, "co_tydzien"},
		{"brak sal", func(d *Dane) { d.Zajecia[0].TypySal = nil }, "puste typy_sal"},
		{"powiązanie bez typu", func(d *Dane) { d.Zajecia[0].Powiazanie = "POW-X" }, "typ_powiazania"},
		{"sala do planowania bez pojemności", func(d *Dane) { d.Sale[0].Pojemnosc = 0 }, "bez pojemności"},
		{"brakująca kolizja", func(d *Dane) { d.KolizjeGrup = d.KolizjeGrup[1:] }, "brak pary"},
		{"zła liczba wspólnych", func(d *Dane) { d.KolizjeGrup[0].WspolniStudenci++ }, "wspólnych studentów"},
		{"liczebność grupy", func(d *Dane) { d.Grupy[0].LiczbaStudentow++ }, "liczba_studentow"},
		{"dziura w kalendarzu", func(d *Dane) { d.Kalendarz = d.Kalendarz[2:] }, "kalendarz"},
		{"dezyderat poza siatką", func(d *Dane) { d.DezyderatyTerminy[0].Slot = 13 }, "slot"},
		{"blokada poza siatką", func(d *Dane) { d.BlokadyStale[0].Dzien = 6 }, "dzien"},
	}
	for _, p := range przypadki {
		t.Run(p.nazwa, func(t *testing.T) {
			d := wczytajDane(t)
			p.zepsuj(d)
			err := d.Sprawdz()
			if err == nil {
				t.Fatal("oczekiwano błędu")
			}
			if !strings.Contains(err.Error(), p.fragment) {
				t.Errorf("błąd nie zawiera %q:\n%v", p.fragment, err)
			}
		})
	}
}

func TestSprawdzTenSamTermin(t *testing.T) {
	d := wczytajDane(t)
	ix := NowyIndeks(d)
	zs := ix.Powiazania["POW-11"]
	zs[1].TydzienOd = 7 // Z038 zaczyna się w tygodniu, w którym jeszcze trwa Z037
	zs[1].LiczbaSpotkan = 8
	if err := d.Sprawdz(); err == nil || !strings.Contains(err.Error(), "nakładają się") {
		t.Errorf("oczekiwano błędu o nakładających się tygodniach, jest: %v", err)
	}
}
