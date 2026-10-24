package model

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Wczytaj czyta dane_planu.json z pliku i sprawdza ich spójność (Sprawdz).
func Wczytaj(sciezka string) (*Dane, error) {
	f, err := os.Open(sciezka)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Dekoduj(f)
}

// Dekoduj czyta dane w formacie dane_planu.json i sprawdza ich spójność.
// Nieznane pola są błędem, żeby zmiana formatu danych nie przeszła po cichu.
func Dekoduj(r io.Reader) (*Dane, error) {
	var d Dane
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("dekodowanie danych: %w", err)
	}
	if err := d.Sprawdz(); err != nil {
		return nil, fmt.Errorf("niespójne dane: %w", err)
	}
	return &d, nil
}

// WczytajPlan czyta plan.json z pliku.
func WczytajPlan(sciezka string) (*Plan, error) {
	f, err := os.Open(sciezka)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DekodujPlan(f)
}

// DekodujPlan czyta plan w formacie plan.json i sprawdza wersję formatu.
// Nie sprawdza reguł planu: to robi internal/validator.
func DekodujPlan(r io.Reader) (*Plan, error) {
	var p Plan
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("dekodowanie planu: %w", err)
	}
	if p.Meta.WersjaFormatu != WersjaFormatuPlanu {
		return nil, fmt.Errorf("nieobsługiwana wersja formatu planu %d (obsługiwana: %d)",
			p.Meta.WersjaFormatu, WersjaFormatuPlanu)
	}
	return &p, nil
}

// ZapiszPlan zapisuje plan w formacie plan.json. Pusta WersjaFormatu jest uzupełniana.
func ZapiszPlan(w io.Writer, p *Plan) error {
	kopia := *p
	if kopia.Meta.WersjaFormatu == 0 {
		kopia.Meta.WersjaFormatu = WersjaFormatuPlanu
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(&kopia)
}
