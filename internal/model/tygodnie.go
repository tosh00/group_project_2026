package model

import (
	"encoding/json"
	"fmt"
	"math/bits"
)

// Tygodnie to zbiór tygodni 1–14 jako maska bitowa: bit 0 = tydzień 1.
// W JSON zapisujemy go jako rosnącą listę numerów (ADR 0003).
type Tygodnie uint16

// WszystkieTygodnie to tygodnie 1–14.
const WszystkieTygodnie Tygodnie = 1<<LiczbaTygodni - 1

// Tydzien zwraca zbiór z jednym tygodniem t (1–14). Poza zakresem zwraca zbiór pusty.
func Tydzien(t int) Tygodnie {
	if t < 1 || t > LiczbaTygodni {
		return 0
	}
	return 1 << (t - 1)
}

// Zakres zwraca tygodnie od–do włącznie, przycięte do 1–14.
func Zakres(od, do int) Tygodnie {
	var w Tygodnie
	for t := od; t <= do; t++ {
		w |= Tydzien(t)
	}
	return w
}

// TygodnieZListy buduje zbiór z listy numerów. Numery spoza 1–14 są pomijane.
func TygodnieZListy(lista []int) Tygodnie {
	var w Tygodnie
	for _, t := range lista {
		w |= Tydzien(t)
	}
	return w
}

// Parzyste i Nieparzyste ograniczają zbiór do tygodni o danej parzystości.
func (w Tygodnie) Parzyste() Tygodnie    { return w & 0b10101010101010 }
func (w Tygodnie) Nieparzyste() Tygodnie { return w & 0b01010101010101 }

func (w Tygodnie) Zawiera(t int) bool        { return w&Tydzien(t) != 0 }
func (w Tygodnie) Liczba() int               { return bits.OnesCount16(uint16(w)) }
func (w Tygodnie) Nakladaja(v Tygodnie) bool { return w&v != 0 }
func (w Tygodnie) Bez(v Tygodnie) Tygodnie   { return w &^ v }
func (w Tygodnie) Podzbior(v Tygodnie) bool  { return w&^v == 0 }

// Lista zwraca rosnącą listę numerów tygodni.
func (w Tygodnie) Lista() []int {
	lista := make([]int, 0, w.Liczba())
	for t := 1; t <= LiczbaTygodni; t++ {
		if w.Zawiera(t) {
			lista = append(lista, t)
		}
	}
	return lista
}

func (w Tygodnie) MarshalJSON() ([]byte, error) { return json.Marshal(w.Lista()) }

// UnmarshalJSON przyjmuje listę numerów 1–14 bez powtórzeń.
func (w *Tygodnie) UnmarshalJSON(b []byte) error {
	var lista []int
	if err := json.Unmarshal(b, &lista); err != nil {
		return fmt.Errorf("tygodnie: oczekiwano listy numerów: %w", err)
	}
	var wynik Tygodnie
	for _, t := range lista {
		if t < 1 || t > LiczbaTygodni {
			return fmt.Errorf("tygodnie: numer %d poza zakresem 1–%d", t, LiczbaTygodni)
		}
		if wynik.Zawiera(t) {
			return fmt.Errorf("tygodnie: powtórzony tydzień %d", t)
		}
		wynik |= Tydzien(t)
	}
	*w = wynik
	return nil
}
