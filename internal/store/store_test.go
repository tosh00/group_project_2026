package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/tosh00/group_project_2026/internal/store"
	"github.com/tosh00/group_project_2026/internal/store/storetest"
)

func TestZestawy(t *testing.T) {
	s := storetest.NowaBaza(t)
	ctx := context.Background()

	z, err := s.DodajZestaw(ctx, "zima", json.RawMessage(`{"sale":[{"id":"A1"}]}`),
		json.RawMessage(`{"zajecia":1}`), json.RawMessage(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	if z.ID == uuid.Nil || z.Nazwa != "zima" || z.Utworzono.IsZero() {
		t.Fatalf("zestaw = %+v", z)
	}

	pobrany, err := s.Zestaw(ctx, z.ID)
	if err != nil || pobrany.ID != z.ID || string(pobrany.Podsumowanie) != `{"zajecia": 1}` {
		t.Fatalf("Zestaw = %+v, %v", pobrany, err)
	}
	dane, err := s.DaneZestawu(ctx, z.ID)
	if err != nil || string(dane) != `{"sale": [{"id": "A1"}]}` {
		t.Fatalf("DaneZestawu = %s, %v", dane, err)
	}
	lista, err := s.ListaZestawow(ctx)
	if err != nil || len(lista) != 1 {
		t.Fatalf("ListaZestawow = %v, %v", lista, err)
	}

	if _, err := s.Zestaw(ctx, uuid.New()); !errors.Is(err, store.ErrNieZnaleziono) {
		t.Errorf("brak zestawu: err = %v, chcę ErrNieZnaleziono", err)
	}
	if _, err := s.DaneZestawu(ctx, uuid.New()); !errors.Is(err, store.ErrNieZnaleziono) {
		t.Errorf("brak danych: err = %v, chcę ErrNieZnaleziono", err)
	}
}

func TestCyklUruchomienia(t *testing.T) {
	s := storetest.NowaBaza(t)
	ctx := context.Background()
	z, err := s.DodajZestaw(ctx, "z", json.RawMessage(`{}`), json.RawMessage(`{}`), json.RawMessage(`[]`))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.DodajUruchomienie(ctx, uuid.New(), json.RawMessage(`{}`), 1, 0); !errors.Is(err, store.ErrNieZnaleziono) {
		t.Fatalf("uruchomienie bez zestawu: err = %v, chcę ErrNieZnaleziono", err)
	}

	u, err := s.DodajUruchomienie(ctx, z.ID, json.RawMessage(`{"po_17":4}`), 42, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if u.Status != store.StatusWKolejce || u.Seed != 42 || u.Iteracje != 1000 || u.Rozpoczeto != nil {
		t.Fatalf("nowe uruchomienie = %+v", u)
	}
	if _, err := s.Plan(ctx, u.ID); !errors.Is(err, store.ErrNieZnaleziono) {
		t.Fatalf("plan przed końcem: err = %v", err)
	}

	u, err = s.Rozpocznij(ctx, u.ID)
	if err != nil || u.Status != store.StatusTrwa || u.Rozpoczeto == nil {
		t.Fatalf("Rozpocznij = %+v, %v", u, err)
	}
	if _, err := s.Rozpocznij(ctx, u.ID); !errors.Is(err, store.ErrZlyStan) {
		t.Fatalf("ponowne Rozpocznij: err = %v, chcę ErrZlyStan", err)
	}

	naruszenia, koszt := int32(0), 412.5
	u, err = s.Zakoncz(ctx, store.Zakonczenie{
		ID: u.ID, Status: store.StatusPoprawny, LiczbaNaruszen: &naruszenia, Koszt: &koszt,
		Plan: json.RawMessage(`{"meta":{}}`), Raport: json.RawMessage(`{"status":"poprawny"}`),
	})
	if err != nil || u.Status != store.StatusPoprawny || u.Zakonczono == nil || *u.Koszt != koszt {
		t.Fatalf("Zakoncz = %+v, %v", u, err)
	}
	if p, err := s.Plan(ctx, u.ID); err != nil || string(p) != `{"meta": {}}` {
		t.Fatalf("Plan = %s, %v", p, err)
	}
	if r, err := s.Raport(ctx, u.ID); err != nil || string(r) != `{"status": "poprawny"}` {
		t.Fatalf("Raport = %s, %v", r, err)
	}
	blad := "x"
	if _, err := s.Zakoncz(ctx, store.Zakonczenie{ID: u.ID, Status: store.StatusBlad, Blad: &blad}); !errors.Is(err, store.ErrZlyStan) {
		t.Fatalf("ponowne Zakoncz: err = %v, chcę ErrZlyStan", err)
	}

	tegoZestawu, err := s.ListaUruchomien(ctx, &z.ID)
	if err != nil || len(tegoZestawu) != 1 {
		t.Fatalf("ListaUruchomien(zestaw) = %d, %v", len(tegoZestawu), err)
	}
	inny := uuid.New()
	if l, err := s.ListaUruchomien(ctx, &inny); err != nil || len(l) != 0 {
		t.Fatalf("ListaUruchomien(inny) = %d, %v", len(l), err)
	}
}

func TestPrzerwijNiedokonczone(t *testing.T) {
	s := storetest.NowaBaza(t)
	ctx := context.Background()
	z, err := s.DodajZestaw(ctx, "z", json.RawMessage(`{}`), json.RawMessage(`{}`), json.RawMessage(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	czeka, _ := s.DodajUruchomienie(ctx, z.ID, json.RawMessage(`{}`), 1, 0)
	trwa, _ := s.DodajUruchomienie(ctx, z.ID, json.RawMessage(`{}`), 2, 0)
	if _, err := s.Rozpocznij(ctx, trwa.ID); err != nil {
		t.Fatal(err)
	}
	gotowe, _ := s.DodajUruchomienie(ctx, z.ID, json.RawMessage(`{}`), 3, 0)
	zero := int32(0)
	if _, err := s.Zakoncz(ctx, store.Zakonczenie{ID: gotowe.ID, Status: store.StatusPoprawny, LiczbaNaruszen: &zero,
		Plan: json.RawMessage(`{}`), Raport: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}

	n, err := s.PrzerwijNiedokonczone(ctx, "restart")
	if err != nil || n != 2 {
		t.Fatalf("PrzerwijNiedokonczone = %d, %v; chcę 2", n, err)
	}
	for _, id := range []uuid.UUID{czeka.ID, trwa.ID} {
		u, _ := s.Uruchomienie(ctx, id)
		if u.Status != store.StatusBlad || u.Blad == nil || *u.Blad != "restart" {
			t.Errorf("po przerwaniu %v = %+v", id, u)
		}
	}
	if u, _ := s.Uruchomienie(ctx, gotowe.ID); u.Status != store.StatusPoprawny {
		t.Errorf("zakończone zmienione: %+v", u)
	}
}

func TestMigracjePowtorne(t *testing.T) {
	s := storetest.NowaBaza(t)
	if err := s.Migruj(context.Background()); err != nil {
		t.Fatalf("drugie Migruj: %v", err)
	}
}
