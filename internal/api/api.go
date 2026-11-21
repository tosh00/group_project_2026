// Package api to serwer HTTP planera: handlery z api/openapi.yaml, pula workerów solvera i SSE z postępem.
//
// Router: chi (ADR 0007). Baza: internal/store (ADR 0008, 0009). Opis kontraktu: docs/contracts/api.md.
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/tosh00/group_project_2026/internal/store"
)

// Repozytorium to operacje na bazie używane przez API; implementuje je *store.Store.
type Repozytorium interface {
	Ping(ctx context.Context) error
	DodajZestaw(ctx context.Context, nazwa string, dane, podsumowanie, ostrzezenia json.RawMessage) (store.Zestaw, error)
	ListaZestawow(ctx context.Context) ([]store.Zestaw, error)
	Zestaw(ctx context.Context, id uuid.UUID) (store.Zestaw, error)
	DaneZestawu(ctx context.Context, id uuid.UUID) (json.RawMessage, error)
	UsunZestaw(ctx context.Context, id uuid.UUID) error
	DodajUruchomienie(ctx context.Context, zestawID uuid.UUID, wagi json.RawMessage, seed int64, iteracje int32) (store.Uruchomienie, error)
	Uruchomienie(ctx context.Context, id uuid.UUID) (store.Uruchomienie, error)
	ListaUruchomien(ctx context.Context, zestawID *uuid.UUID) ([]store.Uruchomienie, error)
	Rozpocznij(ctx context.Context, id uuid.UUID) (store.Uruchomienie, error)
	Zakoncz(ctx context.Context, z store.Zakonczenie) (store.Uruchomienie, error)
	Plan(ctx context.Context, uruchomienieID uuid.UUID) (json.RawMessage, error)
	Raport(ctx context.Context, uruchomienieID uuid.UUID) (json.RawMessage, error)
}

var _ Repozytorium = (*store.Store)(nil)

// Limity rozmiaru treści żądań.
const (
	MaksRozmiarDanych       = 10 << 20 // POST /zestawy
	MaksRozmiarUruchomienia = 64 << 10 // POST /zestawy/{id}/uruchomienia
)

// Serwer obsługuje API. Tworzy go Nowy; pulę zamyka właściciel (cmd/server).
type Serwer struct {
	repo      Repozytorium
	pula      *Pula
	walidator Walidator
	log       *slog.Logger
	ping      time.Duration // co ile komentarz ": ping" w SSE
}

// Nowy tworzy serwer API.
func Nowy(repo Repozytorium, pula *Pula, w Walidator, log *slog.Logger) *Serwer {
	return &Serwer{repo: repo, pula: pula, walidator: w, log: log, ping: 15 * time.Second}
}

// Handler zwraca router z trasami pod /api.
func (s *Serwer) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, s.logowanie, middleware.Recoverer)
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		zapiszBlad(w, http.StatusNotFound, Blad{Kod: "nie_znaleziono", Komunikat: "Nie ma takiej ścieżki."})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		zapiszBlad(w, http.StatusMethodNotAllowed, Blad{Kod: "zle_zadanie", Komunikat: "Ta metoda nie jest obsługiwana dla tej ścieżki."})
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/zdrowie", s.zdrowie)
		r.Get("/wagi/domyslne", s.domyslneWagi)

		r.Get("/zestawy", s.listaZestawow)
		r.Post("/zestawy", s.wgrajZestaw)
		r.Get("/zestawy/{zestawId}", s.pobierzZestaw)
		r.Delete("/zestawy/{zestawId}", s.usunZestaw)
		r.Get("/zestawy/{zestawId}/dane", s.pobierzDaneZestawu)
		r.Post("/zestawy/{zestawId}/uruchomienia", s.uruchomSolver)

		r.Get("/uruchomienia", s.listaUruchomien)
		r.Get("/uruchomienia/{uruchomienieId}", s.pobierzUruchomienie)
		r.Post("/uruchomienia/{uruchomienieId}/zatrzymaj", s.zatrzymaj)
		r.Get("/uruchomienia/{uruchomienieId}/postep", s.postep)
		r.Get("/uruchomienia/{uruchomienieId}/plan", s.pobierzPlan)
		r.Get("/uruchomienia/{uruchomienieId}/raport", s.pobierzRaport)
	})
	return r
}

func (s *Serwer) logowanie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		s.log.Info("http", "metoda", r.Method, "sciezka", r.URL.Path, "kod", ww.Status(),
			"czas", time.Since(start).Round(time.Microsecond), "id", middleware.GetReqID(r.Context()))
	})
}

// idZeSciezki czyta UUID z parametru ścieżki; przy błędzie odpowiada 404.
func idZeSciezki(w http.ResponseWriter, r *http.Request, nazwa string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, nazwa))
	if err != nil {
		zapiszBlad(w, http.StatusNotFound, Blad{Kod: "nie_znaleziono", Komunikat: "Niepoprawny identyfikator."})
		return uuid.Nil, false
	}
	return id, true
}

func (s *Serwer) bladSerwera(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("błąd obsługi żądania", "sciezka", r.URL.Path, "blad", err, "id", middleware.GetReqID(r.Context()))
	zapiszBlad(w, http.StatusInternalServerError, Blad{Kod: "blad_serwera", Komunikat: "Wewnętrzny błąd serwera."})
}

func (s *Serwer) zdrowie(w http.ResponseWriter, r *http.Request) {
	odp := Zdrowie{Status: "ok", Baza: "ok", Walidator: s.walidator.Nazwa()}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.repo.Ping(ctx); err != nil {
		odp.Status, odp.Baza = "blad", "blad"
		zapiszJSON(w, http.StatusServiceUnavailable, odp)
		return
	}
	zapiszJSON(w, http.StatusOK, odp)
}
