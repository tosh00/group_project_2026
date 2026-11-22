// Komenda waliduj sprawdza plan.json pod kątem 10 ograniczeń twardych.
//
//	go run ./cmd/waliduj -plan plan.json [-dane data/dane_planu.json] [-json]
//
// Kod wyjścia: 0 = plan poprawny, 1 = są naruszenia, 2 = błąd wczytywania.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tosh00/group_project_2026/internal/model"
	"github.com/tosh00/group_project_2026/internal/validator"
)

func main() {
	sciezkaDanych := flag.String("dane", "data/dane_planu.json", "plik z danymi wejściowymi")
	sciezkaPlanu := flag.String("plan", "", "plik z planem (plan.json)")
	jakoJSON := flag.Bool("json", false, "wypisz naruszenia jako JSON")
	flag.Parse()
	if *sciezkaPlanu == "" {
		fmt.Fprintln(os.Stderr, "brak -plan")
		flag.Usage()
		os.Exit(2)
	}

	dane, err := model.Wczytaj(*sciezkaDanych)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	plan, err := model.WczytajPlan(*sciezkaPlanu)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	naruszenia := validator.Nowy(dane).Sprawdz(plan)
	if *jakoJSON {
		if naruszenia == nil {
			naruszenia = []validator.Naruszenie{}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(naruszenia); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		for _, n := range naruszenia {
			fmt.Println(n)
		}
		if len(naruszenia) == 0 {
			fmt.Println("plan poprawny: 0 naruszeń reguł twardych")
		} else {
			fmt.Printf("%d naruszeń reguł twardych\n", len(naruszenia))
		}
	}
	if len(naruszenia) > 0 {
		os.Exit(1)
	}
}
