package solver

import (
	"context"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"

	"github.com/tosh00/group_project_2026/internal/model"
)

func TestOkienkaIPrzerwy(t *testing.T) {
	przypadki := []struct {
		nazwa              string
		maska              uint16
		okienka, przerwy2h int
	}{
		{"pusty dzień", 0, 0, 0},
		{"jeden blok", 0b0000_0011_1100, 0, 0},
		{"jedno okienko", 0b0000_0000_1011, 1, 0},
		{"przerwa 3 sloty", 0b0000_0100_0110, 3, 1},
		{"dwie przerwy", 0b1000_0100_0001, 9, 5},
	}
	for _, p := range przypadki {
		t.Run(p.nazwa, func(t *testing.T) {
			if got := okienka(p.maska); got != p.okienka {
				t.Errorf("okienka = %d, oczekiwano %d", got, p.okienka)
			}
			if got := przerwyPonad(p.maska, 2); got != p.przerwy2h {
				t.Errorf("przerwy > 2 = %d, oczekiwano %d", got, p.przerwy2h)
			}
		})
	}
}

func TestRozpoznajUwagi(t *testing.T) {
	przypadki := []struct {
		tekst string
		chce  uwagiProwadzacego
	}{
		{"Proszę nie planować zajęć po 17:00.", uwagiProwadzacego{niePo17: true}},
		{"Wolę zajęcia w godzinach porannych.", uwagiProwadzacego{rano: true}},
		{"Wolę zajęcia w godzinach popołudniowych.", uwagiProwadzacego{popoludnie: true}},
		{"Proszę o zajęcia w budynku A, jeśli to możliwe.", uwagiProwadzacego{saleA: true}},
		{"Proszę o jeden dzień w tygodniu bez zajęć.", uwagiProwadzacego{dzienWolny: true}},
		{"Proszę o przerwę między zajęciami nie dłuższą niż 2 godziny.", uwagiProwadzacego{przerwaMax2h: true}},
		{"", uwagiProwadzacego{}},
	}
	for _, p := range przypadki {
		if got := rozpoznajUwagi(p.tekst); got != p.chce {
			t.Errorf("%q: %+v, oczekiwano %+v", p.tekst, got, p.chce)
		}
	}
}

// Wszystkie uwagi z danych muszą być rozpoznane: nowa treść bez słowa kluczowego zostałaby po cichu pominięta.
func TestUwagiZDanychRozpoznane(t *testing.T) {
	for _, o := range wczytajDane(t).DezyderatyOgolne {
		if o.InneUwagi != "" && rozpoznajUwagi(o.InneUwagi) == (uwagiProwadzacego{}) {
			t.Errorf("%s: nierozpoznana uwaga %q", o.Prowadzacy, o.InneUwagi)
		}
	}
}

func startoweWyzarzanie(t *testing.T, d *model.Dane, limitNapraw int) (*problem, *wyzarzanie) {
	t.Helper()
	pr := przygotuj(d)
	wagi := DomyslneWagi()
	rng := rand.New(rand.NewPCG(3, 4))
	st, _, err := pr.konstruuj(context.Background(), rng, &wagi, limitNapraw)
	if err != nil {
		t.Fatal(err)
	}
	return pr, noweWyzarzanie(pr, st, &wagi, rng)
}

// Koszt i naruszenia liczone przyrostowo muszą się zgadzać z liczonymi od zera po wielu ruchach.
func TestKosztPrzyrostowyZgodnyOdZera(t *testing.T) {
	pr, s := startoweWyzarzanie(t, wczytajDane(t), limitNaprawNaWezel*173)
	przyjete := 0
	for i := 0; i < 20_000; i++ {
		if !s.losujRuch() {
			continue
		}
		dk, dn := s.wykonaj()
		if s.rng.IntN(2) == 0 {
			s.cofnij(dk)
		} else {
			s.naruszenia += dn
			przyjete++
		}
		if i%997 == 0 {
			odZera := nowyKoszt(pr, s.st, s.wagi).suma
			if math.Abs(odZera-s.k.suma) > 1e-6 {
				t.Fatalf("iteracja %d: koszt przyrostowy %.6f, od zera %.6f", i, s.k.suma, odZera)
			}
			if n := s.policzNaruszenia(); n != s.naruszenia {
				t.Fatalf("iteracja %d: naruszenia przyrostowo %d, od zera %d", i, s.naruszenia, n)
			}
		}
	}
	if przyjete == 0 {
		t.Fatal("żaden ruch nie został przyjęty")
	}
}

func TestRozbicieKosztuSumujeSie(t *testing.T) {
	pr, s := startoweWyzarzanie(t, wczytajDane(t), limitNaprawNaWezel*173)
	var suma float64
	for _, v := range pr.rozbicieKosztu(s.st, s.wagi) {
		suma += v
	}
	if math.Abs(suma-s.k.suma) > 1e-6 {
		t.Errorf("suma kryteriów %.6f, koszt %.6f", suma, s.k.suma)
	}
}

func TestWyzarzanieObnizaKoszt(t *testing.T) {
	d := wczytajDane(t)
	_, s := startoweWyzarzanie(t, d, limitNaprawNaWezel*173)
	start := s.k.suma
	best := s.uruchom(context.Background(), 100_000, nil)
	t.Logf("koszt: konstrukcja %.1f → wyżarzanie %.1f", start, best.koszt)
	if best.naruszenia != 0 || best.koszt >= start {
		t.Errorf("naruszenia %d, koszt %.1f (start %.1f)", best.naruszenia, best.koszt, start)
	}
	st := odbudujStan(s.pr, best.przyp)
	if n := naruszeniaTwarde(d, model.Plan{Przypisania: st.przypisania()}); len(n) > 0 {
		t.Errorf("naruszenia po wyżarzaniu: %v", n)
	}
}

// Tryb awaryjny: konstrukcja bez napraw na ciasnych danych zostawia naruszenia, wyżarzanie je usuwa.
func TestWyzarzanieUsuwaNaruszenia(t *testing.T) {
	d := wczytajDane(t)
	for i := range d.Sale {
		if slices.Contains([]string{"A121", "A122", "A123", "A171", "A172"}, d.Sale[i].ID) {
			d.Sale[i].DostepnaDoPlanowania = false
		}
	}
	_, s := startoweWyzarzanie(t, d, 0)
	if s.naruszenia == 0 {
		t.Skip("konstrukcja bez napraw nie zostawiła naruszeń; test nic nie sprawdza")
	}
	start := s.naruszenia
	best := s.uruchom(context.Background(), 100_000, nil)
	t.Logf("naruszenia: %d → %d", start, best.naruszenia)
	if best.naruszenia != 0 {
		t.Errorf("zostało %d naruszeń", best.naruszenia)
	}
}

func TestSolvePostepIPrzerwanie(t *testing.T) {
	d := wczytajDane(t)
	var mu sync.Mutex
	pracownicy := make(map[int]bool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	plan, err := Solve(ctx, d, DomyslneWagi(), ZWorkerami(2), ZIteracjami(1_000_000), ZPostepem(func(p Postep) {
		mu.Lock()
		defer mu.Unlock()
		pracownicy[p.Pracownik] = true
		if p.Iteracja >= 20_000 {
			cancel() // przerwanie w trakcie wyżarzania: plan zamiast błędu
		}
	}))
	if err != nil {
		t.Fatalf("przerwanie w trakcie wyżarzania nie powinno dać błędu: %v", err)
	}
	if len(pracownicy) != 2 {
		t.Errorf("postęp od %d workerów, oczekiwano 2", len(pracownicy))
	}
	if n := naruszeniaTwarde(d, plan); len(n) > 0 {
		t.Errorf("naruszenia: %v", n)
	}
}

func TestRozbicieKosztuPlanuZSolve(t *testing.T) {
	d := wczytajDane(t)
	plan, err := Solve(context.Background(), d, DomyslneWagi(), ZIteracjami(20_000))
	if err != nil {
		t.Fatal(err)
	}
	rozbicie, err := RozbicieKosztu(d, plan, DomyslneWagi())
	if err != nil {
		t.Fatal(err)
	}
	var suma float64
	for _, v := range rozbicie {
		suma += v
	}
	if math.Abs(suma-plan.Meta.Koszt) > 1e-6 {
		t.Errorf("suma kryteriów %.6f, meta.koszt %.6f", suma, plan.Meta.Koszt)
	}
}
