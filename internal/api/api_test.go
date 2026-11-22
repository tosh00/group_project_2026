package api_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tosh00/group_project_2026/internal/api"
	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/store/storetest"
)

// Testy z bazą: TEST_DATABASE_URL (patrz internal/store/storetest). Solver działa na prawdziwych danych.

type srodowisko struct {
	t    *testing.T
	url  string
	dane []byte
}

func nowe(t *testing.T, ust api.UstawieniaPuli) *srodowisko {
	t.Helper()
	st := storetest.NowaBaza(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if ust.CoIlePostep == 0 {
		ust.CoIlePostep = 10 * time.Millisecond
	}
	pula := api.NowaPula(st, api.PelnyWalidator{}, log, ust)
	srv := httptest.NewServer(api.Nowy(st, pula, api.PelnyWalidator{}, log).Handler())
	t.Cleanup(func() {
		srv.CloseClientConnections()
		srv.Close()
		_ = pula.Zamknij(context.Background())
	})
	dane, err := os.ReadFile("../../data/dane_planu.json")
	if err != nil {
		t.Fatal(err)
	}
	return &srodowisko{t: t, url: srv.URL + "/api", dane: dane}
}

// zadanie wysyła żądanie i dekoduje odpowiedź JSON do v (jeśli v != nil). Zwraca kod HTTP.
func (s *srodowisko) zadanie(metoda, sciezka string, tresc []byte, v any) int {
	s.t.Helper()
	req, err := http.NewRequest(metoda, s.url+sciezka, bytes.NewReader(tresc))
	if err != nil {
		s.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if v != nil {
		if err := json.Unmarshal(b, v); err != nil {
			s.t.Fatalf("%s %s: %v; treść %s", metoda, sciezka, err, b)
		}
	}
	return resp.StatusCode
}

func (s *srodowisko) wgraj() api.Zestaw {
	s.t.Helper()
	var z api.Zestaw
	if kod := s.zadanie("POST", "/zestawy?nazwa=zima", s.dane, &z); kod != http.StatusCreated {
		s.t.Fatalf("wgranie: kod %d", kod)
	}
	return z
}

func (s *srodowisko) uruchom(zestaw api.Zestaw, ustawienia string) api.Uruchomienie {
	s.t.Helper()
	var u api.Uruchomienie
	if kod := s.zadanie("POST", "/zestawy/"+zestaw.ID.String()+"/uruchomienia", []byte(ustawienia), &u); kod != http.StatusAccepted {
		s.t.Fatalf("uruchomienie: kod %d", kod)
	}
	return u
}

type zdarzenie struct {
	typ  string
	dane json.RawMessage
}

// sluchaj czyta zdarzenia SSE i przekazuje je na kanał, aż strumień się skończy.
func (s *srodowisko) sluchaj(id string) <-chan zdarzenie {
	s.t.Helper()
	resp, err := http.Get(s.url + "/uruchomienia/" + id + "/postep")
	if err != nil {
		s.t.Fatal(err)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		s.t.Fatalf("Content-Type %q", ct)
	}
	ch := make(chan zdarzenie, 1024)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		sc := bufio.NewScanner(resp.Body)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		var z zdarzenie
		for sc.Scan() {
			linia := sc.Text()
			switch {
			case strings.HasPrefix(linia, "event: "):
				z.typ = strings.TrimPrefix(linia, "event: ")
			case strings.HasPrefix(linia, "data: "):
				z.dane = json.RawMessage(strings.TrimPrefix(linia, "data: "))
			case linia == "" && z.typ != "":
				ch <- z
				z = zdarzenie{}
			}
		}
	}()
	return ch
}

// doKonca zbiera zdarzenia do "koniec" (z limitem czasu) i zwraca końcowe uruchomienie.
func doKonca(t *testing.T, ch <-chan zdarzenie) (api.Uruchomienie, map[string]int) {
	t.Helper()
	liczniki := map[string]int{}
	limit := time.After(60 * time.Second)
	for {
		select {
		case z, ok := <-ch:
			if !ok {
				t.Fatalf("strumień zamknięty bez zdarzenia koniec (%v)", liczniki)
			}
			liczniki[z.typ]++
			if z.typ == api.ZdarzenieKoniec {
				var u api.Uruchomienie
				if err := json.Unmarshal(z.dane, &u); err != nil {
					t.Fatal(err)
				}
				return u, liczniki
			}
		case <-limit:
			t.Fatalf("brak zdarzenia koniec po 60 s (%v)", liczniki)
		}
	}
}

func TestWgranieDanych(t *testing.T) {
	s := nowe(t, api.UstawieniaPuli{})
	z := s.wgraj()
	if z.Nazwa != "zima" {
		t.Errorf("nazwa %q", z.Nazwa)
	}
	var p api.Podsumowanie
	if err := json.Unmarshal(z.Podsumowanie, &p); err != nil || p.Zajecia != 174 || p.SaleDoPlanowania != 30 {
		t.Errorf("podsumowanie %s (%v)", z.Podsumowanie, err)
	}
	var o []api.Ostrzezenie
	if err := json.Unmarshal(z.Ostrzezenia, &o); err != nil || len(o) != 3 {
		t.Errorf("ostrzeżenia %s", z.Ostrzezenia)
	}

	var lista []api.Zestaw
	if kod := s.zadanie("GET", "/zestawy", nil, &lista); kod != 200 || len(lista) != 1 || lista[0].ID != z.ID {
		t.Errorf("lista: kod %d, %v", kod, lista)
	}
	var pobrany api.Zestaw
	if kod := s.zadanie("GET", "/zestawy/"+z.ID.String(), nil, &pobrany); kod != 200 || pobrany.ID != z.ID {
		t.Errorf("pobranie: kod %d", kod)
	}

	// Dane wracają w formacie, który model przyjmuje bez zmian.
	resp, err := http.Get(s.url + "/zestawy/" + z.ID.String() + "/dane")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	d, err := model.Dekoduj(resp.Body)
	if err != nil || len(d.Zajecia) != 174 {
		t.Fatalf("dane zestawu: %v", err)
	}
}

func TestUsuniecieZestawu(t *testing.T) {
	s := nowe(t, api.UstawieniaPuli{})
	z := s.wgraj()

	if kod := s.zadanie("DELETE", "/zestawy/"+z.ID.String(), nil, nil); kod != http.StatusNoContent {
		t.Fatalf("usunięcie: kod %d", kod)
	}
	if kod := s.zadanie("GET", "/zestawy/"+z.ID.String(), nil, nil); kod != http.StatusNotFound {
		t.Errorf("zestaw po usunięciu: kod %d", kod)
	}
	var b api.Blad
	if kod := s.zadanie("DELETE", "/zestawy/"+z.ID.String(), nil, &b); kod != http.StatusNotFound || b.Kod != "nie_znaleziono" {
		t.Errorf("ponowne usunięcie: kod %d, %+v", kod, b)
	}
}

func TestWgranieBledy(t *testing.T) {
	s := nowe(t, api.UstawieniaPuli{})
	testy := []struct {
		nazwa  string
		tresc  string
		kod    int
		kodAPI string
	}{
		{"nie JSON", "{", 400, "zle_zadanie"},
		{"pusty", "", 400, "zle_zadanie"},
		{"nieznane pole", `{"nieznane":1}`, 422, "niespojne_dane"},
		{"zły typ", `{"sale":5}`, 422, "niespojne_dane"},
		{"tablica", `[]`, 422, "niespojne_dane"},
	}
	for _, tt := range testy {
		t.Run(tt.nazwa, func(t *testing.T) {
			var b api.Blad
			if kod := s.zadanie("POST", "/zestawy", []byte(tt.tresc), &b); kod != tt.kod || b.Kod != tt.kodAPI {
				t.Errorf("kod %d %q, chcę %d %q (%+v)", kod, b.Kod, tt.kod, tt.kodAPI, b)
			}
		})
	}

	t.Run("niespójne dane: wszystkie problemy", func(t *testing.T) {
		var d map[string]any
		if err := json.Unmarshal(s.dane, &d); err != nil {
			t.Fatal(err)
		}
		d["sale"] = []any{} // zajęcia odwołują się do nieistniejących sal
		tresc, _ := json.Marshal(d)
		var b api.Blad
		if kod := s.zadanie("POST", "/zestawy", tresc, &b); kod != 422 || len(b.Szczegoly) < 2 {
			t.Errorf("kod %d, szczegóły %d", kod, len(b.Szczegoly))
		}
	})

	t.Run("za duży", func(t *testing.T) {
		tresc := append([]byte(`{"x":"`), bytes.Repeat([]byte("a"), api.MaksRozmiarDanych)...)
		tresc = append(tresc, []byte(`"}`)...)
		var b api.Blad
		if kod := s.zadanie("POST", "/zestawy", tresc, &b); kod != 413 || b.Kod != "za_duze" {
			t.Errorf("kod %d %q", kod, b.Kod)
		}
	})
}

func TestUruchomienieDoKonca(t *testing.T) {
	s := nowe(t, api.UstawieniaPuli{})
	z := s.wgraj()

	var b api.Blad
	if kod := s.zadanie("POST", "/zestawy/"+z.ID.String()+"/uruchomienia", []byte(`{"wagi":{"x":1}}`), &b); kod != 400 {
		t.Errorf("złe wagi: kod %d", kod)
	}

	u := s.uruchom(z, `{"seed":42,"iteracje":20000,"wagi":{"po_17":8}}`)
	if u.Seed != 42 || u.Iteracje != 20000 || u.ZestawID != z.ID {
		t.Fatalf("uruchomienie %+v", u)
	}
	var wagi map[string]float64
	if err := json.Unmarshal(u.Wagi, &wagi); err != nil || wagi["po_17"] != 8 || wagi["okienka_studentow"] != 10 {
		t.Errorf("wagi %s", u.Wagi)
	}

	koniec, liczniki := doKonca(t, s.sluchaj(u.ID.String()))
	if koniec.Status != "poprawny" || koniec.Zatrzymane || koniec.Koszt == nil || koniec.LiczbaNaruszen == nil || koniec.Zakonczono == nil {
		t.Fatalf("koniec %+v", koniec)
	}
	if liczniki[api.ZdarzenieStan] < 1 {
		t.Errorf("zdarzenia %v", liczniki)
	}

	var plan model.Plan
	if kod := s.zadanie("GET", "/uruchomienia/"+u.ID.String()+"/plan", nil, &plan); kod != 200 {
		t.Fatalf("plan: kod %d", kod)
	}
	if plan.Meta.WersjaFormatu != 1 || plan.Meta.Seed != 42 || len(plan.Przypisania) != 174 || plan.Meta.Koszt != *koniec.Koszt {
		t.Errorf("plan meta %+v, przypisań %d", plan.Meta, len(plan.Przypisania))
	}

	var raport api.Raport
	if kod := s.zadanie("GET", "/uruchomienia/"+u.ID.String()+"/raport", nil, &raport); kod != 200 {
		t.Fatalf("raport: kod %d", kod)
	}
	suma := 0.0
	for _, v := range raport.KosztKryteria {
		suma += v
	}
	if raport.Walidator != "internal/validator" || raport.Status != "poprawny" || raport.Naruszenia == nil ||
		raport.KosztMiekki != plan.Meta.Koszt || len(raport.KosztKryteria) == 0 || fmt.Sprintf("%.3f", suma) != fmt.Sprintf("%.3f", raport.KosztMiekki) {
		t.Errorf("raport %+v (suma kryteriów %.3f)", raport, suma)
	}

	// Ten sam seed i ustawienia dają ten sam plan (docs/contracts/solver.md).
	u2 := s.uruchom(z, `{"seed":42,"iteracje":20000,"wagi":{"po_17":8}}`)
	doKonca(t, s.sluchaj(u2.ID.String()))
	var plan2 model.Plan
	s.zadanie("GET", "/uruchomienia/"+u2.ID.String()+"/plan", nil, &plan2)
	if plan2.Meta.Koszt != plan.Meta.Koszt {
		t.Errorf("powtórka: koszt %v, wcześniej %v", plan2.Meta.Koszt, plan.Meta.Koszt)
	}

	// Strumień zakończonego uruchomienia: od razu stan i koniec.
	_, liczniki = doKonca(t, s.sluchaj(u.ID.String()))
	if liczniki[api.ZdarzenieStan] != 1 {
		t.Errorf("zakończone: zdarzenia %v", liczniki)
	}

	var lista []api.Uruchomienie
	if kod := s.zadanie("GET", "/uruchomienia?zestaw_id="+z.ID.String(), nil, &lista); kod != 200 || len(lista) != 2 {
		t.Errorf("lista: kod %d, %d", kod, len(lista))
	}
	if kod := s.zadanie("POST", "/uruchomienia/"+u.ID.String()+"/zatrzymaj", nil, &b); kod != 409 || b.Kod != "zakonczone" {
		t.Errorf("zatrzymanie zakończonego: kod %d %q", kod, b.Kod)
	}
}

func TestZatrzymanie(t *testing.T) {
	s := nowe(t, api.UstawieniaPuli{Workery: 1, Kolejka: 1})
	z := s.wgraj()

	dlugie := s.uruchom(z, fmt.Sprintf(`{"seed":1,"iteracje":%d}`, api.MaksIteracji))
	zdarzenia := s.sluchaj(dlugie.ID.String())

	// Czeka na postęp optymalizacji, żeby zatrzymanie było po konstrukcji.
	limit := time.After(30 * time.Second)
	for czeka := true; czeka; {
		select {
		case zd, ok := <-zdarzenia:
			if !ok {
				t.Fatal("strumień zamknięty przed postępem")
			}
			if zd.typ == api.ZdarzeniePostep {
				var p api.Postep
				if err := json.Unmarshal(zd.dane, &p); err != nil || p.Faza != api.FazaOptymalizacja || p.Iteracji != 4*api.MaksIteracji {
					t.Fatalf("postęp %s", zd.dane)
				}
				czeka = false
			}
		case <-limit:
			t.Fatal("brak postępu po 30 s")
		}
	}

	// Plan trwającego uruchomienia jeszcze nie istnieje.
	var b api.Blad
	if kod := s.zadanie("GET", "/uruchomienia/"+dlugie.ID.String()+"/plan", nil, &b); kod != 404 || b.Kod != "brak_planu" {
		t.Errorf("plan w trakcie: kod %d %q", kod, b.Kod)
	}
	var trwa api.Uruchomienie
	s.zadanie("GET", "/uruchomienia/"+dlugie.ID.String(), nil, &trwa)
	if trwa.Status != "trwa" || trwa.Faza != api.FazaOptymalizacja {
		t.Errorf("w trakcie: %+v", trwa)
	}

	// Jedno czeka w kolejce (Kolejka: 1), kolejne jest odrzucane.
	czekajace := s.uruchom(z, `{"iteracje":0}`)
	if kod := s.zadanie("POST", "/zestawy/"+z.ID.String()+"/uruchomienia", []byte(`{"iteracje":0}`), &b); kod != 503 || b.Kod != "kolejka_pelna" {
		t.Errorf("pełna kolejka: kod %d %q", kod, b.Kod)
	}

	var u api.Uruchomienie
	if kod := s.zadanie("POST", "/uruchomienia/"+czekajace.ID.String()+"/zatrzymaj", nil, &u); kod != 202 {
		t.Fatalf("zatrzymanie czekającego: kod %d", kod)
	}
	if u.Status != "blad" || !u.Zatrzymane {
		t.Errorf("zatrzymane w kolejce: %+v", u)
	}

	if kod := s.zadanie("POST", "/uruchomienia/"+dlugie.ID.String()+"/zatrzymaj", nil, &u); kod != 202 {
		t.Fatalf("zatrzymanie trwającego: kod %d", kod)
	}
	koniec, _ := doKonca(t, zdarzenia)
	if koniec.Status != "poprawny" || !koniec.Zatrzymane || koniec.Koszt == nil {
		t.Fatalf("zatrzymane w trakcie: %+v", koniec)
	}
	if kod := s.zadanie("GET", "/uruchomienia/"+dlugie.ID.String()+"/plan", nil, nil); kod != 200 {
		t.Errorf("plan zatrzymanego: kod %d", kod)
	}
}

func TestNieZnaleziono(t *testing.T) {
	s := nowe(t, api.UstawieniaPuli{})
	brak := "00000000-0000-0000-0000-000000000001"
	for _, sciezka := range []string{
		"/zestawy/" + brak, "/zestawy/" + brak + "/dane", "/zestawy/nie-uuid",
		"/uruchomienia/" + brak, "/uruchomienia/" + brak + "/plan", "/uruchomienia/" + brak + "/raport",
		"/uruchomienia/" + brak + "/postep", "/nie-ma-takiej",
	} {
		var b api.Blad
		if kod := s.zadanie("GET", sciezka, nil, &b); kod != 404 || b.Kod != "nie_znaleziono" {
			t.Errorf("GET %s: kod %d %q", sciezka, kod, b.Kod)
		}
	}
	var b api.Blad
	if kod := s.zadanie("POST", "/zestawy/"+brak+"/uruchomienia", nil, &b); kod != 404 {
		t.Errorf("uruchomienie bez zestawu: kod %d", kod)
	}
	if kod := s.zadanie("GET", "/uruchomienia?zestaw_id=x", nil, &b); kod != 400 {
		t.Errorf("zły zestaw_id: kod %d", kod)
	}
}

func TestZdrowieIWagi(t *testing.T) {
	s := nowe(t, api.UstawieniaPuli{})
	var zdrowie map[string]string
	if kod := s.zadanie("GET", "/zdrowie", nil, &zdrowie); kod != 200 || zdrowie["baza"] != "ok" || zdrowie["walidator"] != "internal/validator" {
		t.Errorf("zdrowie: kod %d, %v", kod, zdrowie)
	}
	var wagi map[string]float64
	if kod := s.zadanie("GET", "/wagi/domyslne", nil, &wagi); kod != 200 || len(wagi) != 13 || wagi["nie_po_17"] != 20 {
		t.Errorf("wagi: kod %d, %v", kod, wagi)
	}
}
