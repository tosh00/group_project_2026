package solver

// Wagi kryteriów miękkich (docs/plan.md, „Funkcja kosztu”). Wartości startowe są w DomyslneWagi;
// ostateczne wartości to otwarta kwestia #4 (docs/open-questions.md), dlatego wagi są parametrem Solve.
type Wagi struct {
	OkienkaStudentow    float64 `json:"okienka_studentow"`    // wolne sloty między zajęciami grupy w dniu (× studenci/25)
	PrzeciazonyDzien    float64 `json:"przeciazony_dzien"`    // każdy slot studenta ponad 8 w dniu
	OkienkaProwadzacych float64 `json:"okienka_prowadzacych"` // wolne sloty między zajęciami prowadzącego
	Po17                float64 `json:"po_17"`                // sloty 10–12 × studenci/25
	Przed9              float64 `json:"przed_9"`              // slot 1 × studenci/25
	ZaDuzaSala          float64 `json:"za_duza_sala"`         // (pojemność − osoby) / pojemność
	PreferencjaDni      float64 `json:"preferencja_dni"`      // skumulowane / nieskumulowane: odchylenie liczby dni
	NiePo17             float64 `json:"nie_po_17"`            // inne_uwagi „nie po 17:00”: slot ≥ 10
	DzienWolny          float64 `json:"dzien_wolny"`          // inne_uwagi „jeden dzień wolny”: zajęcia w 5 dniach
	PrzerwaMax2h        float64 `json:"przerwa_max_2h"`       // inne_uwagi „przerwa ≤ 2 h”: sloty przerwy ponad 2
	PoraDnia            float64 `json:"pora_dnia"`            // inne_uwagi rano / popołudnie: sloty poza preferowaną połową dnia
	SaleA               float64 `json:"sale_a"`               // inne_uwagi „sale A”: zajęcia w sali z oznaczeniem B
	WykladyRocznika     float64 `json:"wyklady_rocznika"`     // wykłady rocznika w jednym dniu ponad limit
}

// DomyslneWagi to tabela startowa z docs/plan.md.
func DomyslneWagi() Wagi {
	return Wagi{
		OkienkaStudentow:    10,
		PrzeciazonyDzien:    15,
		OkienkaProwadzacych: 5,
		Po17:                4,
		Przed9:              3,
		ZaDuzaSala:          6,
		PreferencjaDni:      8,
		NiePo17:             20,
		DzienWolny:          10,
		PrzerwaMax2h:        8,
		PoraDnia:            3,
		SaleA:               3,
		WykladyRocznika:     2,
	}
}
