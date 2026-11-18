// Package storetest daje testom świeżą bazę PostgreSQL z migracjami.
//
// Testy z bazą uruchamiają się tylko z ustawionym TEST_DATABASE_URL (w przeciwnym razie są pomijane), np.:
//
//	docker run -d --name planer-pg -e POSTGRES_USER=planer -e POSTGRES_PASSWORD=planer -p 55432:5432 postgres:16
//	TEST_DATABASE_URL=postgres://planer:planer@localhost:55432/planer?sslmode=disable go test ./...
package storetest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/tosh00/group_project_2026/internal/store"
)

// NowaBaza tworzy osobną bazę dla testu, stosuje migracje i usuwa bazę po teście.
func NowaBaza(t testing.TB) *store.Store {
	t.Helper()
	adres := os.Getenv("TEST_DATABASE_URL")
	if adres == "" {
		t.Skip("brak TEST_DATABASE_URL: pomijam test z bazą")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, adres)
	if err != nil {
		t.Fatalf("połączenie z TEST_DATABASE_URL: %v", err)
	}
	nazwa := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+nazwa); err != nil {
		admin.Close(ctx)
		t.Fatalf("tworzenie bazy testowej: %v", err)
	}

	u, err := url.Parse(adres)
	if err != nil {
		t.Fatalf("TEST_DATABASE_URL: %v", err)
	}
	u.Path = "/" + nazwa
	s, err := store.Polacz(ctx, u.String())
	if err != nil {
		t.Fatalf("połączenie z bazą testową: %v", err)
	}
	t.Cleanup(func() {
		s.Zamknij()
		if _, err := admin.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", nazwa)); err != nil {
			t.Errorf("usuwanie bazy testowej: %v", err)
		}
		admin.Close(ctx)
	})
	if err := s.Migruj(ctx); err != nil {
		t.Fatalf("migracje: %v", err)
	}
	return s
}
