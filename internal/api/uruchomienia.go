package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/tosh00/group_project_2026/internal/solver"
	"github.com/tosh00/group_project_2026/internal/store"
)

// maksDomyslnySeed: losowy seed mieści się w 2^53, żeby JavaScript (number) odczytał go dokładnie.
const maksDomyslnySeed = 1 << 53

func zakonczone(status string) bool {
	return status != store.StatusWKolejce && status != store.StatusTrwa
}

func (s *Serwer) uruchomienieDTO(u store.Uruchomienie) Uruchomienie {
	return uruchomienieDTO(u, s.pula.Faza(u.ID))
}

// odczytajUstawienia czyta NoweUruchomienie; brakujące wagi przyjmują wartości domyślne.
func odczytajUstawienia(tresc []byte) (wagi solver.Wagi, seed int64, iteracje int, err error) {
	wagi = solver.DomyslneWagi()
	seed = rand.Int64N(maksDomyslnySeed)
	iteracje = solver.DomyslneIteracje
	if len(bytes.TrimSpace(tresc)) == 0 {
		return wagi, seed, iteracje, nil
	}
	var nowe NoweUruchomienie
	dec := json.NewDecoder(bytes.NewReader(tresc))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&nowe); err != nil {
		return wagi, seed, iteracje, fmt.Errorf("niepoprawne ustawienia uruchomienia: %w", err)
	}
	if len(nowe.Wagi) > 0 && string(nowe.Wagi) != "null" {
		dec := json.NewDecoder(bytes.NewReader(nowe.Wagi))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&wagi); err != nil {
			return wagi, seed, iteracje, fmt.Errorf("niepoprawne wagi: %w", err)
		}
		var mapa map[string]float64
		b, _ := json.Marshal(wagi)
		_ = json.Unmarshal(b, &mapa)
		for k, v := range mapa {
			if v < 0 {
				return wagi, seed, iteracje, fmt.Errorf("waga %s nie może być ujemna", k)
			}
		}
	}
	if nowe.Seed != nil {
		seed = *nowe.Seed
	}
	if nowe.Iteracje != nil {
		iteracje = *nowe.Iteracje
		if iteracje < 0 || iteracje > MaksIteracji {
			return wagi, seed, iteracje, fmt.Errorf("iteracje muszą być w zakresie 0–%d", MaksIteracji)
		}
	}
	return wagi, seed, iteracje, nil
}

func (s *Serwer) uruchomSolver(w http.ResponseWriter, r *http.Request) {
	zestawID, ok := idZeSciezki(w, r, "zestawId")
	if !ok {
		return
	}
	tresc, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaksRozmiarUruchomienia))
	if err != nil {
		zapiszBlad(w, http.StatusBadRequest, Blad{Kod: "zle_zadanie", Komunikat: "Nie udało się odczytać ustawień uruchomienia."})
		return
	}
	wagi, seed, iteracje, err := odczytajUstawienia(tresc)
	if err != nil {
		zapiszBlad(w, http.StatusBadRequest, Blad{Kod: "zle_zadanie", Komunikat: "Niepoprawne ustawienia uruchomienia.", Szczegoly: []string{err.Error()}})
		return
	}
	wagiJSON, _ := json.Marshal(wagi)

	u, err := s.repo.DodajUruchomienie(r.Context(), zestawID, wagiJSON, seed, int32(iteracje))
	if errors.Is(err, store.ErrNieZnaleziono) {
		zapiszBlad(w, http.StatusNotFound, Blad{Kod: "nie_znaleziono", Komunikat: "Nie ma takiego zestawu danych."})
		return
	}
	if err != nil {
		s.bladSerwera(w, r, err)
		return
	}
	if err := s.pula.Zglos(u); err != nil {
		opis := "kolejka uruchomień była pełna"
		if _, err := s.repo.Zakoncz(r.Context(), store.Zakonczenie{ID: u.ID, Status: store.StatusBlad, Blad: &opis}); err != nil {
			s.log.Error("zapis odrzuconego uruchomienia", "uruchomienie", u.ID, "blad", err)
		}
		zapiszBlad(w, http.StatusServiceUnavailable, Blad{Kod: "kolejka_pelna", Komunikat: "Za dużo uruchomień naraz. Spróbuj za chwilę."})
		return
	}
	w.Header().Set("Location", "/api/uruchomienia/"+u.ID.String())
	zapiszJSON(w, http.StatusAccepted, s.uruchomienieDTO(u))
}

func (s *Serwer) listaUruchomien(w http.ResponseWriter, r *http.Request) {
	var zestawID *uuid.UUID
	if v := r.URL.Query().Get("zestaw_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			zapiszBlad(w, http.StatusBadRequest, Blad{Kod: "zle_zadanie", Komunikat: "Niepoprawny zestaw_id."})
			return
		}
		zestawID = &id
	}
	lista, err := s.repo.ListaUruchomien(r.Context(), zestawID)
	if err != nil {
		s.bladSerwera(w, r, err)
		return
	}
	wynik := make([]Uruchomienie, len(lista))
	for i, u := range lista {
		wynik[i] = s.uruchomienieDTO(u)
	}
	zapiszJSON(w, http.StatusOK, wynik)
}

// uruchomienie czyta uruchomienie ze ścieżki; przy błędzie odpowiada i zwraca ok = false.
func (s *Serwer) uruchomienie(w http.ResponseWriter, r *http.Request) (store.Uruchomienie, bool) {
	id, ok := idZeSciezki(w, r, "uruchomienieId")
	if !ok {
		return store.Uruchomienie{}, false
	}
	u, err := s.repo.Uruchomienie(r.Context(), id)
	if errors.Is(err, store.ErrNieZnaleziono) {
		zapiszBlad(w, http.StatusNotFound, Blad{Kod: "nie_znaleziono", Komunikat: "Nie ma takiego uruchomienia."})
		return u, false
	}
	if err != nil {
		s.bladSerwera(w, r, err)
		return u, false
	}
	return u, true
}

func (s *Serwer) pobierzUruchomienie(w http.ResponseWriter, r *http.Request) {
	if u, ok := s.uruchomienie(w, r); ok {
		zapiszJSON(w, http.StatusOK, s.uruchomienieDTO(u))
	}
}

func (s *Serwer) zatrzymaj(w http.ResponseWriter, r *http.Request) {
	u, ok := s.uruchomienie(w, r)
	if !ok {
		return
	}
	jestWPuli := false
	if !zakonczone(u.Status) {
		var err error
		if jestWPuli, err = s.pula.Zatrzymaj(r.Context(), u.ID); err != nil {
			s.bladSerwera(w, r, err)
			return
		}
	}
	if !jestWPuli {
		zapiszBlad(w, http.StatusConflict, Blad{Kod: "zakonczone", Komunikat: "Uruchomienie już się zakończyło."})
		return
	}
	if teraz, err := s.repo.Uruchomienie(r.Context(), u.ID); err == nil {
		u = teraz
	}
	zapiszJSON(w, http.StatusAccepted, s.uruchomienieDTO(u))
}

func (s *Serwer) pobierzPlan(w http.ResponseWriter, r *http.Request) {
	s.pobierzWynik(w, r, s.repo.Plan)
}

func (s *Serwer) pobierzRaport(w http.ResponseWriter, r *http.Request) {
	s.pobierzWynik(w, r, s.repo.Raport)
}

func (s *Serwer) pobierzWynik(w http.ResponseWriter, r *http.Request, pobierz func(context.Context, uuid.UUID) (json.RawMessage, error)) {
	u, ok := s.uruchomienie(w, r)
	if !ok {
		return
	}
	wynik, err := pobierz(r.Context(), u.ID)
	if errors.Is(err, store.ErrNieZnaleziono) {
		komunikat := "Uruchomienie nie ma planu: jeszcze trwa."
		if zakonczone(u.Status) {
			komunikat = "Uruchomienie nie ma planu: zakończyło się błędem."
		}
		zapiszBlad(w, http.StatusNotFound, Blad{Kod: "brak_planu", Komunikat: komunikat})
		return
	}
	if err != nil {
		s.bladSerwera(w, r, err)
		return
	}
	zapiszSurowy(w, http.StatusOK, wynik)
}

// postep to strumień SSE z postępem uruchomienia.
func (s *Serwer) postep(w http.ResponseWriter, r *http.Request) {
	id, ok := idZeSciezki(w, r, "uruchomienieId")
	if !ok {
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		s.bladSerwera(w, r, errors.New("ResponseWriter nie obsługuje Flush"))
		return
	}

	zdarzenia, przestan, wPuli := s.pula.Subskrybuj(id)
	if !wPuli {
		// Uruchomienie się skończyło (albo nigdy nie trafiło do puli): stan z bazy i koniec.
		u, ok := s.uruchomienie(w, r)
		if !ok {
			return
		}
		naglowkiSSE(w)
		dto := s.uruchomienieDTO(u)
		zapiszZdarzenie(w, ZdarzenieStan, dto)
		if zakonczone(u.Status) {
			zapiszZdarzenie(w, ZdarzenieKoniec, dto)
		}
		flusher.Flush()
		return
	}
	defer przestan()

	naglowkiSSE(w)
	flusher.Flush()
	ping := time.NewTicker(s.ping)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case z, otwarty := <-zdarzenia:
			if !otwarty {
				return
			}
			zapiszZdarzenie(w, z.Typ, z.Dane)
			flusher.Flush()
			if z.Typ == ZdarzenieKoniec {
				return
			}
		}
	}
}

func naglowkiSSE(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream; charset=utf-8")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
}

func zapiszZdarzenie(w io.Writer, typ string, dane any) {
	b, err := json.Marshal(dane)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", typ, b)
}
