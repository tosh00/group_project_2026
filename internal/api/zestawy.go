package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/solver"
	"github.com/tosh00/group_project_2026/internal/store"
)

func (s *Serwer) domyslneWagi(w http.ResponseWriter, _ *http.Request) {
	zapiszJSON(w, http.StatusOK, solver.DomyslneWagi())
}

func (s *Serwer) listaZestawow(w http.ResponseWriter, r *http.Request) {
	lista, err := s.repo.ListaZestawow(r.Context())
	if err != nil {
		s.bladSerwera(w, r, err)
		return
	}
	wynik := make([]Zestaw, len(lista))
	for i, z := range lista {
		wynik[i] = zestawDTO(z)
	}
	zapiszJSON(w, http.StatusOK, wynik)
}

func (s *Serwer) wgrajZestaw(w http.ResponseWriter, r *http.Request) {
	tresc, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaksRozmiarDanych))
	var zaDuze *http.MaxBytesError
	if errors.As(err, &zaDuze) {
		zapiszBlad(w, http.StatusRequestEntityTooLarge, Blad{Kod: "za_duze", Komunikat: "Plik z danymi może mieć najwyżej 10 MB."})
		return
	}
	if err != nil {
		zapiszBlad(w, http.StatusBadRequest, Blad{Kod: "zle_zadanie", Komunikat: "Nie udało się odczytać pliku."})
		return
	}
	if !json.Valid(tresc) {
		zapiszBlad(w, http.StatusBadRequest, Blad{Kod: "zle_zadanie", Komunikat: "Plik nie jest poprawnym JSON-em."})
		return
	}
	dane, err := model.Dekoduj(bytes.NewReader(tresc))
	if err != nil {
		zapiszBlad(w, http.StatusUnprocessableEntity, Blad{
			Kod: "niespojne_dane", Komunikat: "Dane są niezgodne z formatem dane_planu.json.", Szczegoly: szczegolyBledu(err),
		})
		return
	}

	nazwa := strings.TrimSpace(r.URL.Query().Get("nazwa"))
	if nazwa == "" {
		nazwa = strings.TrimSpace(dane.Meta.Opis)
	}
	if nazwa == "" {
		nazwa = "dane_planu"
	}
	if len([]rune(nazwa)) > 200 {
		nazwa = string([]rune(nazwa)[:200])
	}

	podsumowanie, ostrzezenia := podsumuj(dane)
	pJSON, _ := json.Marshal(podsumowanie)
	oJSON, _ := json.Marshal(ostrzezenia)
	z, err := s.repo.DodajZestaw(r.Context(), nazwa, tresc, pJSON, oJSON)
	if err != nil {
		s.bladSerwera(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/zestawy/"+z.ID.String())
	zapiszJSON(w, http.StatusCreated, zestawDTO(z))
}

func (s *Serwer) pobierzZestaw(w http.ResponseWriter, r *http.Request) {
	id, ok := idZeSciezki(w, r, "zestawId")
	if !ok {
		return
	}
	z, err := s.repo.Zestaw(r.Context(), id)
	if errors.Is(err, store.ErrNieZnaleziono) {
		zapiszBlad(w, http.StatusNotFound, Blad{Kod: "nie_znaleziono", Komunikat: "Nie ma takiego zestawu danych."})
		return
	}
	if err != nil {
		s.bladSerwera(w, r, err)
		return
	}
	zapiszJSON(w, http.StatusOK, zestawDTO(z))
}

func (s *Serwer) pobierzDaneZestawu(w http.ResponseWriter, r *http.Request) {
	id, ok := idZeSciezki(w, r, "zestawId")
	if !ok {
		return
	}
	dane, err := s.repo.DaneZestawu(r.Context(), id)
	if errors.Is(err, store.ErrNieZnaleziono) {
		zapiszBlad(w, http.StatusNotFound, Blad{Kod: "nie_znaleziono", Komunikat: "Nie ma takiego zestawu danych."})
		return
	}
	if err != nil {
		s.bladSerwera(w, r, err)
		return
	}
	zapiszSurowy(w, http.StatusOK, dane)
}
