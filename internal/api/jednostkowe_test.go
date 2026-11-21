package api

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/solver"
)

func TestOdczytajUstawienia(t *testing.T) {
	domyslne := solver.DomyslneWagi()
	testy := []struct {
		nazwa   string
		tresc   string
		blad    string
		sprawdz func(t *testing.T, w solver.Wagi, seed int64, iteracje int)
	}{
		{nazwa: "puste", tresc: "", sprawdz: func(t *testing.T, w solver.Wagi, seed int64, it int) {
			if w != domyslne || it != solver.DomyslneIteracje || seed < 0 || seed >= maksDomyslnySeed {
				t.Errorf("wagi %+v, seed %d, iteracje %d", w, seed, it)
			}
		}},
		{nazwa: "pusty obiekt", tresc: "{}", sprawdz: func(t *testing.T, w solver.Wagi, _ int64, it int) {
			if w != domyslne || it != solver.DomyslneIteracje {
				t.Errorf("wagi %+v, iteracje %d", w, it)
			}
		}},
		{nazwa: "część wag", tresc: `{"wagi":{"po_17":40},"seed":7,"iteracje":0}`,
			sprawdz: func(t *testing.T, w solver.Wagi, seed int64, it int) {
				chce := domyslne
				chce.Po17 = 40
				if w != chce || seed != 7 || it != 0 {
					t.Errorf("wagi %+v, seed %d, iteracje %d", w, seed, it)
				}
			}},
		{nazwa: "wagi null", tresc: `{"wagi":null}`, sprawdz: func(t *testing.T, w solver.Wagi, _ int64, _ int) {
			if w != domyslne {
				t.Errorf("wagi %+v", w)
			}
		}},
		{nazwa: "nieznana waga", tresc: `{"wagi":{"po_18":1}}`, blad: "niepoprawne wagi"},
		{nazwa: "nieznane pole", tresc: `{"wagi":{},"x":1}`, blad: "niepoprawne ustawienia"},
		{nazwa: "ujemna waga", tresc: `{"wagi":{"sale_a":-1}}`, blad: "sale_a nie może być ujemna"},
		{nazwa: "ujemne iteracje", tresc: `{"iteracje":-1}`, blad: "iteracje muszą być"},
		{nazwa: "za dużo iteracji", tresc: fmt.Sprintf(`{"iteracje":%d}`, MaksIteracji+1), blad: "iteracje muszą być"},
		{nazwa: "zły typ", tresc: `{"seed":"x"}`, blad: "niepoprawne ustawienia"},
	}
	for _, tt := range testy {
		t.Run(tt.nazwa, func(t *testing.T) {
			w, seed, it, err := odczytajUstawienia([]byte(tt.tresc))
			if tt.blad != "" {
				if err == nil || !strings.Contains(err.Error(), tt.blad) {
					t.Fatalf("err = %v, chcę %q", err, tt.blad)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tt.sprawdz(t, w, seed, it)
		})
	}
}

func TestSzczegolyBledu(t *testing.T) {
	polaczony := fmt.Errorf("niespójne dane: %w", errors.Join(errors.New("a"), errors.New("b")))
	if got := szczegolyBledu(polaczony); strings.Join(got, "|") != "a|b" {
		t.Errorf("połączony: %v", got)
	}
	if got := szczegolyBledu(fmt.Errorf("dekodowanie: %w", errors.New("x"))); len(got) != 1 || got[0] != "dekodowanie: x" {
		t.Errorf("pojedynczy: %v", got)
	}
}

func TestPodsumujPrawdziweDane(t *testing.T) {
	d, err := model.Wczytaj("../../data/dane_planu.json")
	if err != nil {
		t.Fatal(err)
	}
	p, o := podsumuj(d)
	chce := Podsumowanie{Zajecia: 174, Sale: 55, SaleDoPlanowania: 30, Prowadzacy: 50, Roczniki: 6, Grupy: 45, Studenci: 372}
	if p != chce {
		t.Errorf("podsumowanie %+v, chcę %+v", p, chce)
	}
	// docs/notes.md: szacowaną pojemność mają tylko B151, B153, B472.
	var sale []string
	for _, x := range o {
		sale = append(sale, x.SalaID)
	}
	if strings.Join(sale, ",") != "B151,B153,B472" {
		t.Errorf("ostrzeżenia dla sal %v", sale)
	}
}

func TestAgregatorWybieraNajlepszegoWorkera(t *testing.T) {
	a := &agregator{ostatnie: make([]solver.Postep, 3), iteracji: 300}
	a.ostatnie[0] = solver.Postep{Pracownik: 0, Iteracja: 50, Iteracji: 100, Koszt: 100, Naruszenia: 1}
	a.ostatnie[1] = solver.Postep{Pracownik: 1, Iteracja: 40, Iteracji: 100, Koszt: 500, Naruszenia: 0}
	a.ostatnie[2] = solver.Postep{Pracownik: 2, Iteracja: 30, Iteracji: 100, Koszt: 400, Naruszenia: 0}
	p := a.postep(a.start)
	if p.Iteracja != 120 || p.Iteracji != 300 || p.Koszt != 400 || p.Naruszenia != 0 {
		t.Errorf("postęp %+v", p)
	}
}

func TestNazwyRegul(t *testing.T) {
	for r := 1; r <= 10; r++ {
		if NazwaReguly(r) == "" {
			t.Errorf("reguła %d bez nazwy", r)
		}
	}
}
