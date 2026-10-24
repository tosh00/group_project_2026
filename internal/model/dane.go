// Package model zawiera struktury danych wejściowych (data/dane_planu.json),
// ich import i sprawdzenie spójności oraz format planu wyjściowego (plan.json).
//
// To kontrakt dla solvera, walidatora i backendu. Zmiany tylko przez PR
// z review wszystkich zainteresowanych (patrz docs/contracts/).
//
// Konwencja dla pól, które w JSON mogą być null: tekst to "", liczba to 0.
// Dni, sloty i tygodnie numerujemy od 1, więc 0 zawsze oznacza „brak”.
package model

// Stałe siatki czasu (data/README.md, sekcja 3).
const (
	LiczbaDni      = 5  // dni planu 1–5 (pon–pt)
	LiczbaSlotow   = 12 // sloty 1–12 po 45 minut
	LiczbaTygodni  = 14 // każdy dzień planu odbywa się 14 razy
	limitWykluczen = 12 // maks. godzin wykluczonych na prowadzącego (dezyderaty)
)

// Dane to wszystkie tabele z dane_planu.json.
type Dane struct {
	Meta              Meta              `json:"meta"`
	Kalendarz         []DzienKalendarza `json:"kalendarz"`
	Sloty             []Slot            `json:"sloty"`
	Sale              []Sala            `json:"sale"`
	Prowadzacy        []Prowadzacy      `json:"prowadzacy"`
	Roczniki          []Rocznik         `json:"roczniki"`
	Grupy             []Grupa           `json:"grupy"`
	Studenci          []Student         `json:"studenci"`
	StudenciGrupy     []StudentGrupa    `json:"studenci_grupy"`
	KolizjeGrup       []KolizjaGrup     `json:"kolizje_grup"`
	Przedmioty        []Przedmiot       `json:"przedmioty"`
	Zajecia           []Zajecia         `json:"zajecia"`
	BlokadyStale      []BlokadaStala    `json:"blokady_stale"`
	RadaWydzialu      []TerminRady      `json:"rada_wydzialu"`
	DezyderatyTerminy []DezyderatTermin `json:"dezyderaty_terminy"`
	DezyderatyOgolne  []DezyderatOgolny `json:"dezyderaty_ogolne"`
	DezyderatySale    []DezyderatSali   `json:"dezyderaty_sale"`
}

type Meta struct {
	Opis   string            `json:"opis"`
	Wersja string            `json:"wersja"`
	Data   string            `json:"data"`
	Tabele map[string]string `json:"tabele"`
}

// DzienKalendarza to jeden dzień semestru. Dla dni wolnych PlanDnia i NrTygodnia są 0.
type DzienKalendarza struct {
	Data          string  `json:"data"` // RRRR-MM-DD
	DzienTygodnia string  `json:"dzien_tygodnia"`
	TypDnia       TypDnia `json:"typ_dnia"`
	PlanDnia      int     `json:"plan_dnia"`   // 1–5: wg planu którego dnia odbywają się zajęcia
	NrTygodnia    int     `json:"nr_tygodnia"` // 1–14: kolejne wystąpienie tego dnia planu
	Opis          string  `json:"opis"`
}

type TypDnia string

const (
	DzienZajec        TypDnia = "zajęcia"
	DzienWolny        TypDnia = "wolne"
	DzienOdpracowania TypDnia = "odpracowanie"
)

type Slot struct {
	Slot int    `json:"slot"`
	Od   string `json:"od"` // "8:15"
	Do   string `json:"do"` // "9:00"
}

// Sala. Pojemnosc = 0 oznacza brak danych (tylko sale niedostępne do planowania).
type Sala struct {
	ID                   string   `json:"id"`
	Nazwa                string   `json:"nazwa"`
	Budynek              string   `json:"budynek"`
	Typ                  TypSali  `json:"typ"`
	Pojemnosc            int      `json:"pojemnosc"`
	PojemnoscSzacowana   bool     `json:"pojemnosc_szacowana"`
	MetrazM2             float64  `json:"metraz_m2"`
	Wyposazenie          []string `json:"wyposazenie"`
	Przeznaczenie        string   `json:"przeznaczenie"`
	DostepnaDoPlanowania bool     `json:"dostepna_do_planowania"`
	Uwagi                string   `json:"uwagi"`
}

type TypSali string

const (
	Aula                     TypSali = "aula"
	SalaDydaktyczna          TypSali = "sala_dydaktyczna"
	PracowniaKomputerowa     TypSali = "pracownia_komputerowa"
	PracowniaRobotyki        TypSali = "pracownia_robotyki"
	PracowniaSpecjalistyczna TypSali = "pracownia_specjalistyczna"
)

type Prowadzacy struct {
	ID                  string `json:"id"`
	Stopien             string `json:"stopien"`
	Nazwa               string `json:"nazwa"`
	Jednostka           string `json:"jednostka"`
	CzlonekRadyWydzialu bool   `json:"czlonek_rady_wydzialu"`
}

type Rocznik struct {
	ID              string `json:"id"`
	Kierunek        string `json:"kierunek"`
	StopienStudiow  string `json:"stopien_studiow"`
	Rok             int    `json:"rok"`
	Semestr         int    `json:"semestr"`
	Nazwa           string `json:"nazwa"`
	LiczbaStudentow int    `json:"liczba_studentow"`
	Syntetyczny     bool   `json:"syntetyczny"`
}

// Grupa. O kolizjach decyduje StudenciGrupy (wspólni studenci), nie Podzial ani GrupaNadrzedna.
type Grupa struct {
	ID              string `json:"id"`
	Rocznik         string `json:"rocznik"`
	Nazwa           string `json:"nazwa"`
	Typ             string `json:"typ"`
	Podzial         string `json:"podzial"`
	GrupaNadrzedna  string `json:"grupa_nadrzedna"`
	LiczbaStudentow int    `json:"liczba_studentow"`
}

type Student struct {
	ID      string `json:"id"`
	Rocznik string `json:"rocznik"`
}

type StudentGrupa struct {
	Student string `json:"student"`
	Grupa   string `json:"grupa"`
}

type KolizjaGrup struct {
	GrupaA          string `json:"grupa_a"`
	GrupaB          string `json:"grupa_b"`
	WspolniStudenci int    `json:"wspolni_studenci"`
}

type Przedmiot struct {
	ID      string `json:"id"`
	Rocznik string `json:"rocznik"`
	Nazwa   string `json:"nazwa"`
}

// Zajecia to jeden cykl spotkań do zaplanowania (data/README.md, sekcja 4).
type Zajecia struct {
	ID                  string        `json:"id"`
	Rocznik             string        `json:"rocznik"`
	PrzedmiotID         string        `json:"przedmiot_id"`
	Przedmiot           string        `json:"przedmiot"`
	Forma               Forma         `json:"forma"`
	FormaOpis           string        `json:"forma_opis"`
	Grupy               []string      `json:"grupy"`
	LiczbaStudentow     int           `json:"liczba_studentow"`
	LaczoneZ            string        `json:"laczone_z"`
	DodatkowiStudenci   int           `json:"dodatkowi_studenci"`
	Prowadzacy          []string      `json:"prowadzacy"` // wszyscy muszą być wolni
	GodzinyProgramowe   int           `json:"godziny_programowe"`
	LiczbaSpotkan       int           `json:"liczba_spotkan"`
	SlotyNaSpotkanie    int           `json:"sloty_na_spotkanie"`
	Czestotliwosc       Czestotliwosc `json:"czestotliwosc"`
	TydzienOd           int           `json:"tydzien_od"`
	TydzienDo           int           `json:"tydzien_do"`
	TypySal             []TypSali     `json:"typy_sal"`       // gdy DozwoloneSale puste
	DozwoloneSale       []string      `json:"dozwolone_sale"` // jeśli niepuste, tylko te sale
	WymaganeWyposazenie []string      `json:"wymagane_wyposazenie"`
	Tryb                Tryb          `json:"tryb"`
	Powiazanie          string        `json:"powiazanie"`
	TypPowiazania       TypPowiazania `json:"typ_powiazania"`
	TygodnieNiedostepne []int         `json:"tygodnie_niedostepne"`
	Uwagi               string        `json:"uwagi"`
}

// LiczbaOsob to liczba miejsc potrzebnych w sali (reguła twarda 7).
func (z *Zajecia) LiczbaOsob() int { return z.LiczbaStudentow + z.DodatkowiStudenci }

// DozwoloneTygodnie to tygodnie z zakresu TydzienOd–TydzienDo bez TygodnieNiedostepne.
func (z *Zajecia) DozwoloneTygodnie() Tygodnie {
	return Zakres(z.TydzienOd, z.TydzienDo).Bez(TygodnieZListy(z.TygodnieNiedostepne))
}

type Forma string

const (
	Wyklad           Forma = "W"
	Konwersatorium   Forma = "K"
	LabInformatyczne Forma = "LI"
	Pracownia        Forma = "P"
	Lektorat         Forma = "LEK"
	Seminarium       Forma = "SEM"
)

type Czestotliwosc string

const (
	CoTydzien       Czestotliwosc = "co_tydzien"
	Co2Tygodnie     Czestotliwosc = "co_2_tygodnie"
	WybraneTygodnie Czestotliwosc = "wybrane_tygodnie"
)

type Tryb string

const (
	Stacjonarny Tryb = "stacjonarny"
	Hybrydowy   Tryb = "hybrydowy" // też potrzebuje sali
)

// TypPowiazania łączy zajęcia o tym samym Powiazanie. Pusty = brak powiązania.
type TypPowiazania string

const (
	RozneDni     TypPowiazania = "rozne_dni"      // spotkania w różnych dniach
	TenSamTermin TypPowiazania = "ten_sam_termin" // ten sam dzień, slot i sala
)

// BlokadaStala to termin grupy ustalony poza wydziałem (np. WF), w każdym tygodniu.
type BlokadaStala struct {
	Grupa  string `json:"grupa"`
	Dzien  int    `json:"dzien"`
	SlotOd int    `json:"slot_od"`
	SlotDo int    `json:"slot_do"`
	Opis   string `json:"opis"`
}

// TerminRady blokuje prowadzących z CzlonekRadyWydzialu = true.
type TerminRady struct {
	Dzien  int    `json:"dzien"`
	SlotOd int    `json:"slot_od"`
	SlotDo int    `json:"slot_do"`
	Opis   string `json:"opis"`
}

// DezyderatTermin to jedna godzina (dzień + slot), w której prowadzący nie może prowadzić zajęć.
type DezyderatTermin struct {
	Prowadzacy string        `json:"prowadzacy"`
	Dzien      int           `json:"dzien"`
	Slot       int           `json:"slot"`
	Typ        TypDezyderatu `json:"typ"`
}

type TypDezyderatu string

const (
	Wykluczony    TypDezyderatu = "wykluczony"
	DezyderatRady TypDezyderatu = "rada_wydzialu"
)

type DezyderatOgolny struct {
	Prowadzacy               string         `json:"prowadzacy"`
	LiczbaGodzinWykluczonych int            `json:"liczba_godzin_wykluczonych"`
	PowodWykluczen           string         `json:"powod_wykluczen"`
	PreferencjaDni           PreferencjaDni `json:"preferencja_dni"`
	InneUwagi                string         `json:"inne_uwagi"`
}

type PreferencjaDni string

const (
	Skumulowane     PreferencjaDni = "skumulowane"
	Nieskumulowane  PreferencjaDni = "nieskumulowane"
	BrakPreferencji PreferencjaDni = "brak"
)

// DezyderatSali ma charakter informacyjny: te same wymagania są już w Zajecia.DozwoloneSale / WymaganeWyposazenie.
type DezyderatSali struct {
	Prowadzacy   string `json:"prowadzacy"`
	Przedmiot    string `json:"przedmiot"`
	Forma        Forma  `json:"forma"`
	RodzajSaliNr string `json:"rodzaj_sali_nr"`
	Sprzet       string `json:"sprzet"`
}
