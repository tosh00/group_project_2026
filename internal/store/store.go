// Package store to dostęp do PostgreSQL: migracje (goose) i zapytania (sqlc, pakiet db).
//
// Schemat i zasady: ADR 0008 (sqlc + pgx + goose) i ADR 0009 (dane, plan i raport jako jsonb).
// Zapytania są w zapytania/*.sql; po ich zmianie trzeba wygenerować db/ ponownie (patrz sqlc.yaml).
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/tosh00/group_project_2026/internal/store/db"
	"github.com/tosh00/group_project_2026/migrations"
)

var (
	// ErrNieZnaleziono: brak wiersza o danym id (albo brak zestawu przy dodawaniu uruchomienia).
	ErrNieZnaleziono = errors.New("nie znaleziono")
	// ErrZlyStan: uruchomienie nie jest w stanie, który pozwala na zmianę (np. już zakończone).
	ErrZlyStan = errors.New("niedozwolona zmiana stanu uruchomienia")
	// ErrTrwaGenerowanie: zestawu nie można usunąć, bo ma uruchomienie w kolejce albo trwające.
	ErrTrwaGenerowanie = errors.New("trwa generowanie planu dla zestawu")
)

// Statusy uruchomienia (kolumna uruchomienia.status, API: StatusUruchomienia).
const (
	StatusWKolejce      = "w_kolejce"
	StatusTrwa          = "trwa"
	StatusPoprawny      = "poprawny"
	StatusZNaruszeniami = "z_naruszeniami"
	StatusBlad          = "blad"
)

// Uruchomienie to wiersz tabeli uruchomienia.
type Uruchomienie = db.Uruchomienie

// Zestaw to wgrany zestaw danych bez samych danych (te zwraca DaneZestawu).
type Zestaw struct {
	ID           uuid.UUID
	Nazwa        string
	Podsumowanie json.RawMessage
	Ostrzezenia  json.RawMessage
	Utworzono    time.Time
}

// Zakonczenie to stan końcowy uruchomienia. Plan i Raport są puste, gdy uruchomienie skończyło się błędem.
type Zakonczenie struct {
	ID             uuid.UUID
	Status         string
	Zatrzymane     bool
	LiczbaNaruszen *int32
	Koszt          *float64
	Blad           *string
	Plan           json.RawMessage
	Raport         json.RawMessage
}

// Store to połączenie z bazą. Bezpieczny współbieżnie.
type Store struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// Polacz otwiera pulę połączeń i sprawdza, czy baza odpowiada.
func Polacz(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("konfiguracja bazy: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("połączenie z bazą: %w", err)
	}
	return &Store{pool: pool, q: db.New(pool)}, nil
}

// Zamknij zamyka pulę połączeń.
func (s *Store) Zamknij() { s.pool.Close() }

// Ping sprawdza połączenie z bazą (GET /api/zdrowie).
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// Migruj stosuje brakujące migracje z katalogu migrations/ (osadzone w binarce).
func (s *Store) Migruj(ctx context.Context) error {
	sqlDB := stdlib.OpenDBFromPool(s.pool)
	defer sqlDB.Close()
	return migruj(ctx, sqlDB)
}

func migruj(ctx context.Context, sqlDB *sql.DB) error {
	prov, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS)
	if err != nil {
		return fmt.Errorf("migracje: %w", err)
	}
	if _, err := prov.Up(ctx); err != nil {
		return fmt.Errorf("migracje: %w", err)
	}
	return nil
}

// DodajZestaw zapisuje wgrany plik (już sprawdzony przez model.Dekoduj) z podsumowaniem i ostrzeżeniami.
func (s *Store) DodajZestaw(ctx context.Context, nazwa string, dane, podsumowanie, ostrzezenia json.RawMessage) (Zestaw, error) {
	r, err := s.q.DodajZestaw(ctx, db.DodajZestawParams{
		Nazwa: nazwa, Dane: dane, Podsumowanie: podsumowanie, Ostrzezenia: ostrzezenia,
	})
	if err != nil {
		return Zestaw{}, err
	}
	return Zestaw(r), nil
}

// ListaZestawow zwraca zestawy od najnowszego.
func (s *Store) ListaZestawow(ctx context.Context) ([]Zestaw, error) {
	rows, err := s.q.ListaZestawow(ctx)
	if err != nil {
		return nil, err
	}
	wynik := make([]Zestaw, len(rows))
	for i, r := range rows {
		wynik[i] = Zestaw(r)
	}
	return wynik, nil
}

// Zestaw zwraca zestaw bez danych.
func (s *Store) Zestaw(ctx context.Context, id uuid.UUID) (Zestaw, error) {
	r, err := s.q.PobierzZestaw(ctx, id)
	if err != nil {
		return Zestaw{}, mapuj(err)
	}
	return Zestaw(r), nil
}

// DaneZestawu zwraca treść dane_planu.json zestawu.
func (s *Store) DaneZestawu(ctx context.Context, id uuid.UUID) (json.RawMessage, error) {
	d, err := s.q.PobierzDaneZestawu(ctx, id)
	return d, mapuj(err)
}

// UsunZestaw usuwa zestaw razem z uruchomieniami i planami. Zestaw, dla którego trwa generowanie,
// zostaje (ErrTrwaGenerowanie); brak zestawu to ErrNieZnaleziono.
func (s *Store) UsunZestaw(ctx context.Context, id uuid.UUID) error {
	usuniete, err := s.q.UsunZestaw(ctx, id)
	if err != nil {
		return err
	}
	if usuniete > 0 {
		return nil
	}

	if _, err := s.q.PobierzZestaw(ctx, id); err != nil {
		return mapuj(err)
	}
	return ErrTrwaGenerowanie
}

// DodajUruchomienie tworzy uruchomienie w kolejce. Brak zestawu to ErrNieZnaleziono.
func (s *Store) DodajUruchomienie(ctx context.Context, zestawID uuid.UUID, wagi json.RawMessage, seed int64, iteracje int32) (Uruchomienie, error) {
	u, err := s.q.DodajUruchomienie(ctx, db.DodajUruchomienieParams{
		ZestawID: zestawID, Wagi: wagi, Seed: seed, Iteracje: iteracje,
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
		return Uruchomienie{}, ErrNieZnaleziono
	}
	return u, err
}

// Uruchomienie zwraca uruchomienie po id.
func (s *Store) Uruchomienie(ctx context.Context, id uuid.UUID) (Uruchomienie, error) {
	u, err := s.q.PobierzUruchomienie(ctx, id)
	return u, mapuj(err)
}

// ListaUruchomien zwraca uruchomienia od najnowszego; zestawID == nil to wszystkie.
func (s *Store) ListaUruchomien(ctx context.Context, zestawID *uuid.UUID) ([]Uruchomienie, error) {
	return s.q.ListaUruchomien(ctx, zestawID)
}

// Rozpocznij przestawia uruchomienie z w_kolejce na trwa. ErrZlyStan, jeśli już nie czeka.
func (s *Store) Rozpocznij(ctx context.Context, id uuid.UUID) (Uruchomienie, error) {
	u, err := s.q.RozpocznijUruchomienie(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Uruchomienie{}, ErrZlyStan
	}
	return u, err
}

// Zakoncz zapisuje stan końcowy i (jeśli jest) plan z raportem w jednej transakcji.
// ErrZlyStan, jeśli uruchomienie było już zakończone.
func (s *Store) Zakoncz(ctx context.Context, z Zakonczenie) (Uruchomienie, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Uruchomienie{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // po Commit nic nie robi
	q := s.q.WithTx(tx)
	u, err := q.ZakonczUruchomienie(ctx, db.ZakonczUruchomienieParams{
		ID: z.ID, Status: z.Status, Zatrzymane: z.Zatrzymane,
		LiczbaNaruszen: z.LiczbaNaruszen, Koszt: z.Koszt, Blad: z.Blad,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Uruchomienie{}, ErrZlyStan
	}
	if err != nil {
		return Uruchomienie{}, err
	}
	if z.Plan != nil {
		if err := q.ZapiszPlan(ctx, db.ZapiszPlanParams{UruchomienieID: z.ID, Plan: z.Plan, Raport: z.Raport}); err != nil {
			return Uruchomienie{}, err
		}
	}
	return u, tx.Commit(ctx)
}

// PrzerwijNiedokonczone kończy błędem uruchomienia w_kolejce i trwa (przy starcie serwera).
func (s *Store) PrzerwijNiedokonczone(ctx context.Context, powod string) (int64, error) {
	return s.q.PrzerwijNiedokonczone(ctx, &powod)
}

// Plan zwraca plan.json uruchomienia; ErrNieZnaleziono, jeśli planu nie ma.
func (s *Store) Plan(ctx context.Context, uruchomienieID uuid.UUID) (json.RawMessage, error) {
	p, err := s.q.PobierzPlan(ctx, uruchomienieID)
	return p, mapuj(err)
}

// Raport zwraca raport walidatora uruchomienia; ErrNieZnaleziono, jeśli planu nie ma.
func (s *Store) Raport(ctx context.Context, uruchomienieID uuid.UUID) (json.RawMessage, error) {
	r, err := s.q.PobierzRaport(ctx, uruchomienieID)
	return r, mapuj(err)
}

func mapuj(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNieZnaleziono
	}
	return err
}
