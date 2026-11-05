// Polecenie solve generuje plan zajęć z danych wejściowych i zapisuje go jako plan.json.
//
//	solve -in data/dane_planu.json -out plan.json -seed 42
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/solver"
)

func main() {
	in := flag.String("in", "data/dane_planu.json", "plik z danymi wejściowymi")
	out := flag.String("out", "plan.json", "plik wynikowy; - oznacza standardowe wyjście")
	seed := flag.Int64("seed", 1, "ziarno losowania")
	flag.Parse()

	if err := uruchom(*in, *out, *seed); err != nil {
		fmt.Fprintln(os.Stderr, "solve:", err)
		os.Exit(1)
	}
}

func uruchom(in, out string, seed int64) error {
	dane, err := model.Wczytaj(in)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	plan, err := solver.Solve(ctx, dane, solver.DomyslneWagi(), solver.ZSeed(seed))
	if err != nil {
		return err
	}
	if out == "-" {
		return model.ZapiszPlan(os.Stdout, &plan)
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if err := model.ZapiszPlan(f, &plan); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
