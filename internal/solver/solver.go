// Package solver generuje plan zajęć własnym algorytmem (ADR 0002) w trzech fazach z docs/plan.md:
// przygotowanie domen, konstrukcja zachłanna z naprawą i symulowane wyżarzanie.
//
// Solver zawsze zwraca plan z przypisaniem dla każdych zajęć (ADR 0006). Czy plan spełnia
// reguły twarde, rozstrzyga niezależny pakiet internal/validator.
package solver

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Silnik to wartość meta.silnik w plan.json.
const Silnik = "go-hybrid"

const (
	// limitNaprawNaWezel: ile napraw ślepych zaułków (na węzeł) wolno zrobić w fazie 2,
	// zanim solver przejdzie w tryb awaryjny.
	limitNaprawNaWezel = 20

	DomyslneIteracje = 1_000_000 // iteracje wyżarzania na workera (ok. 1,5 s na 4 rdzeniach)
	DomyslniWorkerzy = 4         // stała liczba, żeby wynik zależał tylko od seeda, nie od maszyny
)

// Postep to stan jednego workera w trakcie wyżarzania (np. do SSE).
type Postep struct {
	Pracownik  int     // 0 … liczba workerów − 1
	Iteracja   int     // wykonane iteracje
	Iteracji   int     // wszystkie iteracje
	Koszt      float64 // bieżący koszt miękki (na końcu: najlepszy)
	Naruszenia int     // bieżąca liczba naruszeń twardych (na końcu: najlepsza)
}

// Opcja zmienia ustawienia uruchomienia. Opcje są opcjonalne, więc wywołanie
// Solve(ctx, dane, wagi) z kontraktu działa bez nich.
type Opcja func(*ustawienia)

type ustawienia struct {
	seed     int64
	iteracje int
	workerzy int
	postep   func(Postep)
	teraz    func() time.Time
}

// ZSeed ustawia ziarno losowania (powtarzalność, meta.seed).
func ZSeed(seed int64) Opcja { return func(u *ustawienia) { u.seed = seed } }

// ZIteracjami ustawia liczbę iteracji wyżarzania na workera (0 = sama konstrukcja, bez fazy 3).
func ZIteracjami(n int) Opcja { return func(u *ustawienia) { u.iteracje = max(0, n) } }

// ZWorkerami ustawia liczbę równoległych workerów (każdy z innym seedem pochodnym).
func ZWorkerami(n int) Opcja { return func(u *ustawienia) { u.workerzy = max(1, n) } }

// ZPostepem ustawia funkcję wołaną co jakiś czas przez każdego workera.
// Wywołania przychodzą z wielu goroutines naraz, więc funkcja musi być bezpieczna współbieżnie.
func ZPostepem(f func(Postep)) Opcja { return func(u *ustawienia) { u.postep = f } }

// Solve generuje plan dla danych, które przeszły model.Sprawdz (robi to model.Wczytaj).
// Błąd oznacza tylko przerwanie przez ctx przed końcem konstrukcji (ADR 0006); przerwanie
// w trakcie wyżarzania zwraca najlepszy plan znaleziony do tej pory.
func Solve(ctx context.Context, dane *model.Dane, wagi Wagi, opcje ...Opcja) (model.Plan, error) {
	u := ustawienia{seed: 1, iteracje: DomyslneIteracje, workerzy: DomyslniWorkerzy, teraz: time.Now}
	for _, o := range opcje {
		o(&u)
	}
	if err := ctx.Err(); err != nil {
		return model.Plan{}, err
	}
	pr := przygotuj(dane)

	wyniki := make([]wynikPracownika, u.workerzy)
	var wg sync.WaitGroup
	for i := range wyniki {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wyniki[i] = pr.pracuj(ctx, i, &u, &wagi)
		}()
	}
	wg.Wait()

	// Najlepszy wynik: najpierw naruszenia, potem koszt; remis rozstrzyga numer workera (powtarzalność).
	najlepszy := -1
	for i, w := range wyniki {
		if w.err != nil {
			continue
		}
		if najlepszy < 0 || lepszy(w.koszt, w.naruszenia, wyniki[najlepszy].koszt, wyniki[najlepszy].naruszenia) {
			najlepszy = i
		}
	}
	if najlepszy < 0 {
		return model.Plan{}, wyniki[0].err
	}
	st := wyniki[najlepszy].st
	return nowyPlan(u, nowyKoszt(pr, st, &wagi).suma, st.przypisania()), nil
}

type wynikPracownika struct {
	st         *stan
	koszt      float64
	naruszenia int
	err        error
}

// pracuj: konstrukcja (faza 2) i wyżarzanie (faza 3) z seedem pochodnym od numeru workera.
func (pr *problem) pracuj(ctx context.Context, nr int, u *ustawienia, wagi *Wagi) wynikPracownika {
	rng := rand.New(rand.NewPCG(uint64(u.seed), uint64(nr)+0x9e3779b97f4a7c15))
	st, _, err := pr.konstruuj(ctx, rng, wagi, limitNaprawNaWezel*len(pr.wezly))
	if err != nil {
		return wynikPracownika{err: err}
	}
	s := noweWyzarzanie(pr, st, wagi, rng)
	var postep func(int, float64, int)
	if u.postep != nil {
		postep = func(it int, koszt float64, nar int) {
			u.postep(Postep{Pracownik: nr, Iteracja: it, Iteracji: u.iteracje, Koszt: koszt, Naruszenia: nar})
		}
	}
	best := s.uruchom(ctx, u.iteracje, postep)
	st = odbudujStan(pr, best.przyp) // stan z najlepszymi przypisaniami i zgodną zajętością sal
	return wynikPracownika{st: st, koszt: best.koszt, naruszenia: best.naruszenia}
}

func odbudujStan(pr *problem, przyp []kandydat) *stan {
	st := nowyStan(pr)
	for w, kd := range przyp {
		st.przypisz(w, kd)
	}
	return st
}

// nowyPlan buduje plan.json v1 z przypisań.
func nowyPlan(u ustawienia, koszt float64, przypisania []model.Przypisanie) model.Plan {
	return model.Plan{
		Meta: model.MetaPlanu{
			WersjaFormatu: model.WersjaFormatuPlanu,
			Silnik:        Silnik,
			Seed:          u.seed,
			Koszt:         koszt,
			Utworzono:     u.teraz().UTC().Truncate(time.Second),
		},
		Przypisania: przypisania,
	}
}

// RozbicieKosztu liczy koszt miękki planu z podziałem na kryteria (klucze jak tagi JSON w Wagi).
// Plan musi używać sal i tygodni z domen solvera (tak jest dla planów z Solve).
func RozbicieKosztu(dane *model.Dane, plan model.Plan, wagi Wagi) (map[string]float64, error) {
	pr := przygotuj(dane)
	st, err := pr.stanZPlanu(plan)
	if err != nil {
		return nil, err
	}
	return pr.rozbicieKosztu(st, &wagi), nil
}

// stanZPlanu odtwarza stan solvera z plan.json.
func (pr *problem) stanZPlanu(plan model.Plan) (*stan, error) {
	numerSali := make(map[string]int, len(pr.sale))
	for i, s := range pr.sale {
		numerSali[s.ID] = i
	}
	poZajeciach := make(map[string]model.Przypisanie, len(plan.Przypisania))
	for _, p := range plan.Przypisania {
		poZajeciach[p.ZajeciaID] = p
	}
	st := nowyStan(pr)
	for _, w := range pr.wezly {
		p, ok := poZajeciach[w.zajecia[0].ID]
		if !ok {
			return nil, fmt.Errorf("solver: brak przypisania dla %s", w.zajecia[0].ID)
		}
		sala, ok := numerSali[p.SalaID]
		if !ok {
			return nil, fmt.Errorf("solver: %s w sali %s spoza domen solvera", p.ZajeciaID, p.SalaID)
		}
		wariant := slices.IndexFunc(w.warianty, func(wr wariant) bool {
			for i, z := range w.zajecia {
				if q, ok := poZajeciach[z.ID]; !ok || q.Tygodnie != wr.tygodnie[i] {
					return false
				}
			}
			return true
		})
		if wariant < 0 {
			return nil, fmt.Errorf("solver: tygodnie %s spoza wariantów solvera", w.nazwa())
		}
		st.przypisz(w.id, kandydat{dzien: p.Dzien, start: p.SlotStart, sala: sala, wariant: wariant})
	}
	return st, nil
}
