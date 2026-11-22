// Package validator sprawdza gotowy plan pod kątem 10 ograniczeń twardych z data/README.md (sekcja 5).
//
// To niezależny dowód poprawności planu: pakiet nie importuje niczego z internal/solver
// i nie zakłada niczego o tym, jak plan powstał.
//
// Wynik to pełna lista naruszeń w formacie raportu z docs/contracts/plan-format.md:
// reguła, zajęcia i położenie (dzień, sloty, sala, tygodnie), jedno naruszenie na regułę i zestaw zajęć.
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

// nazwyRegul to stałe identyfikatory reguł (docs/contracts/api.md, „Nazwy reguł w raporcie”).
var nazwyRegul = [...]string{
	RegulaPrzypisanie: "przypisanie",
	RegulaTygodnie:    "tygodnie",
	RegulaProwadzacy:  "kolizja_prowadzacego",
	RegulaSala:        "kolizja_sali",
	RegulaStudent:     "kolizja_grup",
	RegulaTypSali:     "typ_sali",
	RegulaPojemnosc:   "pojemnosc",
	RegulaDezyderaty:  "dezyderaty",
	RegulaBlokady:     "blokady_stale",
	RegulaPowiazania:  "powiazania",
}

// NazwaReguly zwraca stały identyfikator reguły 1–10 (pusty dla nieznanego numeru).
func NazwaReguly(regula int) string {
	if regula < 1 || regula >= len(nazwyRegul) {
		return ""
	}
	return nazwyRegul[regula]
}

// Naruszenie to jedno złamanie reguły twardej w formacie raportu (docs/contracts/plan-format.md).
// Pola położenia są puste, gdy nie dotyczą naruszenia (np. zajęcia bez przypisania nie mają dnia,
// a sala jest podana tylko wtedy, gdy wszystkie zajęcia naruszenia są w tej samej sali).
type Naruszenie struct {
	Regula   int      `json:"regula"`
	Nazwa    string   `json:"nazwa"`
	Zajecia  []string `json:"zajecia_id"`         // zajęcia, których dotyczy (posortowane)
	Dzien    int      `json:"dzien,omitempty"`    // 1–5
	Sloty    []int    `json:"sloty,omitempty"`    // sloty, w których występuje naruszenie (rosnąco)
	SalaID   string   `json:"sala_id,omitempty"`  // sala, jeśli dotyczy
	Tygodnie []int    `json:"tygodnie,omitempty"` // tygodnie, w których występuje naruszenie (rosnąco)
	Opis     string   `json:"opis"`               // jedno zdanie po polsku do wyświetlenia w UI
}

func (n Naruszenie) String() string {
	s := fmt.Sprintf("reguła %d (%s) [%s]", n.Regula, n.Nazwa, strings.Join(n.Zajecia, ", "))
	if n.Dzien > 0 {
		s += fmt.Sprintf(" dzień %d", n.Dzien)
	}
	if len(n.Sloty) > 0 {
		s += fmt.Sprintf(" sloty %v", n.Sloty)
	}
	if n.SalaID != "" {
		s += " sala " + n.SalaID
	}
	if len(n.Tygodnie) > 0 {
		s += fmt.Sprintf(" tygodnie %v", n.Tygodnie)
	}
	return s + ": " + n.Opis
}

// Powody wykluczenia slotu prowadzącego (reguła 8).
const (
	wykluczonyDezyderat uint8 = 1 << iota
	wykluczonyRada
)

// Walidator sprawdza plany dla jednego zestawu danych. Można go używać wielokrotnie
// i współbieżnie (np. przy sprawdzaniu ruchów w UI), bo nie zmienia swojego stanu.
type Walidator struct {
	dane *model.Dane
	ix   *model.Indeks

	// wykluczone: prowadzący → [dzień][slot] powody wykluczenia (dezyderaty i Rada Wydziału).
	wykluczone map[string]*[model.LiczbaDni + 1][model.LiczbaSlotow + 1]uint8
}

// Nowy przygotowuje walidator. Dane powinny przejść model.Sprawdz (robi to model.Wczytaj).
func Nowy(d *model.Dane) *Walidator {
	w := &Walidator{
		dane:       d,
		ix:         model.NowyIndeks(d),
		wykluczone: make(map[string]*[model.LiczbaDni + 1][model.LiczbaSlotow + 1]uint8),
	}
	zablokuj := func(p string, dzien, slot int, powod uint8) {
		if dzien < 1 || dzien > model.LiczbaDni || slot < 1 || slot > model.LiczbaSlotow {
			return
		}
		if w.wykluczone[p] == nil {
			w.wykluczone[p] = new([model.LiczbaDni + 1][model.LiczbaSlotow + 1]uint8)
		}
		w.wykluczone[p][dzien][slot] |= powod
	}
	for _, t := range d.DezyderatyTerminy {
		// Oba typy (wykluczony, rada_wydzialu) blokują prowadzącego; typ zmienia tylko opis.
		powod := wykluczonyDezyderat
		if t.Typ == model.DezyderatRady {
			powod = wykluczonyRada
		}
		zablokuj(t.Prowadzacy, t.Dzien, t.Slot, powod)
	}
	for _, p := range d.Prowadzacy {
		if !p.CzlonekRadyWydzialu {
			continue
		}
		for _, r := range d.RadaWydzialu {
			for s := r.SlotOd; s <= r.SlotDo; s++ {
				zablokuj(p.ID, r.Dzien, s, wykluczonyRada)
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

func (s *spotkanie) slotDo() int  { return s.p.SlotStart + s.z.SlotyNaSpotkanie - 1 }
func (s *spotkanie) sloty() []int { return zakres(s.p.SlotStart, s.slotDo()) }

// zakres zwraca liczby od … do włącznie (pusty, gdy od > do).
func zakres(od, do int) []int {
	var r []int
	for i := od; i <= do; i++ {
		r = append(r, i)
	}
	return r
}

// nakladaja: ten sam dzień, wspólny slot i wspólny tydzień (data/README.md, sekcja 5).
func nakladaja(a, b *spotkanie) bool {
	return a.p.Dzien == b.p.Dzien &&
		a.p.SlotStart <= b.slotDo() && b.p.SlotStart <= a.slotDo() &&
		a.p.Tygodnie.Nakladaja(b.p.Tygodnie)
}

var (
	dni       = [...]string{1: "poniedziałek", 2: "wtorek", 3: "środa", 4: "czwartek", 5: "piątek"}
	wDniuNazw = [...]string{1: "w poniedziałek", 2: "we wtorek", 3: "w środę", 4: "w czwartek", 5: "w piątek"}
)

// nazwaDnia: „poniedziałek” … „piątek”, a dla dnia spoza zakresu „dzień N”.
func nazwaDnia(d int) string {
	if d >= 1 && d < len(dni) {
		return dni[d]
	}
	return fmt.Sprintf("dzień %d", d)
}

// wDniu: „w poniedziałek”, „we wtorek” … „w piątek”.
func wDniu(d int) string {
	if d >= 1 && d < len(wDniuNazw) {
		return wDniuNazw[d]
	}
	return fmt.Sprintf("w dniu %d", d)
}

// opisSlotow: „slot 3” albo „sloty 3–4”.
func opisSlotow(od, do int) string {
	if od == do {
		return fmt.Sprintf("slot %d", od)
	}
	return fmt.Sprintf("sloty %d–%d", od, do)
}

// Sprawdz zwraca wszystkie naruszenia reguł twardych, posortowane po regule i zajęciach.
// Pusty wynik oznacza plan poprawny. Kilka problemów tej samej reguły z tymi samymi zajęciami
// (np. zła sala i brak wyposażenia) to jedno naruszenie z połączonym opisem.
func (w *Walidator) Sprawdz(plan *model.Plan) []Naruszenie {
	z := &zbieracz{indeks: map[string]int{}}

	spotkania := w.regula1(plan, z)
	for _, s := range spotkania {
		w.regula2(s, z)
		w.regula6i7(s, z)
		w.regula8(s, z)
		w.regula9(s, z)
	}
	for i, a := range spotkania {
		for _, b := range spotkania[i+1:] {
			w.reguly345(a, b, z)
		}
	}
	w.regula10(spotkania, z)

	wynik := z.wynik
	slices.SortStableFunc(wynik, func(a, b Naruszenie) int {
		if a.Regula != b.Regula {
			return a.Regula - b.Regula
		}
		return slices.Compare(a.Zajecia, b.Zajecia)
	})
	return wynik
}

// zbieracz zbiera naruszenia i łączy te o tej samej regule i tym samym zestawie zajęć.
type zbieracz struct {
	wynik  []Naruszenie
	indeks map[string]int // klucz reguła + zajęcia → pozycja w wynik
}

func (z *zbieracz) dodaj(n Naruszenie) {
	n.Zajecia = slices.Clone(n.Zajecia)
	slices.Sort(n.Zajecia)
	n.Nazwa = NazwaReguly(n.Regula)
	n.Sloty = posortowane(n.Sloty)
	n.Tygodnie = posortowane(n.Tygodnie)

	klucz := fmt.Sprint(n.Regula, n.Zajecia)
	i, jest := z.indeks[klucz]
	if !jest {
		z.indeks[klucz] = len(z.wynik)
		z.wynik = append(z.wynik, n)
		return
	}
	s := &z.wynik[i]
	s.Opis += "; " + n.Opis
	if s.Dzien != n.Dzien {
		s.Dzien = 0
	}
	if s.SalaID != n.SalaID {
		s.SalaID = ""
	}
	s.Sloty = posortowane(append(s.Sloty, n.Sloty...))
	s.Tygodnie = posortowane(append(s.Tygodnie, n.Tygodnie...))
}

// posortowane zwraca rosnącą listę bez powtórzeń (nil dla pustej).
func posortowane(l []int) []int {
	if len(l) == 0 {
		return nil
	}
	l = slices.Clone(l)
	slices.Sort(l)
	return slices.Compact(l)
}

// polozenie to dzień, sloty, sala i tygodnie spotkania, do wpisania w naruszenie.
func polozenie(regula int, s *spotkanie, opis string) Naruszenie {
	return Naruszenie{
		Regula:   regula,
		Zajecia:  []string{s.z.ID},
		Dzien:    s.p.Dzien,
		Sloty:    s.sloty(),
		SalaID:   s.sala.ID,
		Tygodnie: s.p.Tygodnie.Lista(),
		Opis:     opis,
	}
}

// regula1 sprawdza, że każde zajęcia mają dokładnie jedno przypisanie z poprawnym dniem, slotami i salą.
// Zwraca tylko przypisania, które przeszły regułę 1. Pozostałe reguły sprawdzamy już tylko dla nich.
func (w *Walidator) regula1(plan *model.Plan, zb *zbieracz) []*spotkanie {
	widziane := make(map[string]bool, len(plan.Przypisania))
	var spotkania []*spotkanie
	for _, p := range plan.Przypisania {
		// Położenie z przypisania, ograniczone do poprawnych wartości.
		n := Naruszenie{Regula: RegulaPrzypisanie, Zajecia: []string{p.ZajeciaID}, Tygodnie: p.Tygodnie.Lista()}
		if p.Dzien >= 1 && p.Dzien <= model.LiczbaDni {
			n.Dzien = p.Dzien
		}
		if w.ix.Sale[p.SalaID] != nil {
			n.SalaID = p.SalaID
		}
		blad := func(opis string) {
			n.Opis = opis
			zb.dodaj(n)
		}

		z := w.ix.Zajecia[p.ZajeciaID]
		if z == nil {
			blad(fmt.Sprintf("Przypisanie dotyczy nieznanych zajęć %s.", p.ZajeciaID))
			continue
		}
		if widziane[p.ZajeciaID] {
			blad(fmt.Sprintf("Zajęcia %s są przypisane więcej niż raz.", z.ID))
			continue
		}
		widziane[p.ZajeciaID] = true
		koniec := p.SlotStart + z.SlotyNaSpotkanie - 1
		n.Sloty = zakres(max(p.SlotStart, 1), min(koniec, model.LiczbaSlotow))

		ok := true
		if n.Dzien == 0 {
			blad(fmt.Sprintf("Zajęcia %s mają dzień %d spoza zakresu 1–%d.", z.ID, p.Dzien, model.LiczbaDni))
			ok = false
		}
		if p.SlotStart < 1 || koniec > model.LiczbaSlotow {
			blad(fmt.Sprintf("Zajęcia %s zajmują %s, poza zakresem 1–%d.", z.ID, opisSlotow(p.SlotStart, koniec), model.LiczbaSlotow))
			ok = false
		}
		sala := w.ix.Sale[p.SalaID]
		if sala == nil {
			blad(fmt.Sprintf("Zajęcia %s mają nieznaną salę %q.", z.ID, p.SalaID))
			ok = false
		}
		if ok {
			spotkania = append(spotkania, &spotkanie{p: p, z: z, sala: sala})
		}
	}
	for _, z := range w.dane.Zajecia {
		if !widziane[z.ID] {
			zb.dodaj(Naruszenie{Regula: RegulaPrzypisanie, Zajecia: []string{z.ID},
				Opis: fmt.Sprintf("Zajęcia %s nie mają przypisania.", z.ID)})
		}
	}
	return spotkania
}

// regula2: tygodnie zgodne z liczbą spotkań, zakresem, częstotliwością i tygodniami niedostępnymi.
// Pole tygodnie naruszenia to tygodnie błędne (spoza dozwolonych) albo brakujące, a gdy chodzi
// o samą liczbę spotkań, wszystkie tygodnie spotkania.
func (w *Walidator) regula2(s *spotkanie, zb *zbieracz) {
	z, tyg := s.z, s.p.Tygodnie
	dozwolone := z.DozwoloneTygodnie()
	n := polozenie(RegulaTygodnie, s, "")
	if bledne := tyg.Bez(dozwolone); bledne != 0 {
		n.Tygodnie = bledne.Lista()
		n.Opis = fmt.Sprintf("Zajęcia %s odbywają się w tygodniach %v spoza dozwolonych (zakres %d–%d",
			z.ID, n.Tygodnie, z.TydzienOd, z.TydzienDo)
		if len(z.TygodnieNiedostepne) > 0 {
			n.Opis += fmt.Sprintf(", niedostępne %v", z.TygodnieNiedostepne)
		}
		n.Opis += ")."
		zb.dodaj(n)
		return
	}
	if z.Czestotliwosc == model.CoTydzien && tyg != dozwolone {
		// Tygodnie są podzbiorem dozwolonych, więc różnica to tygodnie brakujące.
		n.Tygodnie = dozwolone.Bez(tyg).Lista()
		n.Opis = fmt.Sprintf("Zajęcia %s są co tydzień, ale brakuje tygodni %v.", z.ID, n.Tygodnie)
		zb.dodaj(n)
		return
	}
	if tyg.Liczba() != z.LiczbaSpotkan {
		n.Opis = fmt.Sprintf("Zajęcia %s mają %d spotkań, a powinny mieć %d.", z.ID, tyg.Liczba(), z.LiczbaSpotkan)
		zb.dodaj(n)
		return
	}
	switch z.Czestotliwosc {
	case model.CoTydzien:
		// Sprawdzone wyżej.
	case model.Co2Tygodnie:
		if tyg != dozwolone.Parzyste() && tyg != dozwolone.Nieparzyste() {
			n.Opis = fmt.Sprintf("Zajęcia %s są co 2 tygodnie, ale tygodnie %v nie są wszystkimi parzystymi ani wszystkimi nieparzystymi.",
				z.ID, n.Tygodnie)
			zb.dodaj(n)
		}
	case model.WybraneTygodnie:
		// Dowolne LiczbaSpotkan tygodni z dozwolonych (ADR 0004): sprawdzone wyżej.
	}
}

// regula6i7: dozwolona sala z wyposażeniem i wystarczającą pojemnością.
func (w *Walidator) regula6i7(s *spotkanie, zb *zbieracz) {
	z, sala := s.z, s.sala
	blad := func(regula int, opis string) { zb.dodaj(polozenie(regula, s, opis)) }
	if len(z.DozwoloneSale) > 0 {
		// Sale z dozwolone_sale wolno użyć nawet wtedy, gdy nie są dostępne do planowania (data/README.md).
		if !slices.Contains(z.DozwoloneSale, sala.ID) {
			blad(RegulaTypSali, fmt.Sprintf("Sala %s nie jest na liście dozwolonych sal %v zajęć %s.", sala.ID, z.DozwoloneSale, z.ID))
		}
	} else {
		if !slices.Contains(z.TypySal, sala.Typ) {
			blad(RegulaTypSali, fmt.Sprintf("Sala %s jest typu %s, a zajęcia %s wymagają typu %v.", sala.ID, sala.Typ, z.ID, z.TypySal))
		}
		if !sala.DostepnaDoPlanowania {
			blad(RegulaTypSali, fmt.Sprintf("Sala %s nie jest dostępna do planowania.", sala.ID))
		}
	}
	var brak []string
	for _, wyp := range z.WymaganeWyposazenie {
		if !slices.Contains(sala.Wyposazenie, wyp) {
			brak = append(brak, wyp)
		}
	}
	if len(brak) > 0 {
		blad(RegulaTypSali, fmt.Sprintf("Sala %s nie ma wyposażenia %s wymaganego przez zajęcia %s.", sala.ID, strings.Join(brak, ", "), z.ID))
	}
	// Bez marginesu, także dla pojemności szacowanych (ADR 0005).
	if sala.Pojemnosc < z.LiczbaOsob() {
		opis := fmt.Sprintf("Sala %s ma %d miejsc, a zajęcia %s mają %d osób.", sala.ID, sala.Pojemnosc, z.ID, z.LiczbaOsob())
		if sala.PojemnoscSzacowana {
			opis = fmt.Sprintf("Sala %s ma szacunkowo %d miejsc, a zajęcia %s mają %d osób.", sala.ID, sala.Pojemnosc, z.ID, z.LiczbaOsob())
		}
		blad(RegulaPojemnosc, opis)
	}
}

// regula8: prowadzący nie ma zajęć w godzinach wykluczonych ani w czasie Rady Wydziału.
// Sloty naruszenia to sloty wykluczone (wszystkich prowadzących zajęć razem).
func (w *Walidator) regula8(s *spotkanie, zb *zbieracz) {
	for _, p := range s.z.Prowadzacy {
		wyk := w.wykluczone[p]
		if wyk == nil {
			continue
		}
		var sloty []int
		var powody uint8
		for slot := s.p.SlotStart; slot <= s.slotDo(); slot++ {
			if v := wyk[s.p.Dzien][slot]; v != 0 {
				sloty = append(sloty, slot)
				powody |= v
			}
		}
		if len(sloty) == 0 {
			continue
		}
		var co []string
		if powody&wykluczonyDezyderat != 0 {
			co = append(co, "godziny wykluczone w dezyderatach")
		}
		if powody&wykluczonyRada != 0 {
			co = append(co, "Radę Wydziału")
		}
		n := polozenie(RegulaDezyderaty, s, fmt.Sprintf("Prowadzący %s ma zajęcia %s %s (sloty %v), a ma tam %s.",
			p, s.z.ID, wDniu(s.p.Dzien), sloty, strings.Join(co, " i ")))
		n.Sloty = sloty
		zb.dodaj(n)
	}
}

// regula9: zajęcia nie nakładają się na blokady stałe grup, z którymi mają wspólnych studentów.
// Blokady obowiązują w każdym tygodniu, więc wystarczy dzień i sloty.
func (w *Walidator) regula9(s *spotkanie, zb *zbieracz) {
	for _, b := range w.dane.BlokadyStale {
		if b.Dzien != s.p.Dzien || b.SlotOd > s.slotDo() || s.p.SlotStart > b.SlotDo {
			continue
		}
		for _, g := range s.z.Grupy {
			if !w.ix.GrupyKoliduja(g, b.Grupa) {
				continue
			}
			opis := fmt.Sprintf("Zajęcia %s grupy %s nakładają się na blokadę grupy %s (%s, %s)",
				s.z.ID, g, b.Grupa, nazwaDnia(b.Dzien), opisSlotow(b.SlotOd, b.SlotDo))
			if b.Opis != "" {
				opis += ": " + b.Opis
			}
			n := polozenie(RegulaBlokady, s, opis+".")
			n.Sloty = zakres(max(b.SlotOd, s.p.SlotStart), min(b.SlotDo, s.slotDo()))
			zb.dodaj(n)
			break
		}
	}
}

// reguly345: dwa nakładające się spotkania nie mogą mieć wspólnego prowadzącego, sali ani studenta.
// Położenie to wspólne sloty i wspólne tygodnie; sala tylko, gdy obie zajęcia są w tej samej.
func (w *Walidator) reguly345(a, b *spotkanie, zb *zbieracz) {
	if !nakladaja(a, b) {
		return
	}
	n := Naruszenie{
		Zajecia:  []string{a.z.ID, b.z.ID},
		Dzien:    a.p.Dzien,
		Sloty:    zakres(max(a.p.SlotStart, b.p.SlotStart), min(a.slotDo(), b.slotDo())),
		Tygodnie: (a.p.Tygodnie & b.p.Tygodnie).Lista(),
	}
	if a.sala.ID == b.sala.ID {
		n.SalaID = a.sala.ID
	}
	kiedy := fmt.Sprintf("%s, %s, tygodnie %v", wDniu(n.Dzien), opisSlotow(n.Sloty[0], n.Sloty[len(n.Sloty)-1]), n.Tygodnie)

	var wspolni []string
	for _, p := range a.z.Prowadzacy {
		if slices.Contains(b.z.Prowadzacy, p) {
			wspolni = append(wspolni, p)
		}
	}
	if len(wspolni) > 0 {
		n.Regula = RegulaProwadzacy
		kto := "Prowadzący " + wspolni[0] + " ma"
		if len(wspolni) > 1 {
			kto = "Prowadzący " + strings.Join(wspolni, ", ") + " mają"
		}
		n.Opis = fmt.Sprintf("%s jednocześnie zajęcia %s i %s %s.", kto, a.z.ID, b.z.ID, kiedy)
		zb.dodaj(n)
	}
	if a.sala.ID == b.sala.ID {
		n.Regula = RegulaSala
		n.Opis = fmt.Sprintf("W sali %s są jednocześnie zajęcia %s i %s %s.", a.sala.ID, a.z.ID, b.z.ID, kiedy)
		zb.dodaj(n)
	}
	if w.ix.ZajeciaKolidujaStudenci(a.z, b.z) {
		n.Regula = RegulaStudent
		ga, gb := strings.Join(a.z.Grupy, ", "), strings.Join(b.z.Grupy, ", ")
		if ga == gb {
			n.Opis = fmt.Sprintf("Grupa %s ma jednocześnie zajęcia %s i %s %s.", ga, a.z.ID, b.z.ID, kiedy)
		} else {
			n.Opis = fmt.Sprintf("Grupy %s i %s mają wspólnych studentów, a zajęcia %s i %s są jednocześnie %s.",
				ga, gb, a.z.ID, b.z.ID, kiedy)
		}
		zb.dodaj(n)
	}
}

// regula10: zestawy rozne_dni w różnych dniach, zestawy ten_sam_termin w tym samym dniu, slocie i sali.
// Zajęcia bez poprawnego przypisania (reguła 1) pomijamy, bo ich naruszenie jest już zgłoszone.
func (w *Walidator) regula10(spotkania []*spotkanie, zb *zbieracz) {
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
				n := Naruszenie{Regula: RegulaPowiazania, Zajecia: []string{z1.ID, z2.ID}}
				switch z1.TypPowiazania {
				case model.RozneDni:
					if a.p.Dzien != b.p.Dzien {
						continue
					}
					n.Dzien = a.p.Dzien
					n.Opis = fmt.Sprintf("Zajęcia %s i %s z powiązania %s mają być w różnych dniach, a oba są %s.",
						z1.ID, z2.ID, nazwa, wDniu(a.p.Dzien))
				case model.TenSamTermin:
					if a.p.Dzien == b.p.Dzien && a.p.SlotStart == b.p.SlotStart && a.sala.ID == b.sala.ID {
						continue
					}
					if a.p.Dzien == b.p.Dzien {
						n.Dzien = a.p.Dzien
					}
					if a.sala.ID == b.sala.ID {
						n.SalaID = a.sala.ID
					}
					n.Opis = fmt.Sprintf("Zajęcia %s i %s z powiązania %s mają być w tym samym terminie i sali, a są: %s %s, %s, sala %s; %s %s, %s, sala %s.",
						z1.ID, z2.ID, nazwa,
						z1.ID, wDniu(a.p.Dzien), opisSlotow(a.p.SlotStart, a.slotDo()), a.sala.ID,
						z2.ID, wDniu(b.p.Dzien), opisSlotow(b.p.SlotStart, b.slotDo()), b.sala.ID)
				default:
					continue
				}
				zb.dodaj(n)
			}
		}
	}
}
