package validator

import (
	"slices"
	"testing"

	"github.com/tosh00/group_project_2026/internal/model"
)

// maleDane to ręcznie zbudowany zestaw: rocznik R z grupami R-G1 (s1, s2) i R-G2 (s3, s4),
// pięć sal, trzech prowadzących (P1 w Radzie), jedna blokada i jeden dezyderat.
func maleDane() *model.Dane {
	wszystkie := func(z model.Zajecia) model.Zajecia {
		z.Rocznik, z.Czestotliwosc, z.Tryb = "R", model.CoTydzien, model.Stacjonarny
		if z.TydzienOd == 0 {
			z.TydzienOd, z.TydzienDo = 1, 14
		}
		if z.LiczbaSpotkan == 0 {
			z.LiczbaSpotkan = z.TydzienDo - z.TydzienOd + 1
		}
		return z
	}
	z4 := wszystkie(model.Zajecia{ID: "Z4", Grupy: []string{"R-G1"}, LiczbaStudentow: 2, Prowadzacy: []string{"P3"}, SlotyNaSpotkanie: 1, TypySal: []model.TypSali{model.SalaDydaktyczna}})
	z4.Czestotliwosc, z4.LiczbaSpotkan, z4.TygodnieNiedostepne = model.WybraneTygodnie, 9, []int{6}
	z3 := wszystkie(model.Zajecia{ID: "Z3", Grupy: []string{"R-G2"}, LiczbaStudentow: 2, Prowadzacy: []string{"P3"}, SlotyNaSpotkanie: 2, TypySal: []model.TypSali{model.PracowniaKomputerowa}})
	z3.Czestotliwosc, z3.LiczbaSpotkan = model.Co2Tygodnie, 7

	return &model.Dane{
		Sale: []model.Sala{
			{ID: "S1", Typ: model.SalaDydaktyczna, Pojemnosc: 30, DostepnaDoPlanowania: true},
			{ID: "S3", Typ: model.SalaDydaktyczna, Pojemnosc: 30, DostepnaDoPlanowania: true},
			{ID: "S2", Typ: model.PracowniaKomputerowa, Pojemnosc: 20, DostepnaDoPlanowania: true, Wyposazenie: []string{"MAC"}},
			{ID: "AULA", Typ: model.Aula, Pojemnosc: 100, DostepnaDoPlanowania: true},
			{ID: "X", Typ: model.PracowniaSpecjalistyczna, Pojemnosc: 16, PojemnoscSzacowana: true},
		},
		Prowadzacy: []model.Prowadzacy{{ID: "P1", CzlonekRadyWydzialu: true}, {ID: "P2"}, {ID: "P3"}},
		Roczniki:   []model.Rocznik{{ID: "R"}},
		Grupy:      []model.Grupa{{ID: "R-ALL"}, {ID: "R-G1"}, {ID: "R-G2"}},
		StudenciGrupy: []model.StudentGrupa{
			{Student: "s1", Grupa: "R-ALL"}, {Student: "s2", Grupa: "R-ALL"}, {Student: "s3", Grupa: "R-ALL"}, {Student: "s4", Grupa: "R-ALL"},
			{Student: "s1", Grupa: "R-G1"}, {Student: "s2", Grupa: "R-G1"},
			{Student: "s3", Grupa: "R-G2"}, {Student: "s4", Grupa: "R-G2"},
		},
		Zajecia: []model.Zajecia{
			wszystkie(model.Zajecia{ID: "Z1", Forma: model.Wyklad, Grupy: []string{"R-ALL"}, LiczbaStudentow: 4, Prowadzacy: []string{"P1"}, SlotyNaSpotkanie: 2, TypySal: []model.TypSali{model.Aula, model.SalaDydaktyczna}}),
			wszystkie(model.Zajecia{ID: "Z2", Grupy: []string{"R-G1"}, LiczbaStudentow: 2, Prowadzacy: []string{"P2"}, SlotyNaSpotkanie: 2, TypySal: []model.TypSali{model.PracowniaKomputerowa}, WymaganeWyposazenie: []string{"MAC"}}),
			z3,
			z4,
			wszystkie(model.Zajecia{ID: "Z5", Grupy: []string{"R-G2"}, LiczbaStudentow: 2, Prowadzacy: []string{"P2"}, SlotyNaSpotkanie: 2, TypySal: []model.TypSali{model.SalaDydaktyczna}, Powiazanie: "POW-A", TypPowiazania: model.RozneDni}),
			wszystkie(model.Zajecia{ID: "Z6", Grupy: []string{"R-G2"}, LiczbaStudentow: 2, Prowadzacy: []string{"P2"}, SlotyNaSpotkanie: 1, TypySal: []model.TypSali{model.SalaDydaktyczna}, Powiazanie: "POW-A", TypPowiazania: model.RozneDni}),
			wszystkie(model.Zajecia{ID: "Z7", Grupy: []string{"R-ALL"}, LiczbaStudentow: 4, Prowadzacy: []string{"P2"}, SlotyNaSpotkanie: 2, TydzienOd: 1, TydzienDo: 7, TypySal: []model.TypSali{model.SalaDydaktyczna}, Powiazanie: "POW-B", TypPowiazania: model.TenSamTermin}),
			wszystkie(model.Zajecia{ID: "Z8", Grupy: []string{"R-ALL"}, LiczbaStudentow: 4, Prowadzacy: []string{"P3"}, SlotyNaSpotkanie: 2, TydzienOd: 8, TydzienDo: 14, TypySal: []model.TypSali{model.SalaDydaktyczna}, Powiazanie: "POW-B", TypPowiazania: model.TenSamTermin}),
			wszystkie(model.Zajecia{ID: "Z9", Grupy: []string{"R-G1"}, LiczbaStudentow: 2, Prowadzacy: []string{"P1"}, SlotyNaSpotkanie: 1, DozwoloneSale: []string{"X"}}),
		},
		BlokadyStale:      []model.BlokadaStala{{Grupa: "R-G2", Dzien: 4, SlotOd: 1, SlotDo: 4}},
		RadaWydzialu:      []model.TerminRady{{Dzien: 2, SlotOd: 9, SlotDo: 10}},
		DezyderatyTerminy: []model.DezyderatTermin{{Prowadzacy: "P2", Dzien: 1, Slot: 1, Typ: model.Wykluczony}},
	}
}

// poprawnyPlan spełnia wszystkie reguły dla maleDane.
func poprawnyPlan() *model.Plan {
	wszystkie := model.WszystkieTygodnie
	return &model.Plan{Przypisania: []model.Przypisanie{
		{ZajeciaID: "Z1", Dzien: 1, SlotStart: 3, SalaID: "AULA", Tygodnie: wszystkie},
		{ZajeciaID: "Z2", Dzien: 1, SlotStart: 5, SalaID: "S2", Tygodnie: wszystkie},
		{ZajeciaID: "Z3", Dzien: 2, SlotStart: 1, SalaID: "S2", Tygodnie: wszystkie.Nieparzyste()},
		{ZajeciaID: "Z4", Dzien: 3, SlotStart: 1, SalaID: "S1", Tygodnie: model.TygodnieZListy([]int{1, 2, 3, 4, 5, 7, 8, 9, 10})},
		{ZajeciaID: "Z5", Dzien: 3, SlotStart: 3, SalaID: "S1", Tygodnie: wszystkie},
		{ZajeciaID: "Z6", Dzien: 4, SlotStart: 5, SalaID: "S1", Tygodnie: wszystkie},
		{ZajeciaID: "Z7", Dzien: 5, SlotStart: 1, SalaID: "S1", Tygodnie: model.Zakres(1, 7)},
		{ZajeciaID: "Z8", Dzien: 5, SlotStart: 1, SalaID: "S1", Tygodnie: model.Zakres(8, 14)},
		{ZajeciaID: "Z9", Dzien: 2, SlotStart: 5, SalaID: "X", Tygodnie: wszystkie},
	}}
}

func przypisanie(p *model.Plan, id string) *model.Przypisanie {
	for i := range p.Przypisania {
		if p.Przypisania[i].ZajeciaID == id {
			return &p.Przypisania[i]
		}
	}
	panic("brak przypisania " + id)
}

func przesun(id string, dzien, slot int, sala string) func(*model.Dane, *model.Plan) {
	return func(_ *model.Dane, p *model.Plan) {
		pr := przypisanie(p, id)
		pr.Dzien, pr.SlotStart, pr.SalaID = dzien, slot, sala
	}
}

func ustawTygodnie(id string, t model.Tygodnie) func(*model.Dane, *model.Plan) {
	return func(_ *model.Dane, p *model.Plan) { przypisanie(p, id).Tygodnie = t }
}

func reguly(ns []Naruszenie) []int {
	var r []int
	for _, n := range ns {
		if !slices.Contains(r, n.Regula) {
			r = append(r, n.Regula)
		}
	}
	return r
}

func TestPoprawnyPlan(t *testing.T) {
	if ns := Nowy(maleDane()).Sprawdz(poprawnyPlan()); len(ns) != 0 {
		t.Fatalf("poprawny plan ma naruszenia: %v", ns)
	}
}

// Każdy przypadek łamie dokładnie jedną regułę i sprawdza, że walidator zgłasza tylko ją.
func TestRegulyTwarde(t *testing.T) {
	przypadki := []struct {
		nazwa  string
		zmien  func(*model.Dane, *model.Plan)
		regula int
	}{
		{"1: brak przypisania", func(_ *model.Dane, p *model.Plan) { p.Przypisania = p.Przypisania[:len(p.Przypisania)-1] }, RegulaPrzypisanie},
		{"1: podwójne przypisanie", func(_ *model.Dane, p *model.Plan) { p.Przypisania = append(p.Przypisania, p.Przypisania[0]) }, RegulaPrzypisanie},
		{"1: nieznane zajęcia", func(_ *model.Dane, p *model.Plan) {
			p.Przypisania = append(p.Przypisania, model.Przypisanie{ZajeciaID: "Z99", Dzien: 1, SlotStart: 1, SalaID: "S1"})
		}, RegulaPrzypisanie},
		{"1: dzień poza zakresem", przesun("Z1", 6, 3, "AULA"), RegulaPrzypisanie},
		{"1: slot startowy 0", przesun("Z1", 1, 0, "AULA"), RegulaPrzypisanie},
		{"1: spotkanie wychodzi poza slot 12", przesun("Z1", 1, 12, "AULA"), RegulaPrzypisanie},
		{"1: nieznana sala", przesun("Z1", 1, 3, "C999"), RegulaPrzypisanie},

		{"2: co_tydzien bez jednego tygodnia", ustawTygodnie("Z1", model.Zakres(1, 13)), RegulaTygodnie},
		{"2: tydzień spoza zakresu", func(d *model.Dane, _ *model.Plan) { d.Zajecia[0].TydzienOd, d.Zajecia[0].LiczbaSpotkan = 2, 13 }, RegulaTygodnie},
		{"2: co_2_tygodnie z mieszaną parzystością", ustawTygodnie("Z3", model.Zakres(1, 7)), RegulaTygodnie},
		{"2: wybrane_tygodnie z tygodniem niedostępnym", ustawTygodnie("Z4", model.Zakres(1, 9)), RegulaTygodnie},
		{"2: wybrane_tygodnie za mało spotkań", ustawTygodnie("Z4", model.Zakres(1, 5)), RegulaTygodnie},

		{"3: prowadzący w dwóch salach", przesun("Z4", 2, 2, "S1"), RegulaProwadzacy},
		{"4: sala zajęta dwa razy", przesun("Z3", 1, 5, "S2"), RegulaSala},
		{"5: student na dwóch zajęciach", przesun("Z5", 1, 3, "S1"), RegulaStudent},

		{"6: zły typ sali i brak wyposażenia", przesun("Z2", 1, 5, "S1"), RegulaTypSali},
		{"6: sala spoza dozwolone_sale", przesun("Z9", 2, 5, "S1"), RegulaTypSali},
		{"6: sala niedostępna do planowania", func(d *model.Dane, p *model.Plan) {
			d.Zajecia[0].TypySal = append(d.Zajecia[0].TypySal, model.PracowniaSpecjalistyczna)
			przesun("Z1", 1, 3, "X")(d, p)
		}, RegulaTypSali},
		{"7: za mała sala", func(d *model.Dane, _ *model.Plan) { d.Zajecia[1].DodatkowiStudenci = 19 }, RegulaPojemnosc},

		{"8: godzina wykluczona", przesun("Z2", 1, 1, "S2"), RegulaDezyderaty},
		{"8: Rada Wydziału", przesun("Z9", 2, 9, "X"), RegulaDezyderaty},
		{"9: blokada własnej grupy", przesun("Z3", 4, 3, "S2"), RegulaBlokady},
		{"9: blokada grupy ze wspólnymi studentami", przesun("Z1", 4, 1, "AULA"), RegulaBlokady},

		{"10: rozne_dni w tym samym dniu", przesun("Z6", 3, 5, "S1"), RegulaPowiazania},
		{"10: ten_sam_termin w innym slocie", przesun("Z8", 5, 3, "S1"), RegulaPowiazania},
		{"10: ten_sam_termin w innej sali", przesun("Z8", 5, 1, "S3"), RegulaPowiazania},
	}
	for _, p := range przypadki {
		t.Run(p.nazwa, func(t *testing.T) {
			d, plan := maleDane(), poprawnyPlan()
			p.zmien(d, plan)
			ns := Nowy(d).Sprawdz(plan)
			if got := reguly(ns); !slices.Equal(got, []int{p.regula}) {
				t.Errorf("zgłoszone reguły %v, oczekiwano [%d]:\n%v", got, p.regula, ns)
			}
		})
	}
}

func TestNakladanieWymagaWspolnegoTygodnia(t *testing.T) {
	// Z2 w parzyste, Z3 w nieparzyste tygodnie: ta sama sala i godzina, ale bez wspólnego tygodnia.
	d, plan := maleDane(), poprawnyPlan()
	d.Zajecia[1].Czestotliwosc, d.Zajecia[1].LiczbaSpotkan = model.Co2Tygodnie, 7
	ustawTygodnie("Z2", model.WszystkieTygodnie.Parzyste())(d, plan)
	przesun("Z3", 1, 5, "S2")(d, plan)
	if ns := Nowy(d).Sprawdz(plan); len(ns) != 0 {
		t.Errorf("rozłączne tygodnie nie powinny kolidować: %v", ns)
	}
}

func TestPustyPlanNaPrawdziwychDanych(t *testing.T) {
	d, err := model.Wczytaj("../../data/dane_planu.json")
	if err != nil {
		t.Fatal(err)
	}
	ns := Nowy(d).Sprawdz(&model.Plan{})
	if len(ns) != len(d.Zajecia) || !slices.Equal(reguly(ns), []int{RegulaPrzypisanie}) {
		t.Errorf("pusty plan: %d naruszeń reguł %v, oczekiwano %d naruszeń reguły 1", len(ns), reguly(ns), len(d.Zajecia))
	}
}
