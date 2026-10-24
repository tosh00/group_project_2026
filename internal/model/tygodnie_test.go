package model

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestTygodnie(t *testing.T) {
	w := Zakres(6, 14)
	if got := w.Lista(); !slices.Equal(got, []int{6, 7, 8, 9, 10, 11, 12, 13, 14}) {
		t.Errorf("Zakres(6, 14).Lista() = %v", got)
	}
	if got := Zakres(1, 14); got != WszystkieTygodnie {
		t.Errorf("Zakres(1, 14) = %b, oczekiwano %b", got, WszystkieTygodnie)
	}
	if got := WszystkieTygodnie.Nieparzyste().Lista(); !slices.Equal(got, []int{1, 3, 5, 7, 9, 11, 13}) {
		t.Errorf("Nieparzyste() = %v", got)
	}
	if got := WszystkieTygodnie.Parzyste().Lista(); !slices.Equal(got, []int{2, 4, 6, 8, 10, 12, 14}) {
		t.Errorf("Parzyste() = %v", got)
	}
	if got := WszystkieTygodnie.Bez(Tydzien(6)).Liczba(); got != 13 {
		t.Errorf("bez tygodnia 6: Liczba() = %d, oczekiwano 13", got)
	}
	if Zakres(1, 7).Nakladaja(Zakres(8, 14)) {
		t.Error("tygodnie 1–7 i 8–14 nie powinny się nakładać")
	}
	if !Zakres(1, 7).Podzbior(WszystkieTygodnie) || WszystkieTygodnie.Podzbior(Zakres(1, 7)) {
		t.Error("Podzbior() zwraca złe wyniki")
	}
	if Tydzien(0) != 0 || Tydzien(15) != 0 {
		t.Error("Tydzien() poza zakresem powinien dawać zbiór pusty")
	}
}

func TestTygodnieJSON(t *testing.T) {
	b, err := json.Marshal(TygodnieZListy([]int{14, 2, 1}))
	if err != nil || string(b) != "[1,2,14]" {
		t.Fatalf("Marshal = %s, %v; oczekiwano [1,2,14]", b, err)
	}
	var w Tygodnie
	if err := json.Unmarshal([]byte("[3,5]"), &w); err != nil || !slices.Equal(w.Lista(), []int{3, 5}) {
		t.Fatalf("Unmarshal([3,5]) = %v, %v", w.Lista(), err)
	}
	for _, zle := range []string{"[0]", "[15]", "[2,2]", `"1"`, "5"} {
		if err := json.Unmarshal([]byte(zle), &w); err == nil {
			t.Errorf("Unmarshal(%s): oczekiwano błędu", zle)
		}
	}
}
