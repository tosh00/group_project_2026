package api

import (
	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/validator"
)

// Walidator ocenia gotowy plan pod kątem reguł twardych 1–10 (ADR 0006: ma ostatnie słowo).
// Serwer używa PelnyWalidator; Atrapa zostaje do testów, które nie sprawdzają reguł.
type Walidator interface {
	// Nazwa trafia do raportu i /api/zdrowie: "atrapa" albo "internal/validator".
	Nazwa() string
	// Sprawdz zwraca wszystkie naruszenia (pusta lista = plan poprawny). Musi być bezpieczna współbieżnie.
	Sprawdz(dane *model.Dane, plan *model.Plan) []Naruszenie
}

// PelnyWalidator to adapter na internal/validator.
type PelnyWalidator struct{}

func (PelnyWalidator) Nazwa() string { return "internal/validator" }

func (PelnyWalidator) Sprawdz(dane *model.Dane, plan *model.Plan) []Naruszenie {
	ns := validator.Nowy(dane).Sprawdz(plan)
	wynik := make([]Naruszenie, len(ns))
	for i, n := range ns {
		wynik[i] = Naruszenie{
			Regula:    n.Regula,
			Nazwa:     n.Nazwa,
			ZajeciaID: n.Zajecia,
			Dzien:     n.Dzien,
			Sloty:     n.Sloty,
			SalaID:    n.SalaID,
			Tygodnie:  n.Tygodnie,
			Opis:      n.Opis,
		}
	}
	return wynik
}

// Atrapa udaje walidator i nie zgłasza żadnych naruszeń. Raport z atrapy nie jest dowodem poprawności planu.
type Atrapa struct{}

func (Atrapa) Nazwa() string                                 { return "atrapa" }
func (Atrapa) Sprawdz(*model.Dane, *model.Plan) []Naruszenie { return nil }

// Raport to raport walidatora (szkic z docs/contracts/plan-format.md, schemat Raport w api/openapi.yaml).
type Raport struct {
	Status         string             `json:"status"` // poprawny | z_naruszeniami
	LiczbaNaruszen int                `json:"liczba_naruszen"`
	Naruszenia     []Naruszenie       `json:"naruszenia"`
	KosztMiekki    float64            `json:"koszt_miekki"`
	KosztKryteria  map[string]float64 `json:"koszt_kryteria"`
	Walidator      string             `json:"walidator"`
}

// Naruszenie to jedno złamanie reguły twardej (format z docs/contracts/plan-format.md).
// Pola położenia są puste, gdy nie dotyczą naruszenia (np. zajęcia bez przypisania).
type Naruszenie struct {
	Regula    int      `json:"regula"`
	Nazwa     string   `json:"nazwa"`
	ZajeciaID []string `json:"zajecia_id"`
	Dzien     int      `json:"dzien,omitempty"`
	Sloty     []int    `json:"sloty,omitempty"`
	SalaID    string   `json:"sala_id,omitempty"`
	Tygodnie  []int    `json:"tygodnie,omitempty"`
	Opis      string   `json:"opis"`
}

// NazwaReguly zwraca identyfikator reguły 1–10 (pusty dla nieznanego numeru).
func NazwaReguly(regula int) string { return validator.NazwaReguly(regula) }
