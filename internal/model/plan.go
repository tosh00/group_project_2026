package model

import "time"

// WersjaFormatuPlanu to obsługiwana wersja plan.json (ADR 0003, docs/contracts/plan-format.md).
const WersjaFormatuPlanu = 1

// Plan to zawartość plan.json: wynik solvera i wejście walidatora.
type Plan struct {
	Meta        MetaPlanu     `json:"meta"`
	Przypisania []Przypisanie `json:"przypisania"`
}

type MetaPlanu struct {
	WersjaFormatu int       `json:"wersja_formatu"`
	Silnik        string    `json:"silnik"`
	Seed          int64     `json:"seed"`
	Koszt         float64   `json:"koszt"`
	Utworzono     time.Time `json:"utworzono"`
}

// Przypisanie to termin i sala jednych zajęć. Spotkanie zajmuje sloty
// SlotStart … SlotStart + SlotyNaSpotkanie − 1 w każdym tygodniu z Tygodnie.
type Przypisanie struct {
	ZajeciaID   string   `json:"zajecia_id"`
	Dzien       int      `json:"dzien"`      // 1–5
	SlotStart   int      `json:"slot_start"` // 1–12
	SalaID      string   `json:"sala_id"`
	Tygodnie    Tygodnie `json:"tygodnie"`
	Zablokowane bool     `json:"zablokowane,omitempty"` // przypięte ręcznie, solver nie przesuwa
}
