package solver

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/tosh00/group_project_2026/internal/model"
)

// naruszeniaTwarde to uproszczony, niezależny od solvera sprawdzacz reguł 1–10, liczony wprost
// z tabel danych. Pełny dowód poprawności daje internal/validator; ten wystarcza do testów solvera.
func naruszeniaTwarde(d *model.Dane, plan model.Plan) []string {
	ix := model.NowyIndeks(d)
	var out []string
	zle := func(format string, a ...any) { out = append(out, fmt.Sprintf(format, a...)) }

	type sp struct {
		p model.Przypisanie
		z *model.Zajecia
		s *model.Sala
	}
	var sps []sp
	poId := make(map[string]sp)
	for _, p := range plan.Przypisania {
		z, s := ix.Zajecia[p.ZajeciaID], ix.Sale[p.SalaID]
		if z == nil || s == nil || p.Dzien < 1 || p.Dzien > 5 || p.SlotStart < 1 || p.SlotStart+z.SlotyNaSpotkanie-1 > 12 {
			zle("r1 %s", p.ZajeciaID)
			continue
		}
		if _, juz := poId[z.ID]; juz {
			zle("r1 %s dwa razy", z.ID)
		}
		x := sp{p, z, s}
		sps = append(sps, x)
		poId[z.ID] = x
	}
	if len(poId) != len(d.Zajecia) {
		zle("r1: %d przypisań, %d zajęć", len(poId), len(d.Zajecia))
	}
	koniec := func(x sp) int { return x.p.SlotStart + x.z.SlotyNaSpotkanie - 1 }
	for _, x := range sps {
		dozw := x.z.DozwoloneTygodnie()
		ok := x.p.Tygodnie.Podzbior(dozw) && x.p.Tygodnie.Liczba() == x.z.LiczbaSpotkan
		switch x.z.Czestotliwosc {
		case model.CoTydzien:
			ok = ok && x.p.Tygodnie == dozw
		case model.Co2Tygodnie:
			ok = ok && (x.p.Tygodnie == dozw.Parzyste() || x.p.Tygodnie == dozw.Nieparzyste())
		}
		if !ok {
			zle("r2 %s", x.z.ID)
		}
		if !salaPasuje(x.z, x.s) {
			zle("r6/7 %s w %s", x.z.ID, x.s.ID)
		}
		for s := x.p.SlotStart; s <= koniec(x); s++ {
			for _, p := range x.z.Prowadzacy {
				for _, t := range d.DezyderatyTerminy {
					if t.Prowadzacy == p && t.Dzien == x.p.Dzien && t.Slot == s {
						zle("r8 %s %s", x.z.ID, p)
					}
				}
				for _, r := range d.RadaWydzialu {
					if ix.Prowadzacy[p].CzlonekRadyWydzialu && r.Dzien == x.p.Dzien && r.SlotOd <= s && s <= r.SlotDo {
						zle("r8 rada %s %s", x.z.ID, p)
					}
				}
			}
			for _, b := range d.BlokadyStale {
				for _, g := range x.z.Grupy {
					if b.Dzien == x.p.Dzien && b.SlotOd <= s && s <= b.SlotDo && ix.GrupyKoliduja(g, b.Grupa) {
						zle("r9 %s", x.z.ID)
					}
				}
			}
		}
	}
	for i, a := range sps {
		for _, b := range sps[i+1:] {
			if a.p.Dzien != b.p.Dzien || a.p.SlotStart > koniec(b) || b.p.SlotStart > koniec(a) || !a.p.Tygodnie.Nakladaja(b.p.Tygodnie) {
				continue
			}
			if model.WspolnyProwadzacy(a.z, b.z) {
				zle("r3 %s %s", a.z.ID, b.z.ID)
			}
			if a.s.ID == b.s.ID {
				zle("r4 %s %s", a.z.ID, b.z.ID)
			}
			if ix.ZajeciaKolidujaStudenci(a.z, b.z) {
				zle("r5 %s %s", a.z.ID, b.z.ID)
			}
		}
	}
	for nazwa, zs := range ix.Powiazania {
		for i, z1 := range zs {
			for _, z2 := range zs[i+1:] {
				a, b := poId[z1.ID], poId[z2.ID]
				switch z1.TypPowiazania {
				case model.RozneDni:
					if a.p.Dzien == b.p.Dzien {
						zle("r10 %s %s %s", nazwa, z1.ID, z2.ID)
					}
				case model.TenSamTermin:
					if a.p.Dzien != b.p.Dzien || a.p.SlotStart != b.p.SlotStart || a.p.SalaID != b.p.SalaID {
						zle("r10 %s %s %s", nazwa, z1.ID, z2.ID)
					}
				}
			}
		}
	}
	return out
}

func TestSolvePrawdziweDane(t *testing.T) {
	d := wczytajDane(t)
	for _, seed := range []int64{1, 2, 42, 2026} {
		t.Run(fmt.Sprint("seed ", seed), func(t *testing.T) {
			plan, err := Solve(context.Background(), d, DomyslneWagi(), ZSeed(seed))
			if err != nil {
				t.Fatal(err)
			}
			if plan.Meta.WersjaFormatu != model.WersjaFormatuPlanu || plan.Meta.Seed != seed || plan.Meta.Silnik != Silnik {
				t.Errorf("zła meta: %+v", plan.Meta)
			}
			if n := naruszeniaTwarde(d, plan); len(n) > 0 {
				t.Errorf("%d naruszeń, np. %v", len(n), n[:min(len(n), 5)])
			}
		})
	}
}

func TestKonstrukcjaBezTrybuAwaryjnego(t *testing.T) {
	pr := przygotuj(wczytajDane(t))
	wagi := DomyslneWagi()
	_, stat, err := pr.konstruuj(context.Background(), rand.New(rand.NewPCG(1, 2)), &wagi, limitNaprawNaWezel*len(pr.wezly))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("naprawy: %d", stat.naprawy)
	if len(stat.awaryjne) > 0 {
		t.Errorf("tryb awaryjny dla %d węzłów", len(stat.awaryjne))
	}
}

func TestSolvePowtarzalny(t *testing.T) {
	d := wczytajDane(t)
	a, _ := Solve(context.Background(), d, DomyslneWagi(), ZSeed(7))
	b, _ := Solve(context.Background(), d, DomyslneWagi(), ZSeed(7))
	if !slices.Equal(a.Przypisania, b.Przypisania) {
		t.Error("ten sam seed dał różne plany")
	}
}

func TestSolvePrzerwany(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Solve(ctx, wczytajDane(t), DomyslneWagi()); !errors.Is(err, context.Canceled) {
		t.Errorf("oczekiwano context.Canceled, jest %v", err)
	}
}

// Tryb awaryjny (ADR 0006): przy sprzecznych danych solver i tak zwraca przypisanie dla każdych zajęć.
func TestTrybAwaryjny(t *testing.T) {
	przypadki := []struct {
		nazwa       string
		zepsuj      func(d *model.Dane)
		minNaruszen int
	}{
		{
			// Wszystkie laboratoria jednego rocznika w jednej sali: za mało terminów.
			"za mało terminów w sali",
			func(d *model.Dane) {
				for i := range d.Zajecia {
					if d.Zajecia[i].Forma == model.LabInformatyczne {
						d.Zajecia[i].DozwoloneSale = []string{"A234"}
						d.Zajecia[i].WymaganeWyposazenie = nil
						d.Zajecia[i].LiczbaStudentow = min(d.Zajecia[i].LiczbaStudentow, 20)
						d.Zajecia[i].DodatkowiStudenci = 0
					}
				}
			},
			1,
		},
		{
			// Żadna sala nie mieści zajęć: domena sal pusta, rozluźniona awaryjnie.
			"brak sali o takiej pojemności",
			func(d *model.Dane) {
				for i := range d.Zajecia {
					if d.Zajecia[i].ID == "Z001" {
						d.Zajecia[i].LiczbaStudentow = 1000
					}
				}
			},
			1,
		},
	}
	for _, p := range przypadki {
		t.Run(p.nazwa, func(t *testing.T) {
			d := wczytajDane(t)
			p.zepsuj(d)
			plan, err := Solve(context.Background(), d, DomyslneWagi())
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Przypisania) != len(d.Zajecia) {
				t.Fatalf("%d przypisań, oczekiwano %d", len(plan.Przypisania), len(d.Zajecia))
			}
			if n := naruszeniaTwarde(d, plan); len(n) < p.minNaruszen {
				t.Errorf("%d naruszeń, oczekiwano co najmniej %d", len(n), p.minNaruszen)
			}
		})
	}
}

// Mniej pracowni komputerowych: problem jest ciaśniejszy, ale nadal rozwiązywalny bez trybu awaryjnego.
func TestKonstrukcjaCiasneSale(t *testing.T) {
	d := wczytajDane(t)
	for i := range d.Sale {
		if slices.Contains([]string{"A121", "A122", "A123", "A171", "A172"}, d.Sale[i].ID) {
			d.Sale[i].DostepnaDoPlanowania = false
		}
	}
	pr := przygotuj(d)
	wagi := DomyslneWagi()
	st, stat, err := pr.konstruuj(context.Background(), rand.New(rand.NewPCG(1, 2)), &wagi, limitNaprawNaWezel*len(pr.wezly))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("naprawy: %d", stat.naprawy)
	if n := naruszeniaTwarde(d, model.Plan{Przypisania: st.przypisania()}); len(n) > 0 || len(stat.awaryjne) > 0 {
		t.Errorf("awaryjne %d, naruszenia %v", len(stat.awaryjne), n)
	}
}

// maleDane: jedna sala, dwa całodniowe zajęcia A i B. A może być w dniu 1 lub 2, B tylko w dniu 1
// (prowadzący B ma wykluczone dni 2–5).
func maleDane() *model.Dane {
	d := &model.Dane{
		Sale:       []model.Sala{{ID: "S1", Typ: model.SalaDydaktyczna, Pojemnosc: 30, DostepnaDoPlanowania: true}},
		Prowadzacy: []model.Prowadzacy{{ID: "PA"}, {ID: "PB"}},
		Grupy:      []model.Grupa{{ID: "GA"}, {ID: "GB"}},
	}
	for _, id := range []string{"A", "B"} {
		d.Zajecia = append(d.Zajecia, model.Zajecia{
			ID: id, Grupy: []string{"G" + id}, LiczbaStudentow: 10, Prowadzacy: []string{"P" + id},
			LiczbaSpotkan: 14, SlotyNaSpotkanie: 12, Czestotliwosc: model.CoTydzien, TydzienOd: 1, TydzienDo: 14,
			TypySal: []model.TypSali{model.SalaDydaktyczna},
		})
	}
	for dzien := 3; dzien <= 5; dzien++ {
		d.DezyderatyTerminy = append(d.DezyderatyTerminy, model.DezyderatTermin{Prowadzacy: "PA", Dzien: dzien, Slot: 1, Typ: model.Wykluczony})
	}
	for dzien := 2; dzien <= 5; dzien++ {
		d.DezyderatyTerminy = append(d.DezyderatyTerminy, model.DezyderatTermin{Prowadzacy: "PB", Dzien: dzien, Slot: 1, Typ: model.Wykluczony})
	}
	return d
}

// napraw: węzeł bez wolnego kandydata trafia na miejsce, a zajęcia w konflikcie z nim wracają do kolejki.
func TestNaprawWyrzucaKonflikty(t *testing.T) {
	pr := przygotuj(maleDane())
	wagi := DomyslneWagi()
	k := &konstrukcja{
		pr: pr, st: nowyStan(pr), rng: rand.New(rand.NewPCG(1, 2)), wagi: &wagi,
		ostatnioNaprawa: slices.Repeat([]int{-1}, len(pr.wezly)),
	}
	a, b := pr.wezelZajec["A"], pr.wezelZajec["B"]
	k.st.przypisz(a, kandydat{dzien: 1, start: 1, sala: 0, wariant: 0})
	if k.st.maWolnegoKandydata(b) {
		t.Fatal("B nie powinno mieć wolnego kandydata, gdy A zajmuje dzień 1")
	}
	k.napraw(b)
	if !k.st.ok[b] || k.st.przyp[b].dzien != 1 {
		t.Fatalf("B powinno trafić na dzień 1: %+v", k.st.przyp[b])
	}
	if k.st.ok[a] {
		t.Error("A powinno zostać wyrzucone")
	}
	if !k.st.maWolnegoKandydata(a) {
		t.Error("A powinno mieć wolny dzień 2")
	}
}

// Ta sama sytuacja od zera: konstrukcja kończy z planem bez naruszeń.
func TestKonstrukcjaMaleDane(t *testing.T) {
	d := maleDane()
	plan, err := Solve(context.Background(), d, DomyslneWagi())
	if err != nil {
		t.Fatal(err)
	}
	if n := naruszeniaTwarde(d, plan); len(n) > 0 {
		t.Errorf("naruszenia: %v", n)
	}
}
