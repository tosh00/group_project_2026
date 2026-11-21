package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/store"
)

// Fazy trwającego uruchomienia (schemat Faza).
const (
	FazaKonstrukcja   = "konstrukcja"
	FazaOptymalizacja = "optymalizacja"
	FazaWalidacja     = "walidacja"
)

// Zdrowie: schemat Zdrowie (GET /zdrowie, także przy 503).
type Zdrowie struct {
	Status    string `json:"status"`
	Baza      string `json:"baza"`
	Walidator string `json:"walidator"`
}

// Zestaw: schemat Zestaw.
type Zestaw struct {
	ID           uuid.UUID       `json:"id"`
	Nazwa        string          `json:"nazwa"`
	Utworzono    time.Time       `json:"utworzono"`
	Podsumowanie json.RawMessage `json:"podsumowanie"`
	Ostrzezenia  json.RawMessage `json:"ostrzezenia"`
}

func zestawDTO(z store.Zestaw) Zestaw {
	return Zestaw{ID: z.ID, Nazwa: z.Nazwa, Utworzono: z.Utworzono, Podsumowanie: z.Podsumowanie, Ostrzezenia: z.Ostrzezenia}
}

// Podsumowanie: schemat Podsumowanie.
type Podsumowanie struct {
	Zajecia          int `json:"zajecia"`
	Sale             int `json:"sale"`
	SaleDoPlanowania int `json:"sale_do_planowania"`
	Prowadzacy       int `json:"prowadzacy"`
	Roczniki         int `json:"roczniki"`
	Grupy            int `json:"grupy"`
	Studenci         int `json:"studenci"`
}

// Ostrzezenie: schemat Ostrzezenie.
type Ostrzezenie struct {
	Kod    string `json:"kod"`
	Opis   string `json:"opis"`
	SalaID string `json:"sala_id,omitempty"`
}

func podsumuj(d *model.Dane) (Podsumowanie, []Ostrzezenie) {
	p := Podsumowanie{
		Zajecia: len(d.Zajecia), Sale: len(d.Sale), Prowadzacy: len(d.Prowadzacy),
		Roczniki: len(d.Roczniki), Grupy: len(d.Grupy), Studenci: len(d.Studenci),
	}
	ostrzezenia := []Ostrzezenie{}
	for _, s := range d.Sale {
		if s.DostepnaDoPlanowania {
			p.SaleDoPlanowania++
		}
		if s.PojemnoscSzacowana {
			ostrzezenia = append(ostrzezenia, Ostrzezenie{
				Kod:    "pojemnosc_szacowana",
				Opis:   fmt.Sprintf("Sala %s: pojemność szacowana (%d miejsc)", s.ID, s.Pojemnosc),
				SalaID: s.ID,
			})
		}
	}
	return p, ostrzezenia
}

// Uruchomienie: schemat Uruchomienie.
type Uruchomienie struct {
	ID             uuid.UUID       `json:"id"`
	ZestawID       uuid.UUID       `json:"zestaw_id"`
	Status         string          `json:"status"`
	Faza           string          `json:"faza,omitempty"`
	Wagi           json.RawMessage `json:"wagi"`
	Seed           int64           `json:"seed"`
	Iteracje       int32           `json:"iteracje"`
	Zatrzymane     bool            `json:"zatrzymane"`
	LiczbaNaruszen *int32          `json:"liczba_naruszen"`
	Koszt          *float64        `json:"koszt"`
	Blad           *string         `json:"blad"`
	Utworzono      time.Time       `json:"utworzono"`
	Rozpoczeto     *time.Time      `json:"rozpoczeto"`
	Zakonczono     *time.Time      `json:"zakonczono"`
}

func uruchomienieDTO(u store.Uruchomienie, faza string) Uruchomienie {
	if u.Status != store.StatusTrwa {
		faza = ""
	}
	return Uruchomienie{
		ID: u.ID, ZestawID: u.ZestawID, Status: u.Status, Faza: faza, Wagi: u.Wagi, Seed: u.Seed,
		Iteracje: u.Iteracje, Zatrzymane: u.Zatrzymane, LiczbaNaruszen: u.LiczbaNaruszen, Koszt: u.Koszt,
		Blad: u.Blad, Utworzono: u.Utworzono, Rozpoczeto: u.Rozpoczeto, Zakonczono: u.Zakonczono,
	}
}

// Postep: schemat Postep (zdarzenie SSE "postep").
type Postep struct {
	Faza       string  `json:"faza"`
	Iteracja   int     `json:"iteracja"`
	Iteracji   int     `json:"iteracji"`
	Koszt      float64 `json:"koszt"`
	Naruszenia int     `json:"naruszenia"`
	CzasMs     int64   `json:"czas_ms"`
}

// NoweUruchomienie: schemat NoweUruchomienie.
type NoweUruchomienie struct {
	Wagi     json.RawMessage `json:"wagi"`
	Seed     *int64          `json:"seed"`
	Iteracje *int            `json:"iteracje"`
}

// MaksIteracji ogranicza iteracje na workera w jednym uruchomieniu (ok. 30 s przy 4 workerach).
const MaksIteracji = 20_000_000

// Blad: schemat Blad.
type Blad struct {
	Kod       string   `json:"kod"`
	Komunikat string   `json:"komunikat"`
	Szczegoly []string `json:"szczegoly,omitempty"`
}

func zapiszJSON(w http.ResponseWriter, kod int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(kod)
	_ = json.NewEncoder(w).Encode(v)
}

func zapiszSurowy(w http.ResponseWriter, kod int, dane json.RawMessage) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(kod)
	_, _ = w.Write(dane)
}

func zapiszBlad(w http.ResponseWriter, kod int, b Blad) { zapiszJSON(w, kod, b) }

// szczegolyBledu rozwija błąd z errors.Join (np. z model.Sprawdz) na listę komunikatów.
func szczegolyBledu(err error) []string {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if j, ok := e.(interface{ Unwrap() []error }); ok {
			var wynik []string
			for _, w := range j.Unwrap() {
				wynik = append(wynik, w.Error())
			}
			return wynik
		}
	}
	return []string{err.Error()}
}
