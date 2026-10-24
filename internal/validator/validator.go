// Package validator sprawdza gotowy plan pod kątem 10 ograniczeń twardych z data/README.md (sekcja 5).
//
// To niezależny dowód poprawności planu: pakiet nie importuje niczego z internal/solver
// i nie zakłada niczego o tym, jak plan powstał.
package validator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Numery reguł twardych z data/README.md, sekcja 5.
const (
	RegulaPrzypisanie = 1  // jeden dzień, slot startowy i sala; sloty w zakresie 1–12
	RegulaTygodnie    = 2  // liczba spotkań, zakres tygodni, częstotliwość, tygodnie_niedostepne
	RegulaProwadzacy  = 3  // prowadzący bez nakładających się zajęć
	RegulaSala        = 4  // sala bez nakładających się zajęć
	RegulaStudent     = 5  // student bez nakładających się zajęć
	RegulaTypSali     = 6  // typ / dozwolone_sale, wyposażenie, dostępność
	RegulaPojemnosc   = 7  // pojemność ≥ liczba_studentow + dodatkowi_studenci
	RegulaDezyderaty  = 8  // godziny wykluczone prowadzącego i Rada Wydziału
	RegulaBlokady     = 9  // blokady_stale grup
	RegulaPowiazania  = 10 // rozne_dni, ten_sam_termin
)

// Naruszenie to jedno złamanie reguły twardej.
type Naruszenie struct {
	Regula  int      `json:"regula"`
	Zajecia []string `json:"zajecia"` // zajęcia, których dotyczy (posortowane)
	Opis    string   `json:"opis"`
}

func (n Naruszenie) String() string {
	return fmt.Sprintf("reguła %d [%s]: %s", n.Regula, strings.Join(n.Zajecia, ", "), n.Opis)
}

// Walidator sprawdza plany dla jednego zestawu danych. Można go używać wielokrotnie
// i współbieżnie (np. przy sprawdzaniu ruchów w UI), bo nie zmienia swojego stanu.
type Walidator struct {
	dane *model.Dane
	ix   *model.Indeks

	// wykluczone: prowadzący → [dzień][slot] zablokowany (dezyderaty i Rada Wydziału).
	wykluczone map[string]*[model.LiczbaDni + 1][model.LiczbaSlotow + 1]bool
}

// Nowy przygotowuje walidator. Dane powinny przejść model.Sprawdz (robi to model.Wczytaj).
func Nowy(d *model.Dane) *Walidator {
	w := &Walidator{
		dane:       d,
		ix:         model.NowyIndeks(d),
		wykluczone: make(map[string]*[model.LiczbaDni + 1][model.LiczbaSlotow + 1]bool),
	}
	zablokuj := func(p string, dzien, slot int) {
		if dzien < 1 || dzien > model.LiczbaDni || slot < 1 || slot > model.LiczbaSlotow {
			return
		}
		if w.wykluczone[p] == nil {
			w.wykluczone[p] = new([model.LiczbaDni + 1][model.LiczbaSlotow + 1]bool)
		}
		w.wykluczone[p][dzien][slot] = true
	}
	for _, t := range d.DezyderatyTerminy {
		zablokuj(t.Prowadzacy, t.Dzien, t.Slot)
	}
	for _, p := range d.Prowadzacy {
		if !p.CzlonekRadyWydzialu {
			continue
		}
		for _, r := range d.RadaWydzialu {
			for s := r.SlotOd; s <= r.SlotDo; s++ {
				zablokuj(p.ID, r.Dzien, s)
			}
		}
	}
	return w
}

// spotkanie to przypisanie połączone z danymi zajęć i sali, po sprawdzeniu reguły 1.
type spotkanie struct {
	p    model.Przypisanie
	z    *model.Zajecia
	sala *model.Sala
}

func (s *spotkanie) slotDo() int { return s.p.SlotStart + s.z.SlotyNaSpotkanie - 1 }

// nakladaja: ten sam dzień, wspólny slot i wspólny tydzień (data/README.md, sekcja 5).
func nakladaja(a, b *spotkanie) bool {
	return a.p.Dzien == b.p.Dzien &&
		a.p.SlotStart <= b.slotDo() && b.p.SlotStart <= a.slotDo() &&
		a.p.Tygodnie.Nakladaja(b.p.Tygodnie)
}

// Sprawdz zwraca wszystkie naruszenia reguł twardych, posortowane po regule i zajęciach.
// Pusty wynik oznacza plan poprawny.
func (w *Walidator) Sprawdz(plan *model.Plan) []Naruszenie {
	var wynik []Naruszenie
	dodaj := func(regula int, opis string, zajecia ...string) {
		zajecia = slices.Clone(zajecia)
		slices.Sort(zajecia)
		wynik = append(wynik, Naruszenie{Regula: regula, Zajecia: zajecia, Opis: opis})
	}

	spotkania := w.regula1(plan, dodaj)
	for _, s := range spotkania {
		w.regula2(s, dodaj)
		w.regula6i7(s, dodaj)
		w.regula8(s, dodaj)
		w.regula9(s, dodaj)
	}
	for i, a := range spotkania {
		for _, b := range spotkania[i+1:] {
			w.reguly345(a, b, dodaj)
		}
	}
	w.regula10(spotkania, dodaj)

	slices.SortStableFunc(wynik, func(a, b Naruszenie) int {
		if a.Regula != b.Regula {
			return a.Regula - b.Regula
		}
		return slices.Compare(a.Zajecia, b.Zajecia)
	})
	return wynik
}

type dodajFunc func(regula int, opis string, zajecia ...string)

// regula1 sprawdza, że każde zajęcia mają dokładnie jedno przypisanie z poprawnym dniem, slotami i salą.
// Zwraca tylko przypisania, które przeszły regułę 1. Pozostałe reguły sprawdzamy już tylko dla nich.
func (w *Walidator) regula1(plan *model.Plan, dodaj dodajFunc) []*spotkanie {
	widziane := make(map[string]bool, len(plan.Przypisania))
	var spotkania []*spotkanie
	for _, p := range plan.Przypisania {
		z := w.ix.Zajecia[p.ZajeciaID]
		if z == nil {
			dodaj(RegulaPrzypisanie, fmt.Sprintf("nieznane zajęcia %q", p.ZajeciaID), p.ZajeciaID)
			continue
		}
		if widziane[p.ZajeciaID] {
			dodaj(RegulaPrzypisanie, "zajęcia przypisane więcej niż raz", p.ZajeciaID)
			continue
		}
		widziane[p.ZajeciaID] = true

		ok := true
		if p.Dzien < 1 || p.Dzien > model.LiczbaDni {
			dodaj(RegulaPrzypisanie, fmt.Sprintf("dzień %d poza zakresem 1–%d", p.Dzien, model.LiczbaDni), z.ID)
			ok = false
		}
		if koniec := p.SlotStart + z.SlotyNaSpotkanie - 1; p.SlotStart < 1 || koniec > model.LiczbaSlotow {
			dodaj(RegulaPrzypisanie, fmt.Sprintf("sloty %d–%d poza zakresem 1–%d", p.SlotStart, koniec, model.LiczbaSlotow), z.ID)
			ok = false
		}
		sala := w.ix.Sale[p.SalaID]
		if sala == nil {
			dodaj(RegulaPrzypisanie, fmt.Sprintf("nieznana sala %q", p.SalaID), z.ID)
			ok = false
		}
		if ok {
			spotkania = append(spotkania, &spotkanie{p: p, z: z, sala: sala})
		}
	}
	for _, z := range w.dane.Zajecia {
		if !widziane[z.ID] {
			dodaj(RegulaPrzypisanie, "zajęcia bez przypisania", z.ID)
		}
	}
	return spotkania
}

// regula2: tygodnie zgodne z liczbą spotkań, zakresem, częstotliwością i tygodniami niedostępnymi.
func (w *Walidator) regula2(s *spotkanie, dodaj dodajFunc) {
	z, tyg := s.z, s.p.Tygodnie
	dozwolone := z.DozwoloneTygodnie()
	if !tyg.Podzbior(dozwolone) {
		dodaj(RegulaTygodnie, fmt.Sprintf("tygodnie %v poza dozwolonymi (zakres %d–%d, niedostępne %v)",
			tyg.Bez(dozwolone).Lista(), z.TydzienOd, z.TydzienDo, z.TygodnieNiedostepne), z.ID)
		return
	}
	if tyg.Liczba() != z.LiczbaSpotkan {
		dodaj(RegulaTygodnie, fmt.Sprintf("%d spotkań, oczekiwano %d", tyg.Liczba(), z.LiczbaSpotkan), z.ID)
		return
	}
	switch z.Czestotliwosc {
	case model.CoTydzien:
		if tyg != dozwolone {
			dodaj(RegulaTygodnie, fmt.Sprintf("co_tydzien, ale brak tygodni %v", dozwolone.Bez(tyg).Lista()), z.ID)
		}
	case model.Co2Tygodnie:
		if tyg != dozwolone.Parzyste() && tyg != dozwolone.Nieparzyste() {
			dodaj(RegulaTygodnie, fmt.Sprintf("co_2_tygodnie, ale tygodnie %v nie są wszystkimi parzystymi ani nieparzystymi", tyg.Lista()), z.ID)
		}
	case model.WybraneTygodnie:
		// Dowolne LiczbaSpotkan tygodni z dozwolonych (ADR 0004): sprawdzone wyżej.
	}
}

// regula6i7: dozwolona sala z wyposażeniem i wystarczającą pojemnością.
func (w *Walidator) regula6i7(s *spotkanie, dodaj dodajFunc) {
	z, sala := s.z, s.sala
	if len(z.DozwoloneSale) > 0 {
		// Sale z dozwolone_sale wolno użyć nawet wtedy, gdy nie są dostępne do planowania (data/README.md).
		if !slices.Contains(z.DozwoloneSale, sala.ID) {
			dodaj(RegulaTypSali, fmt.Sprintf("sala %s spoza dozwolonych %v", sala.ID, z.DozwoloneSale), z.ID)
		}
	} else {
		if !slices.Contains(z.TypySal, sala.Typ) {
			dodaj(RegulaTypSali, fmt.Sprintf("sala %s typu %s, dozwolone typy %v", sala.ID, sala.Typ, z.TypySal), z.ID)
		}
		if !sala.DostepnaDoPlanowania {
			dodaj(RegulaTypSali, fmt.Sprintf("sala %s niedostępna do planowania", sala.ID), z.ID)
		}
	}
	for _, wyp := range z.WymaganeWyposazenie {
		if !slices.Contains(sala.Wyposazenie, wyp) {
			dodaj(RegulaTypSali, fmt.Sprintf("sala %s bez wymaganego wyposażenia %s", sala.ID, wyp), z.ID)
		}
	}
	// Bez marginesu, także dla pojemności szacowanych (ADR 0005).
	if sala.Pojemnosc < z.LiczbaOsob() {
		dodaj(RegulaPojemnosc, fmt.Sprintf("sala %s ma %d miejsc, potrzeba %d", sala.ID, sala.Pojemnosc, z.LiczbaOsob()), z.ID)
	}
}

// regula8: prowadzący nie ma zajęć w godzinach wykluczonych ani w czasie Rady Wydziału.
func (w *Walidator) regula8(s *spotkanie, dodaj dodajFunc) {
	for _, p := range s.z.Prowadzacy {
		wyk := w.wykluczone[p]
		if wyk == nil {
			continue
		}
		var sloty []int
		for slot := s.p.SlotStart; slot <= s.slotDo(); slot++ {
			if wyk[s.p.Dzien][slot] {
				sloty = append(sloty, slot)
			}
		}
		if len(sloty) > 0 {
			dodaj(RegulaDezyderaty, fmt.Sprintf("prowadzący %s ma wykluczone sloty %v w dniu %d", p, sloty, s.p.Dzien), s.z.ID)
		}
	}
}

// regula9: zajęcia nie nakładają się na blokady stałe grup, z którymi mają wspólnych studentów.
// Blokady obowiązują w każdym tygodniu, więc wystarczy dzień i sloty.
func (w *Walidator) regula9(s *spotkanie, dodaj dodajFunc) {
	for _, b := range w.dane.BlokadyStale {
		if b.Dzien != s.p.Dzien || b.SlotOd > s.slotDo() || s.p.SlotStart > b.SlotDo {
			continue
		}
		for _, g := range s.z.Grupy {
			if w.ix.GrupyKoliduja(g, b.Grupa) {
				dodaj(RegulaBlokady, fmt.Sprintf("grupa %s ma blokadę %s (dzień %d, sloty %d–%d)", g, b.Grupa, b.Dzien, b.SlotOd, b.SlotDo), s.z.ID)
				break
			}
		}
	}
}

// reguly345: dwa nakładające się spotkania nie mogą mieć wspólnego prowadzącego, sali ani studenta.
func (w *Walidator) reguly345(a, b *spotkanie, dodaj dodajFunc) {
	if !nakladaja(a, b) {
		return
	}
	wspolne := fmt.Sprintf("dzień %d, sloty %d–%d i %d–%d, wspólne tygodnie %v",
		a.p.Dzien, a.p.SlotStart, a.slotDo(), b.p.SlotStart, b.slotDo(), (a.p.Tygodnie & b.p.Tygodnie).Lista())
	for _, p := range a.z.Prowadzacy {
		if slices.Contains(b.z.Prowadzacy, p) {
			dodaj(RegulaProwadzacy, fmt.Sprintf("prowadzący %s w dwóch miejscach: %s", p, wspolne), a.z.ID, b.z.ID)
		}
	}
	if a.sala.ID == b.sala.ID {
		dodaj(RegulaSala, fmt.Sprintf("sala %s zajęta dwa razy: %s", a.sala.ID, wspolne), a.z.ID, b.z.ID)
	}
	if w.ix.ZajeciaKolidujaStudenci(a.z, b.z) {
		dodaj(RegulaStudent, "wspólni studenci: "+wspolne, a.z.ID, b.z.ID)
	}
}

// regula10: zestawy rozne_dni w różnych dniach, zestawy ten_sam_termin w tym samym dniu, slocie i sali.
// Zajęcia bez poprawnego przypisania (reguła 1) pomijamy, bo ich naruszenie jest już zgłoszone.
func (w *Walidator) regula10(spotkania []*spotkanie, dodaj dodajFunc) {
	poZajeciach := make(map[string]*spotkanie, len(spotkania))
	for _, s := range spotkania {
		poZajeciach[s.z.ID] = s
	}
	nazwy := make([]string, 0, len(w.ix.Powiazania))
	for n := range w.ix.Powiazania {
		nazwy = append(nazwy, n)
	}
	slices.Sort(nazwy)
	for _, nazwa := range nazwy {
		zs := w.ix.Powiazania[nazwa]
		for i, z1 := range zs {
			for _, z2 := range zs[i+1:] {
				a, b := poZajeciach[z1.ID], poZajeciach[z2.ID]
				if a == nil || b == nil {
					continue
				}
				switch z1.TypPowiazania {
				case model.RozneDni:
					if a.p.Dzien == b.p.Dzien {
						dodaj(RegulaPowiazania, fmt.Sprintf("%s (rozne_dni): oba w dniu %d", nazwa, a.p.Dzien), z1.ID, z2.ID)
					}
				case model.TenSamTermin:
					if a.p.Dzien != b.p.Dzien || a.p.SlotStart != b.p.SlotStart || a.sala.ID != b.sala.ID {
						dodaj(RegulaPowiazania, fmt.Sprintf("%s (ten_sam_termin): dzień/slot/sala %d/%d/%s i %d/%d/%s",
							nazwa, a.p.Dzien, a.p.SlotStart, a.sala.ID, b.p.Dzien, b.p.SlotStart, b.sala.ID), z1.ID, z2.ID)
					}
				}
			}
		}
	}
}
