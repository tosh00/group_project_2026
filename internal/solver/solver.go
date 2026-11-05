// Package solver generuje plan zajęć własnym algorytmem (ADR 0002) w trzech fazach z docs/plan.md:
// przygotowanie domen, konstrukcja zachłanna z naprawą i symulowane wyżarzanie.
//
// Solver zawsze zwraca plan z przypisaniem dla każdych zajęć (ADR 0006). Czy plan spełnia
// reguły twarde, rozstrzyga niezależny pakiet internal/validator.
package solver

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Silnik to wartość meta.silnik w plan.json.
const Silnik = "go-hybrid"

// limitNaprawNaWezel: ile napraw ślepych zaułków (na węzeł) wolno zrobić w fazie 2,
// zanim solver przejdzie w tryb awaryjny.
const limitNaprawNaWezel = 20

// Opcja zmienia ustawienia uruchomienia. Opcje są opcjonalne, więc wywołanie
// Solve(ctx, dane, wagi) z kontraktu działa bez nich.
type Opcja func(*ustawienia)

type ustawienia struct {
	seed  int64
	teraz func() time.Time
}

// ZSeed ustawia ziarno losowania (powtarzalność, meta.seed).
func ZSeed(seed int64) Opcja { return func(u *ustawienia) { u.seed = seed } }

// Solve generuje plan dla danych, które przeszły model.Sprawdz (robi to model.Wczytaj).
// Błąd oznacza tylko niepoprawne dane albo przerwanie przez ctx przed końcem konstrukcji (ADR 0006).
func Solve(ctx context.Context, dane *model.Dane, wagi Wagi, opcje ...Opcja) (model.Plan, error) {
	u := ustawienia{seed: 1, teraz: time.Now}
	for _, o := range opcje {
		o(&u)
	}
	if err := ctx.Err(); err != nil {
		return model.Plan{}, err
	}
	pr := przygotuj(dane)
	rng := rand.New(rand.NewPCG(uint64(u.seed), 0x9e3779b97f4a7c15))
	st, _, err := pr.konstruuj(ctx, rng, &wagi, limitNaprawNaWezel*len(pr.wezly))
	if err != nil {
		return model.Plan{}, err
	}
	// TODO(faza 3): wyżarzanie i pełna funkcja kosztu; na razie meta.koszt to suma kryteriów lokalnych.
	return nowyPlan(u, pr.kosztLokalnyPlanu(st, &wagi), st.przypisania()), nil
}

func (pr *problem) kosztLokalnyPlanu(st *stan, wagi *Wagi) float64 {
	var koszt float64
	for _, w := range pr.wezly {
		koszt += pr.kosztLokalny(w, st.przyp[w.id], wagi)
	}
	return koszt
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
