package model

import (
	"errors"
	"fmt"
	"slices"
)

// Sprawdz sprawdza spójność danych wejściowych: unikalność identyfikatorów, poprawność wartości
// wyliczeniowych i zakresów, odwołania między tabelami, zgodność kolizje_grup ze studenci_grupy
// i kalendarza z siatką 5 dni × 14 tygodni. Zwraca wszystkie znalezione problemy naraz.
//
// To nie jest walidacja planu (reguły twarde 1–10): tę robi internal/validator.
func (d *Dane) Sprawdz() error {
	s := &sprawdzanie{}
	ix := NowyIndeks(d)

	s.unikalne("sale", len(d.Sale), func(i int) string { return d.Sale[i].ID })
	s.unikalne("prowadzacy", len(d.Prowadzacy), func(i int) string { return d.Prowadzacy[i].ID })
	s.unikalne("roczniki", len(d.Roczniki), func(i int) string { return d.Roczniki[i].ID })
	s.unikalne("grupy", len(d.Grupy), func(i int) string { return d.Grupy[i].ID })
	s.unikalne("studenci", len(d.Studenci), func(i int) string { return d.Studenci[i].ID })
	s.unikalne("przedmioty", len(d.Przedmioty), func(i int) string { return d.Przedmioty[i].ID })
	s.unikalne("zajecia", len(d.Zajecia), func(i int) string { return d.Zajecia[i].ID })
	s.unikalne("dezyderaty_ogolne", len(d.DezyderatyOgolne), func(i int) string { return d.DezyderatyOgolne[i].Prowadzacy })

	d.sprawdzSloty(s)
	d.sprawdzKalendarz(s)
	for i := range d.Sale {
		sala := &d.Sale[i]
		s.wyliczenie("sale "+sala.ID+": typ", sala.Typ, typySal)
		if sala.DostepnaDoPlanowania && sala.Pojemnosc <= 0 {
			s.blad("sale %s: dostępna do planowania, ale bez pojemności", sala.ID)
		}
	}
	for _, g := range d.Grupy {
		istnieje(s, "grupy "+g.ID+": rocznik", g.Rocznik, ix.Roczniki)
		if g.GrupaNadrzedna != "" {
			istnieje(s, "grupy "+g.ID+": grupa_nadrzedna", g.GrupaNadrzedna, ix.Grupy)
		}
		if n := len(ix.StudenciGrupy[g.ID]); n != g.LiczbaStudentow {
			s.blad("grupy %s: liczba_studentow = %d, a w studenci_grupy jest %d", g.ID, g.LiczbaStudentow, n)
		}
	}
	studenci := make(map[string]bool, len(d.Studenci))
	for _, st := range d.Studenci {
		studenci[st.ID] = true
		istnieje(s, "studenci "+st.ID+": rocznik", st.Rocznik, ix.Roczniki)
	}
	for _, sg := range d.StudenciGrupy {
		if !studenci[sg.Student] {
			s.blad("studenci_grupy: nieznany student %q", sg.Student)
		}
		istnieje(s, "studenci_grupy: grupa", sg.Grupa, ix.Grupy)
	}
	for _, p := range d.Przedmioty {
		istnieje(s, "przedmioty "+p.ID+": rocznik", p.Rocznik, ix.Roczniki)
	}
	d.sprawdzKolizje(s, ix)
	for i := range d.Zajecia {
		d.Zajecia[i].sprawdz(s, ix)
	}
	d.sprawdzPowiazania(s, ix)
	d.sprawdzTerminy(s, ix)

	return errors.Join(s.bledy...)
}

type sprawdzanie struct{ bledy []error }

func (s *sprawdzanie) blad(format string, args ...any) {
	s.bledy = append(s.bledy, fmt.Errorf(format, args...))
}

func (s *sprawdzanie) unikalne(tabela string, n int, id func(int) string) {
	widziane := make(map[string]bool, n)
	for i := range n {
		if widziane[id(i)] {
			s.blad("%s: powtórzony identyfikator %q", tabela, id(i))
		}
		widziane[id(i)] = true
	}
}

// istnieje to funkcja, a nie metoda, bo metody w Go nie mogą być generyczne.
func istnieje[T any](s *sprawdzanie, gdzie, id string, m map[string]*T) {
	if _, ok := m[id]; !ok {
		s.blad("%s: nieznany identyfikator %q", gdzie, id)
	}
}

func (s *sprawdzanie) wyliczenie(gdzie string, v any, dozwolone []any) {
	if !slices.Contains(dozwolone, v) {
		s.blad("%s: niedozwolona wartość %q", gdzie, v)
	}
}

func (s *sprawdzanie) zakres(gdzie string, v, od, do int) {
	if v < od || v > do {
		s.blad("%s: wartość %d poza zakresem %d–%d", gdzie, v, od, do)
	}
}

var (
	typySal         = []any{Aula, SalaDydaktyczna, PracowniaKomputerowa, PracowniaRobotyki, PracowniaSpecjalistyczna}
	formy           = []any{Wyklad, Konwersatorium, LabInformatyczne, Pracownia, Lektorat, Seminarium}
	czestotliwosci  = []any{CoTydzien, Co2Tygodnie, WybraneTygodnie}
	tryby           = []any{Stacjonarny, Hybrydowy}
	typyPowiazan    = []any{RozneDni, TenSamTermin}
	typyDni         = []any{DzienZajec, DzienWolny, DzienOdpracowania}
	typyDezyderatow = []any{Wykluczony, DezyderatRady}
	preferencjeDni  = []any{Skumulowane, Nieskumulowane, BrakPreferencji}
)

func (d *Dane) sprawdzSloty(s *sprawdzanie) {
	if len(d.Sloty) != LiczbaSlotow {
		s.blad("sloty: oczekiwano %d slotów, jest %d", LiczbaSlotow, len(d.Sloty))
	}
	for i, sl := range d.Sloty {
		if sl.Slot != i+1 {
			s.blad("sloty: pozycja %d ma numer %d", i+1, sl.Slot)
		}
	}
}

// sprawdzKalendarz: każda para (dzień planu, nr tygodnia) występuje dokładnie raz.
func (d *Dane) sprawdzKalendarz(s *sprawdzanie) {
	var widziane [LiczbaDni + 1][LiczbaTygodni + 1]int
	for _, k := range d.Kalendarz {
		s.wyliczenie("kalendarz "+k.Data+": typ_dnia", k.TypDnia, typyDni)
		if k.TypDnia == DzienWolny {
			if k.PlanDnia != 0 || k.NrTygodnia != 0 {
				s.blad("kalendarz %s: dzień wolny z planem dnia", k.Data)
			}
			continue
		}
		if k.PlanDnia < 1 || k.PlanDnia > LiczbaDni || k.NrTygodnia < 1 || k.NrTygodnia > LiczbaTygodni {
			s.blad("kalendarz %s: plan_dnia %d / nr_tygodnia %d poza zakresem", k.Data, k.PlanDnia, k.NrTygodnia)
			continue
		}
		widziane[k.PlanDnia][k.NrTygodnia]++
	}
	for dzien := 1; dzien <= LiczbaDni; dzien++ {
		for t := 1; t <= LiczbaTygodni; t++ {
			if n := widziane[dzien][t]; n != 1 {
				s.blad("kalendarz: dzień planu %d, tydzień %d występuje %d razy (oczekiwano 1)", dzien, t, n)
			}
		}
	}
}

// sprawdzKolizje porównuje dostarczoną tabelę kolizje_grup z kolizjami policzonymi ze studenci_grupy.
func (d *Dane) sprawdzKolizje(s *sprawdzanie, ix *Indeks) {
	podane := make(map[paraGrup]bool, len(d.KolizjeGrup))
	for _, k := range d.KolizjeGrup {
		istnieje(s, "kolizje_grup: grupa_a", k.GrupaA, ix.Grupy)
		istnieje(s, "kolizje_grup: grupa_b", k.GrupaB, ix.Grupy)
		podane[nowaPara(k.GrupaA, k.GrupaB)] = true
		if n := ix.WspolniStudenci(k.GrupaA, k.GrupaB); n != k.WspolniStudenci {
			s.blad("kolizje_grup %s–%s: podano %d wspólnych studentów, w studenci_grupy jest %d",
				k.GrupaA, k.GrupaB, k.WspolniStudenci, n)
		}
	}
	for p := range ix.kolizje {
		if !podane[p] {
			s.blad("kolizje_grup: brak pary %s–%s, choć mają wspólnych studentów", p.a, p.b)
		}
	}
}

func (z *Zajecia) sprawdz(s *sprawdzanie, ix *Indeks) {
	gdzie := "zajecia " + z.ID
	istnieje(s, gdzie+": rocznik", z.Rocznik, ix.Roczniki)
	istnieje(s, gdzie+": przedmiot_id", z.PrzedmiotID, ix.Przedmioty)
	s.wyliczenie(gdzie+": forma", z.Forma, formy)
	s.wyliczenie(gdzie+": czestotliwosc", z.Czestotliwosc, czestotliwosci)
	s.wyliczenie(gdzie+": tryb", z.Tryb, tryby)
	if len(z.Grupy) == 0 {
		s.blad("%s: brak grup", gdzie)
	}
	for _, g := range z.Grupy {
		istnieje(s, gdzie+": grupy", g, ix.Grupy)
	}
	if len(z.Prowadzacy) == 0 {
		s.blad("%s: brak prowadzących", gdzie)
	}
	for _, p := range z.Prowadzacy {
		istnieje(s, gdzie+": prowadzacy", p, ix.Prowadzacy)
	}
	if len(z.TypySal) == 0 && len(z.DozwoloneSale) == 0 {
		s.blad("%s: puste typy_sal i dozwolone_sale", gdzie)
	}
	for _, t := range z.TypySal {
		s.wyliczenie(gdzie+": typy_sal", t, typySal)
	}
	for _, sala := range z.DozwoloneSale {
		istnieje(s, gdzie+": dozwolone_sale", sala, ix.Sale)
	}
	if z.LiczbaStudentow < 0 || z.DodatkowiStudenci < 0 {
		s.blad("%s: ujemna liczba studentów", gdzie)
	}
	s.zakres(gdzie+": sloty_na_spotkanie", z.SlotyNaSpotkanie, 1, LiczbaSlotow)
	s.zakres(gdzie+": tydzien_od", z.TydzienOd, 1, LiczbaTygodni)
	s.zakres(gdzie+": tydzien_do", z.TydzienDo, z.TydzienOd, LiczbaTygodni)
	for _, t := range z.TygodnieNiedostepne {
		s.zakres(gdzie+": tygodnie_niedostepne", t, 1, LiczbaTygodni)
	}

	// Liczba spotkań musi dać się zrealizować przy danej częstotliwości.
	w := z.DozwoloneTygodnie()
	switch z.Czestotliwosc {
	case CoTydzien:
		if z.LiczbaSpotkan != w.Liczba() {
			s.blad("%s: co_tydzien, liczba_spotkan %d ≠ %d dozwolonych tygodni", gdzie, z.LiczbaSpotkan, w.Liczba())
		}
	case Co2Tygodnie:
		if z.LiczbaSpotkan != w.Parzyste().Liczba() && z.LiczbaSpotkan != w.Nieparzyste().Liczba() {
			s.blad("%s: co_2_tygodnie, liczba_spotkan %d nie pasuje ani do tygodni parzystych (%d), ani nieparzystych (%d)",
				gdzie, z.LiczbaSpotkan, w.Parzyste().Liczba(), w.Nieparzyste().Liczba())
		}
	case WybraneTygodnie:
		if z.LiczbaSpotkan < 1 || z.LiczbaSpotkan > w.Liczba() {
			s.blad("%s: wybrane_tygodnie, liczba_spotkan %d spoza 1–%d", gdzie, z.LiczbaSpotkan, w.Liczba())
		}
	}

	if (z.Powiazanie == "") != (z.TypPowiazania == "") {
		s.blad("%s: powiazanie i typ_powiazania muszą być oba puste albo oba ustawione", gdzie)
	}
	if z.TypPowiazania != "" {
		s.wyliczenie(gdzie+": typ_powiazania", z.TypPowiazania, typyPowiazan)
	}
}

// sprawdzPowiazania: zestaw ma co najmniej 2 zajęcia i jeden typ. Zajęcia ten_sam_termin
// muszą dać się postawić w tym samym terminie i sali, więc mają tę samą długość i rozłączne tygodnie.
func (d *Dane) sprawdzPowiazania(s *sprawdzanie, ix *Indeks) {
	for nazwa, zs := range ix.Powiazania {
		if len(zs) < 2 {
			s.blad("powiazanie %s: tylko jedne zajęcia (%s)", nazwa, zs[0].ID)
			continue
		}
		var zajete Tygodnie
		for _, z := range zs {
			if z.TypPowiazania != zs[0].TypPowiazania {
				s.blad("powiazanie %s: różne typy (%s: %s, %s: %s)", nazwa, zs[0].ID, zs[0].TypPowiazania, z.ID, z.TypPowiazania)
			}
			if zs[0].TypPowiazania != TenSamTermin {
				continue
			}
			if z.SlotyNaSpotkanie != zs[0].SlotyNaSpotkanie {
				s.blad("powiazanie %s: ten_sam_termin, a %s i %s mają różną długość", nazwa, zs[0].ID, z.ID)
			}
			if z.DozwoloneTygodnie().Nakladaja(zajete) {
				s.blad("powiazanie %s: ten_sam_termin, a tygodnie %s nakładają się z innymi zajęciami zestawu", nazwa, z.ID)
			}
			zajete |= z.DozwoloneTygodnie()
		}
	}
}

func (d *Dane) sprawdzTerminy(s *sprawdzanie, ix *Indeks) {
	for _, b := range d.BlokadyStale {
		istnieje(s, "blokady_stale: grupa", b.Grupa, ix.Grupy)
		d.sprawdzPrzedzial(s, "blokady_stale "+b.Grupa, b.Dzien, b.SlotOd, b.SlotDo)
	}
	for _, r := range d.RadaWydzialu {
		d.sprawdzPrzedzial(s, "rada_wydzialu", r.Dzien, r.SlotOd, r.SlotDo)
	}
	wykluczenia := make(map[string]int)
	for _, t := range d.DezyderatyTerminy {
		gdzie := "dezyderaty_terminy " + t.Prowadzacy
		istnieje(s, gdzie+": prowadzacy", t.Prowadzacy, ix.Prowadzacy)
		s.wyliczenie(gdzie+": typ", t.Typ, typyDezyderatow)
		s.zakres(gdzie+": dzien", t.Dzien, 1, LiczbaDni)
		s.zakres(gdzie+": slot", t.Slot, 1, LiczbaSlotow)
		if t.Typ == Wykluczony {
			wykluczenia[t.Prowadzacy]++
		}
	}
	for p, n := range wykluczenia {
		if n > limitWykluczen {
			s.blad("dezyderaty_terminy %s: %d godzin wykluczonych (limit %d)", p, n, limitWykluczen)
		}
	}
	for _, o := range d.DezyderatyOgolne {
		istnieje(s, "dezyderaty_ogolne: prowadzacy", o.Prowadzacy, ix.Prowadzacy)
		s.wyliczenie("dezyderaty_ogolne "+o.Prowadzacy+": preferencja_dni", o.PreferencjaDni, preferencjeDni)
	}
	for _, ds := range d.DezyderatySale {
		istnieje(s, "dezyderaty_sale: prowadzacy", ds.Prowadzacy, ix.Prowadzacy)
	}
}

func (d *Dane) sprawdzPrzedzial(s *sprawdzanie, gdzie string, dzien, od, do int) {
	s.zakres(gdzie+": dzien", dzien, 1, LiczbaDni)
	s.zakres(gdzie+": slot_od", od, 1, LiczbaSlotow)
	s.zakres(gdzie+": slot_do", do, od, LiczbaSlotow)
}
