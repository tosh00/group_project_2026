package api

import "github.com/tosh00/group_project_2026/internal/model"

// Walidator ocenia gotowy plan pod kątem reguł twardych 1–10 (ADR 0006: ma ostatnie słowo).
//
// Docelowo to adapter na internal/validator:
//
//	func (pelnyWalidator) Sprawdz(d *model.Dane, p *model.Plan) []Naruszenie {
//		for _, n := range validator.Nowy(d).Sprawdz(p) { ... Naruszenie{Regula: n.Regula, ZajeciaID: n.Zajecia, Opis: n.Opis} }
//	}
//
// Do czasu merge'a gałęzi validator/reguly-twarde używamy Atrapa (docs/contracts/api.md).
type Walidator interface {
	// Nazwa trafia do raportu i /api/zdrowie: "atrapa" albo "internal/validator".
	Nazwa() string
	// Sprawdz zwraca wszystkie naruszenia (pusta lista = plan poprawny). Musi być bezpieczna współbieżnie.
	Sprawdz(dane *model.Dane, plan *model.Plan) []Naruszenie
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

// Naruszenie to jedno złamanie reguły twardej. Pola położenia są opcjonalne,
// bo internal/validator (gałąź validator/reguly-twarde) ich jeszcze nie zwraca.
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

// nazwyRegul to stałe identyfikatory reguł twardych z data/README.md, sekcja 5.
var nazwyRegul = map[int]string{
	1:  "przypisanie",
	2:  "tygodnie",
	3:  "kolizja_prowadzacego",
	4:  "kolizja_sali",
	5:  "kolizja_grup",
	6:  "typ_sali",
	7:  "pojemnosc",
	8:  "dezyderaty",
	9:  "blokady_stale",
	10: "powiazania",
}

// NazwaReguly zwraca identyfikator reguły 1–10 (pusty dla nieznanego numeru).
func NazwaReguly(regula int) string { return nazwyRegul[regula] }
