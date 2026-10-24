package model

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

const sciezkaDanych = "../../data/dane_planu.json"

func wczytajDane(t *testing.T) *Dane {
	t.Helper()
	d, err := Wczytaj(sciezkaDanych)
	if err != nil {
		t.Fatalf("Wczytaj: %v", err)
	}
	return d
}

func TestWczytajPrawdziweDane(t *testing.T) {
	d := wczytajDane(t)

	liczby := []struct {
		tabela string
		got    int
		want   int
	}{
		{"zajecia", len(d.Zajecia), 174},
		{"sale", len(d.Sale), 55},
		{"prowadzacy", len(d.Prowadzacy), 50},
		{"roczniki", len(d.Roczniki), 6},
		{"grupy", len(d.Grupy), 51},
		{"studenci", len(d.Studenci), 372},
		{"kolizje_grup", len(d.KolizjeGrup), 173},
		{"dezyderaty_terminy", len(d.DezyderatyTerminy), 288},
	}
	for _, l := range liczby {
		if l.got != l.want {
			t.Errorf("%s: %d wierszy, oczekiwano %d", l.tabela, l.got, l.want)
		}
	}

	ix := NowyIndeks(d)
	z001 := ix.Zajecia["Z001"]
	if z001 == nil || z001.Czestotliwosc != CoTydzien || z001.SlotyNaSpotkanie != 2 || z001.Prowadzacy[0] != "P39" {
		t.Errorf("Z001 wczytane źle: %+v", z001)
	}
	if got := ix.Zajecia["Z169"].DozwoloneTygodnie(); got != WszystkieTygodnie.Bez(Tydzien(6)) {
		t.Errorf("Z169: dozwolone tygodnie %v, oczekiwano wszystkich bez 6", got.Lista())
	}
	if s := ix.Sale["A281"]; s.Typ != Aula || s.Pojemnosc != 264 {
		t.Errorf("A281: %+v", s)
	}
	if !ix.GrupyKoliduja("INF1-R1-ALL", "INF1-R1-G1") {
		t.Error("cały rocznik powinien kolidować ze swoją grupą")
	}
	if ix.GrupyKoliduja("INF1-R1-ALL", "INF1-R2-ALL") {
		t.Error("różne roczniki nie powinny kolidować")
	}
	if zs := ix.Powiazania["POW-11"]; len(zs) != 2 || zs[0].TypPowiazania != TenSamTermin {
		t.Errorf("POW-11: %v", zs)
	}
}

func TestDekodujOdrzucaNieznanePola(t *testing.T) {
	_, err := Dekoduj(strings.NewReader(`{"zajecia": [], "nowa_tabela": []}`))
	if err == nil || !strings.Contains(err.Error(), "nowa_tabela") {
		t.Errorf("oczekiwano błędu o nieznanym polu, jest: %v", err)
	}
}

func TestPlanJSON(t *testing.T) {
	plan := &Plan{
		Meta: MetaPlanu{Silnik: "test", Seed: 42, Koszt: 1.5, Utworzono: time.Date(2026, 10, 24, 14, 0, 0, 0, time.UTC)},
		Przypisania: []Przypisanie{
			{ZajeciaID: "Z001", Dzien: 1, SlotStart: 3, SalaID: "A281", Tygodnie: WszystkieTygodnie},
			{ZajeciaID: "Z056", Dzien: 2, SlotStart: 5, SalaID: "B153", Tygodnie: WszystkieTygodnie.Parzyste(), Zablokowane: true},
		},
	}
	var buf bytes.Buffer
	if err := ZapiszPlan(&buf, plan); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`"wersja_formatu": 1`, `"tygodnie": [`, `"zablokowane": true`} {
		if !strings.Contains(buf.String(), fragment) {
			t.Errorf("plan.json nie zawiera %s:\n%s", fragment, buf.String())
		}
	}
	if strings.Count(buf.String(), "zablokowane") != 1 {
		t.Errorf("zablokowane = false nie powinno być zapisywane:\n%s", buf.String())
	}

	wczytany, err := DekodujPlan(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if wczytany.Meta.WersjaFormatu != WersjaFormatuPlanu || wczytany.Przypisania[1] != plan.Przypisania[1] {
		t.Errorf("plan po zapisie i odczycie różni się: %+v", wczytany)
	}
}

func TestDekodujPlanSprawdzaWersje(t *testing.T) {
	_, err := DekodujPlan(strings.NewReader(`{"meta": {"wersja_formatu": 2}, "przypisania": []}`))
	if err == nil {
		t.Error("oczekiwano błędu dla wersji 2")
	}
}
