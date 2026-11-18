// Package dotenv wczytuje lokalne ustawienia z pliku .env do zmiennych środowiskowych.
//
// Format: jedna zmienna na linię, KLUCZ=wartość. Puste linie i linie od # są pomijane,
// przedrostek "export " jest dozwolony, a wartość w cudzysłowach "…" albo '…' traci cudzysłowy.
// Zmienna już ustawiona w środowisku wygrywa z plikiem (np. w Dockerze i CI).
//
// Wzór pliku: .env.example w katalogu głównym repo. Sam .env nie trafia do repo (.gitignore).
package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Wczytaj ustawia zmienne z pliku. Brak pliku nie jest błędem.
func Wczytaj(sciezka string) error {
	f, err := os.Open(sciezka)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for nr := 1; sc.Scan(); nr++ {
		linia := strings.TrimSpace(sc.Text())
		if linia == "" || strings.HasPrefix(linia, "#") {
			continue
		}
		linia = strings.TrimPrefix(linia, "export ")
		klucz, wartosc, ok := strings.Cut(linia, "=")
		klucz = strings.TrimSpace(klucz)
		if !ok || klucz == "" {
			return fmt.Errorf("%s:%d: oczekiwano KLUCZ=wartość", sciezka, nr)
		}
		wartosc = strings.TrimSpace(wartosc)
		if n := len(wartosc); n >= 2 && (wartosc[0] == '"' || wartosc[0] == '\'') && wartosc[n-1] == wartosc[0] {
			wartosc = wartosc[1 : n-1]
		}
		if _, jest := os.LookupEnv(klucz); jest {
			continue
		}
		if err := os.Setenv(klucz, wartosc); err != nil {
			return err
		}
	}
	return sc.Err()
}

// WczytajZRepo szuka pliku nazwa (np. ".env") w katalogu bieżącym i wyżej, aż do katalogu z go.mod,
// i wczytuje pierwszy znaleziony. Dzięki temu działa też w testach, które go test uruchamia
// w katalogu pakietu. Brak pliku nie jest błędem.
func WczytajZRepo(nazwa string) error {
	kat, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		sciezka := filepath.Join(kat, nazwa)
		if _, err := os.Stat(sciezka); err == nil {
			return Wczytaj(sciezka)
		}
		if _, err := os.Stat(filepath.Join(kat, "go.mod")); err == nil {
			return nil // katalog główny modułu, dalej nie szukamy
		}
		wyzej := filepath.Dir(kat)
		if wyzej == kat {
			return nil
		}
		kat = wyzej
	}
}
