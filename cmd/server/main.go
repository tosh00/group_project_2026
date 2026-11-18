// Polecenie server uruchamia API planera (api/openapi.yaml) z pulą workerów solvera.
//
//	DATABASE_URL=postgres://planer:planer@localhost:55432/planer?sslmode=disable go run ./cmd/server
//
// Ustawienia (zmienne środowiskowe albo flagi; flaga wygrywa):
//
//	DATABASE_URL  -db       adres PostgreSQL (wymagany)
//	ADDR          -addr     adres nasłuchu, domyślnie :8080
//	WORKERY       -workery  równoległe uruchomienia solvera, domyślnie 1 (każde zajmuje 4 rdzenie)
//	KOLEJKA       -kolejka  ile uruchomień może czekać, domyślnie 16
//
// Przy starcie serwer stosuje migracje i oznacza jako błąd uruchomienia przerwane poprzednim zamknięciem.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/tosh00/group_project_2026/internal/api"
	"github.com/tosh00/group_project_2026/internal/store"
)

func main() {
	addr := flag.String("addr", env("ADDR", ":8080"), "adres nasłuchu")
	dbURL := flag.String("db", os.Getenv("DATABASE_URL"), "adres PostgreSQL")
	workery := flag.Int("workery", envInt("WORKERY", 1), "równoległe uruchomienia solvera")
	kolejka := flag.Int("kolejka", envInt("KOLEJKA", 16), "maksymalna liczba czekających uruchomień")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := uruchom(log, *addr, *dbURL, *workery, *kolejka); err != nil {
		log.Error("server", "blad", err)
		os.Exit(1)
	}
}

func uruchom(log *slog.Logger, addr, dbURL string, workery, kolejka int) error {
	if dbURL == "" {
		return errors.New("brak adresu bazy: ustaw DATABASE_URL albo -db")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.Polacz(ctx, dbURL)
	if err != nil {
		return err
	}
	defer st.Zamknij()
	if err := st.Migruj(ctx); err != nil {
		return err
	}
	if n, err := st.PrzerwijNiedokonczone(ctx, "przerwane restartem serwera"); err != nil {
		return err
	} else if n > 0 {
		log.Warn("uruchomienia przerwane poprzednim zamknięciem serwera", "liczba", n)
	}

	// Do czasu merge'a internal/validator (docs/contracts/api.md).
	walidator := api.Atrapa{}
	log.Warn("walidator: atrapa, raporty nie sprawdzają reguł twardych")

	pula := api.NowaPula(st, walidator, log, api.UstawieniaPuli{Workery: workery, Kolejka: kolejka})
	serwer := api.Nowy(st, pula, walidator, log)

	bazowy, anulujBazowy := context.WithCancel(context.Background())
	defer anulujBazowy()
	srv := &http.Server{
		Addr:              addr,
		Handler:           serwer.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second, // wgranie pliku do 10 MB
		IdleTimeout:       120 * time.Second,
		// Bez WriteTimeout: strumień SSE trwa tyle, co uruchomienie.
		BaseContext: func(net.Listener) context.Context { return bazowy },
	}
	srv.RegisterOnShutdown(anulujBazowy) // kończy otwarte strumienie SSE

	bledy := make(chan error, 1)
	go func() {
		log.Info("nasłuch", "addr", addr, "workery", workery, "kolejka", kolejka)
		bledy <- srv.ListenAndServe()
	}()

	select {
	case err := <-bledy:
		return fmt.Errorf("serwer HTTP: %w", err)
	case <-ctx.Done():
	}
	log.Info("zamykanie")
	koniec, anuluj := context.WithTimeout(context.Background(), 15*time.Second)
	defer anuluj()
	if err := srv.Shutdown(koniec); err != nil {
		log.Error("zamykanie serwera HTTP", "blad", err)
	}
	if err := pula.Zamknij(koniec); err != nil {
		log.Error("zamykanie puli", "blad", err)
	}
	return nil
}

func env(nazwa, domyslna string) string {
	if v := os.Getenv(nazwa); v != "" {
		return v
	}
	return domyslna
}

func envInt(nazwa string, domyslna int) int {
	if v, err := strconv.Atoi(os.Getenv(nazwa)); err == nil {
		return v
	}
	return domyslna
}
