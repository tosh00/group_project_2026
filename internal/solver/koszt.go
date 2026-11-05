package solver

import (
	"math"
	"math/bits"
	"slices"
	"strings"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Pełna funkcja kosztu miękkiego (docs/plan.md, „Funkcja kosztu”), liczona przyrostowo.
// Definicje kryteriów są w docs/contracts/solver.md.
//
// Kryteria globalne liczymy na „jednostkach”:
//   - klasa studentów: studenci z identycznym zbiorem grup (mają dokładnie te same zajęcia),
//     waga = liczba studentów / 25; okienka i przeciążony dzień,
//   - prowadzący: okienka, przerwy > 2 h, preferencja dni, jeden dzień wolny,
//   - rocznik: wykłady w jednym dniu ponad limit.
// Okienka i przeciążenie liczymy w każdym tygodniu osobno i uśredniamy (co_2_tygodnie, Z037/Z038).

const (
	maksSlotowStudenta = 8 // przeciążony dzień: każdy slot ponad 8
	slotowNaDzienProw  = 8 // do celu „skumulowane”: tyle slotów prowadzący może mieć w dniu
	przerwaMaxSlotow   = 2 // „przerwa ≤ 2 h”: każdy slot przerwy ponad 2
	limitWykladowDnia  = 2 // wykłady rocznika w jednym dniu ponad 2
	osobNaJednostke    = 25.0
)

// Nazwy kryteriów (klucze w rozbiciu kosztu, jak tagi JSON w Wagi).
const (
	krOkienkaStudentow    = "okienka_studentow"
	krPrzeciazonyDzien    = "przeciazony_dzien"
	krOkienkaProwadzacych = "okienka_prowadzacych"
	krPo17                = "po_17"
	krPrzed9              = "przed_9"
	krZaDuzaSala          = "za_duza_sala"
	krPreferencjaDni      = "preferencja_dni"
	krNiePo17             = "nie_po_17"
	krDzienWolny          = "dzien_wolny"
	krPrzerwaMax2h        = "przerwa_max_2h"
	krPoraDnia            = "pora_dnia"
	krSaleA               = "sale_a"
	krWykladyRocznika     = "wyklady_rocznika"
)

// oblozenie jednostki: liczniki zajęć w (dzień, tydzień, slot) i maski zajętych slotów.
type oblozenie struct {
	licz  [model.LiczbaDni + 1][model.LiczbaTygodni][model.LiczbaSlotow]int16
	maska [model.LiczbaDni + 1][model.LiczbaTygodni]uint16
}

func (o *oblozenie) zmien(dzien, start, dl int, tyg model.Tygodnie, delta int16) {
	for t := 0; t < model.LiczbaTygodni; t++ {
		if tyg&(1<<t) == 0 {
			continue
		}
		for s := start - 1; s < start-1+dl; s++ {
			o.licz[dzien][t][s] += delta
			if o.licz[dzien][t][s] > 0 {
				o.maska[dzien][t] |= 1 << s
			} else {
				o.maska[dzien][t] &^= 1 << s
			}
		}
	}
}

func (o *oblozenie) dzienUzyty(dzien int) bool {
	for t := 0; t < model.LiczbaTygodni; t++ {
		if o.maska[dzien][t] != 0 {
			return true
		}
	}
	return false
}

// okienka: wolne sloty między pierwszym a ostatnim zajętym.
func okienka(m uint16) int {
	if m == 0 {
		return 0
	}
	lo, hi := bits.TrailingZeros16(m), 15-bits.LeadingZeros16(m)
	return hi - lo + 1 - bits.OnesCount16(m)
}

// przerwyPonad: suma (długość przerwy − limit) po przerwach dłuższych niż limit.
func przerwyPonad(m uint16, limit int) int {
	if m == 0 {
		return 0
	}
	lo, hi := bits.TrailingZeros16(m), 15-bits.LeadingZeros16(m)
	suma, przerwa := 0, 0
	for s := lo; s <= hi; s++ {
		if m&(1<<s) != 0 {
			suma += max(0, przerwa-limit)
			przerwa = 0
		} else {
			przerwa++
		}
	}
	return suma
}

type klasa struct {
	waga float64 // liczba studentów / 25
}

type prowadzacyK struct {
	id                  string
	uwagi               uwagiProwadzacego
	preferencja         model.PreferencjaDni
	celSkum, celNieskum int
}

// udzial to jedne zajęcia węzła: które jednostki obciąża.
type udzial struct {
	czlonek    int // indeks w wezel.zajecia
	dl         int
	klasy      []int
	prowadzacy []int
	rocznik    int // −1, gdy to nie wykład
}

// koszt trzyma obłożenie jednostek i składniki kosztu, żeby liczyć zmianę przy ruchu.
type koszt struct {
	pr   *problem
	st   *stan
	wagi *Wagi

	klasy      []klasa
	prowadzacy []prowadzacyK
	roczniki   []string
	udzialy    [][]udzial // węzeł → jego zajęcia

	obK  []oblozenie
	obP  []oblozenie
	wykR [][model.LiczbaDni + 1]int // rocznik × dzień → liczba wykładów

	// Składniki (pamiętane, żeby liczyć różnicę).
	kK   [][model.LiczbaDni + 1]float64 // klasa × dzień
	kP   [][model.LiczbaDni + 1]float64 // prowadzący × dzień
	kPG  []float64                      // prowadzący: preferencja dni i dzień wolny
	kR   [][model.LiczbaDni + 1]float64 // rocznik × dzień
	kL   []float64                      // węzeł: kryteria lokalne
	suma float64

	// Dziennik zmian do cofnięcia ruchu i znaczniki, żeby nie liczyć składnika dwa razy.
	dziennik []wpisDziennika
	znacznik int
	znK      [][model.LiczbaDni + 1]int
	znP      [][model.LiczbaDni + 1]int
	znPG     []int
	znR      [][model.LiczbaDni + 1]int
	znL      []int
}

type rodzajSkladnika uint8

const (
	skK rodzajSkladnika = iota
	skP
	skPG
	skR
	skL
)

type wpisDziennika struct {
	rodzaj        rodzajSkladnika
	jednostka, dz int
	stara         float64
}

// nowyKoszt buduje strukturę kosztu dla pełnego stanu i liczy koszt od zera.
func nowyKoszt(pr *problem, st *stan, wagi *Wagi) *koszt {
	k := &koszt{pr: pr, st: st, wagi: wagi}
	k.zbudujJednostki()
	for _, w := range pr.wezly {
		if st.ok[w.id] {
			k.oblozWezel(w.id, 1)
		}
	}
	k.przeliczWszystko()
	return k
}

func (k *koszt) zbudujJednostki() {
	pr := k.pr
	// Klasy studentów: zbiór grup studenta → klasa.
	grupyStudenta := make(map[string][]string)
	for _, sg := range pr.dane.StudenciGrupy {
		grupyStudenta[sg.Student] = append(grupyStudenta[sg.Student], sg.Grupa)
	}
	studenci := make([]string, 0, len(grupyStudenta))
	for s := range grupyStudenta {
		studenci = append(studenci, s)
	}
	slices.Sort(studenci) // stała kolejność klas (powtarzalność)
	numerKlasy := make(map[string]int)
	var grupyKlasy [][]string
	var liczebnosc []int
	for _, s := range studenci {
		gs := grupyStudenta[s]
		slices.Sort(gs)
		klucz := strings.Join(gs, ",")
		n, ok := numerKlasy[klucz]
		if !ok {
			n = len(grupyKlasy)
			numerKlasy[klucz] = n
			grupyKlasy = append(grupyKlasy, gs)
			liczebnosc = append(liczebnosc, 0)
		}
		liczebnosc[n]++
	}
	for _, n := range liczebnosc {
		k.klasy = append(k.klasy, klasa{waga: float64(n) / osobNaJednostke})
	}

	numerProw := make(map[string]int)
	tygodniowo := make(map[string]float64)
	liczbaZajec := make(map[string]int)
	for i := range pr.dane.Zajecia {
		z := &pr.dane.Zajecia[i]
		for _, p := range z.Prowadzacy {
			tygodniowo[p] += float64(z.SlotyNaSpotkanie*z.LiczbaSpotkan) / model.LiczbaTygodni
			liczbaZajec[p]++
		}
	}
	for _, p := range pr.dane.Prowadzacy {
		numerProw[p.ID] = len(k.prowadzacy)
		pk := prowadzacyK{
			id:         p.ID,
			uwagi:      pr.uwagi(p.ID),
			celSkum:    max(1, int(math.Ceil(tygodniowo[p.ID]/slotowNaDzienProw))),
			celNieskum: min(model.LiczbaDni, liczbaZajec[p.ID]),
		}
		if o := pr.ix.DezyderatyOgolne[p.ID]; o != nil {
			pk.preferencja = o.PreferencjaDni
		}
		k.prowadzacy = append(k.prowadzacy, pk)
	}

	numerRocznika := make(map[string]int)
	for _, r := range pr.dane.Roczniki {
		numerRocznika[r.ID] = len(k.roczniki)
		k.roczniki = append(k.roczniki, r.ID)
	}

	k.udzialy = make([][]udzial, len(pr.wezly))
	for _, w := range pr.wezly {
		for i, z := range w.zajecia {
			u := udzial{czlonek: i, dl: z.SlotyNaSpotkanie, rocznik: -1}
			for n, gs := range grupyKlasy {
				if slices.ContainsFunc(z.Grupy, func(g string) bool { return slices.Contains(gs, g) }) {
					u.klasy = append(u.klasy, n)
				}
			}
			for _, p := range z.Prowadzacy {
				if n, ok := numerProw[p]; ok {
					u.prowadzacy = append(u.prowadzacy, n)
				}
			}
			if r, ok := numerRocznika[z.Rocznik]; ok && z.Forma == model.Wyklad {
				u.rocznik = r
			}
			k.udzialy[w.id] = append(k.udzialy[w.id], u)
		}
	}

	nK, nP, nR, nW := len(k.klasy), len(k.prowadzacy), len(k.roczniki), len(pr.wezly)
	k.obK, k.obP = make([]oblozenie, nK), make([]oblozenie, nP)
	k.wykR = make([][model.LiczbaDni + 1]int, nR)
	k.kK, k.kP, k.kPG = make([][model.LiczbaDni + 1]float64, nK), make([][model.LiczbaDni + 1]float64, nP), make([]float64, nP)
	k.kR, k.kL = make([][model.LiczbaDni + 1]float64, nR), make([]float64, nW)
	k.znK, k.znP, k.znPG = make([][model.LiczbaDni + 1]int, nK), make([][model.LiczbaDni + 1]int, nP), make([]int, nP)
	k.znR, k.znL = make([][model.LiczbaDni + 1]int, nR), make([]int, nW)
}

// oblozWezel dodaje (delta = 1) albo zdejmuje (delta = −1) obecne przypisanie węzła z obłożenia jednostek.
func (k *koszt) oblozWezel(w int, delta int16) {
	kd := k.st.przyp[w]
	war := k.pr.wezly[w].warianty[kd.wariant]
	for _, u := range k.udzialy[w] {
		tyg := war.tygodnie[u.czlonek]
		for _, c := range u.klasy {
			k.obK[c].zmien(kd.dzien, kd.start, u.dl, tyg, delta)
		}
		for _, p := range u.prowadzacy {
			k.obP[p].zmien(kd.dzien, kd.start, u.dl, tyg, delta)
		}
		if u.rocznik >= 0 {
			k.wykR[u.rocznik][kd.dzien] += int(delta)
		}
	}
}

func (k *koszt) wartoscK(c, d int) float64 {
	ob := &k.obK[c]
	var s float64
	for t := 0; t < model.LiczbaTygodni; t++ {
		m := ob.maska[d][t]
		s += k.wagi.OkienkaStudentow*float64(okienka(m)) +
			k.wagi.PrzeciazonyDzien*float64(max(0, bits.OnesCount16(m)-maksSlotowStudenta))
	}
	return s * k.klasy[c].waga / model.LiczbaTygodni
}

func (k *koszt) wartoscP(p, d int) float64 {
	ob := &k.obP[p]
	przerwa := k.prowadzacy[p].uwagi.przerwaMax2h
	var s float64
	for t := 0; t < model.LiczbaTygodni; t++ {
		m := ob.maska[d][t]
		s += k.wagi.OkienkaProwadzacych * float64(okienka(m))
		if przerwa {
			s += k.wagi.PrzerwaMax2h * float64(przerwyPonad(m, przerwaMaxSlotow))
		}
	}
	return s / model.LiczbaTygodni
}

func (k *koszt) liczbaDniProw(p int) int {
	n := 0
	for d := 1; d <= model.LiczbaDni; d++ {
		if k.obP[p].dzienUzyty(d) {
			n++
		}
	}
	return n
}

func (k *koszt) wartoscPG(p int) float64 {
	pk := &k.prowadzacy[p]
	dni := k.liczbaDniProw(p)
	var s float64
	switch pk.preferencja {
	case model.Skumulowane:
		s += k.wagi.PreferencjaDni * float64(max(0, dni-pk.celSkum))
	case model.Nieskumulowane:
		s += k.wagi.PreferencjaDni * float64(max(0, pk.celNieskum-dni))
	}
	if pk.uwagi.dzienWolny && dni == model.LiczbaDni {
		s += k.wagi.DzienWolny
	}
	return s
}

func (k *koszt) wartoscR(r, d int) float64 {
	return k.wagi.WykladyRocznika * float64(max(0, k.wykR[r][d]-limitWykladowDnia))
}

func (k *koszt) wartoscL(w int) float64 {
	return k.pr.kosztLokalny(k.pr.wezly[w], k.st.przyp[w], k.wagi)
}

// przeliczWszystko liczy wszystkie składniki od zera.
func (k *koszt) przeliczWszystko() {
	k.suma = 0
	for d := 1; d <= model.LiczbaDni; d++ {
		for c := range k.klasy {
			k.kK[c][d] = k.wartoscK(c, d)
			k.suma += k.kK[c][d]
		}
		for p := range k.prowadzacy {
			k.kP[p][d] = k.wartoscP(p, d)
			k.suma += k.kP[p][d]
		}
		for r := range k.roczniki {
			k.kR[r][d] = k.wartoscR(r, d)
			k.suma += k.kR[r][d]
		}
	}
	for p := range k.prowadzacy {
		k.kPG[p] = k.wartoscPG(p)
		k.suma += k.kPG[p]
	}
	for w := range k.pr.wezly {
		k.kL[w] = 0
		if k.st.ok[w] {
			k.kL[w] = k.wartoscL(w)
		}
		k.suma += k.kL[w]
	}
}

// Ruch: zmienWezel przestawia węzeł (stan i obłożenie) i zaznacza składniki do przeliczenia.
// Po wszystkich zmianach ruchu przelicz() aktualizuje sumę; cofnij() przywraca stan sprzed ruchu.

func (k *koszt) poczatekRuchu() {
	k.dziennik = k.dziennik[:0]
	k.znacznik++
}

// zaznacz dopisuje do dziennika składniki zależne od obecnego przypisania węzła.
func (k *koszt) zaznacz(w int) {
	d := k.st.przyp[w].dzien
	for _, u := range k.udzialy[w] {
		for _, c := range u.klasy {
			if k.znK[c][d] != k.znacznik {
				k.znK[c][d] = k.znacznik
				k.dziennik = append(k.dziennik, wpisDziennika{skK, c, d, k.kK[c][d]})
			}
		}
		for _, p := range u.prowadzacy {
			if k.znP[p][d] != k.znacznik {
				k.znP[p][d] = k.znacznik
				k.dziennik = append(k.dziennik, wpisDziennika{skP, p, d, k.kP[p][d]})
			}
			if k.znPG[p] != k.znacznik {
				k.znPG[p] = k.znacznik
				k.dziennik = append(k.dziennik, wpisDziennika{skPG, p, 0, k.kPG[p]})
			}
		}
		if u.rocznik >= 0 && k.znR[u.rocznik][d] != k.znacznik {
			k.znR[u.rocznik][d] = k.znacznik
			k.dziennik = append(k.dziennik, wpisDziennika{skR, u.rocznik, d, k.kR[u.rocznik][d]})
		}
	}
	if k.znL[w] != k.znacznik {
		k.znL[w] = k.znacznik
		k.dziennik = append(k.dziennik, wpisDziennika{skL, w, 0, k.kL[w]})
	}
}

// zmienWezel przestawia przypisany węzeł na kandydata kd.
func (k *koszt) zmienWezel(w int, kd kandydat) {
	k.zaznacz(w)
	k.oblozWezel(w, -1)
	k.st.przypisz(w, kd)
	k.oblozWezel(w, 1)
	k.zaznacz(w)
}

// przelicz liczy na nowo składniki z dziennika i zwraca zmianę kosztu.
func (k *koszt) przelicz() float64 {
	var delta float64
	for _, e := range k.dziennik {
		var nowa float64
		switch e.rodzaj {
		case skK:
			nowa = k.wartoscK(e.jednostka, e.dz)
			k.kK[e.jednostka][e.dz] = nowa
		case skP:
			nowa = k.wartoscP(e.jednostka, e.dz)
			k.kP[e.jednostka][e.dz] = nowa
		case skPG:
			nowa = k.wartoscPG(e.jednostka)
			k.kPG[e.jednostka] = nowa
		case skR:
			nowa = k.wartoscR(e.jednostka, e.dz)
			k.kR[e.jednostka][e.dz] = nowa
		case skL:
			nowa = k.wartoscL(e.jednostka)
			k.kL[e.jednostka] = nowa
		}
		delta += nowa - e.stara
	}
	k.suma += delta
	return delta
}

// cofnijSkladniki przywraca składniki z dziennika i sumę (stan i obłożenie cofa wywołujący).
func (k *koszt) cofnijSkladniki(delta float64) {
	for _, e := range k.dziennik {
		switch e.rodzaj {
		case skK:
			k.kK[e.jednostka][e.dz] = e.stara
		case skP:
			k.kP[e.jednostka][e.dz] = e.stara
		case skPG:
			k.kPG[e.jednostka] = e.stara
		case skR:
			k.kR[e.jednostka][e.dz] = e.stara
		case skL:
			k.kL[e.jednostka] = e.stara
		}
	}
	k.suma -= delta
}

// rozbicie zwraca koszt od zera z podziałem na kryteria (do raportu i testów).
func (pr *problem) rozbicieKosztu(st *stan, wagi *Wagi) map[string]float64 {
	// Każde kryterium osobno: wagi z jednym niezerowym polem.
	kryteria := []struct {
		nazwa string
		pole  func(*Wagi) *float64
	}{
		{krOkienkaStudentow, func(w *Wagi) *float64 { return &w.OkienkaStudentow }},
		{krPrzeciazonyDzien, func(w *Wagi) *float64 { return &w.PrzeciazonyDzien }},
		{krOkienkaProwadzacych, func(w *Wagi) *float64 { return &w.OkienkaProwadzacych }},
		{krPo17, func(w *Wagi) *float64 { return &w.Po17 }},
		{krPrzed9, func(w *Wagi) *float64 { return &w.Przed9 }},
		{krZaDuzaSala, func(w *Wagi) *float64 { return &w.ZaDuzaSala }},
		{krPreferencjaDni, func(w *Wagi) *float64 { return &w.PreferencjaDni }},
		{krNiePo17, func(w *Wagi) *float64 { return &w.NiePo17 }},
		{krDzienWolny, func(w *Wagi) *float64 { return &w.DzienWolny }},
		{krPrzerwaMax2h, func(w *Wagi) *float64 { return &w.PrzerwaMax2h }},
		{krPoraDnia, func(w *Wagi) *float64 { return &w.PoraDnia }},
		{krSaleA, func(w *Wagi) *float64 { return &w.SaleA }},
		{krWykladyRocznika, func(w *Wagi) *float64 { return &w.WykladyRocznika }},
	}
	out := make(map[string]float64, len(kryteria))
	for _, kr := range kryteria {
		var jedno Wagi
		*kr.pole(&jedno) = *kr.pole(wagi)
		out[kr.nazwa] = nowyKoszt(pr, st, &jedno).suma
	}
	return out
}
