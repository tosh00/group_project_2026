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
	"slices"
	"time"

	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/solver"
)

func main() {
	in := flag.String("in", "data/dane_planu.json", "plik z danymi wejściowymi")
	out := flag.String("out", "plan.json", "plik wynikowy; - oznacza standardowe wyjście")
	seed := flag.Int64("seed", 1, "ziarno losowania")
	iteracje := flag.Int("iteracje", solver.DomyslneIteracje, "iteracje wyżarzania na workera (0 = sama konstrukcja)")
	workerzy := flag.Int("workery", solver.DomyslniWorkerzy, "liczba równoległych workerów")
	gadatliwie := flag.Bool("v", false, "wypisz na stderr czas i rozbicie kosztu na kryteria")
	flag.Parse()

	opcje := []solver.Opcja{solver.ZSeed(*seed), solver.ZIteracjami(*iteracje), solver.ZWorkerami(*workerzy)}
	if err := uruchom(*in, *out, *gadatliwie, opcje); err != nil {
		fmt.Fprintln(os.Stderr, "solve:", err)
		os.Exit(1)
	}
}

func uruchom(in, out string, gadatliwie bool, opcje []solver.Opcja) error {
	dane, err := model.Wczytaj(in)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	start := time.Now()
	wagi := solver.DomyslneWagi()
	plan, err := solver.Solve(ctx, dane, wagi, opcje...)
	if err != nil {
		return err
	}
	if gadatliwie {
		wypiszRozbicie(dane, plan, wagi, time.Since(start))
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

func wypiszRozbicie(dane *model.Dane, plan model.Plan, wagi solver.Wagi, czas time.Duration) {
	fmt.Fprintf(os.Stderr, "czas %v, koszt %.1f\n", czas.Round(time.Millisecond), plan.Meta.Koszt)
	rozbicie, err := solver.RozbicieKosztu(dane, plan, wagi)
	if err != nil {
		fmt.Fprintln(os.Stderr, "rozbicie kosztu:", err)
		return
	}
	nazwy := make([]string, 0, len(rozbicie))
	for n := range rozbicie {
		nazwy = append(nazwy, n)
	}
	slices.Sort(nazwy)
	for _, n := range nazwy {
		fmt.Fprintf(os.Stderr, "  %-22s %8.1f\n", n, rozbicie[n])
	}
}
