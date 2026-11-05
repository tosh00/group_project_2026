// Package solver generuje plan zajęć własnym algorytmem (ADR 0002) w trzech fazach z docs/plan.md:
// przygotowanie domen, konstrukcja zachłanna z naprawą i symulowane wyżarzanie.
//
// Solver zawsze zwraca plan z przypisaniem dla każdych zajęć (ADR 0006). Czy plan spełnia
// reguły twarde, rozstrzyga niezależny pakiet internal/validator.
package solver

import (
	"context"
	"errors"
	"time"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Silnik to wartość meta.silnik w plan.json.
const Silnik = "go-hybrid"

// ErrNieZaimplementowane zwraca Solve, dopóki faza 2 nie jest gotowa.
var ErrNieZaimplementowane = errors.New("solver: faza 2 jeszcze nie jest zaimplementowana")

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
	_ = przygotuj(dane)
	_ = wagi
	return model.Plan{}, ErrNieZaimplementowane
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
