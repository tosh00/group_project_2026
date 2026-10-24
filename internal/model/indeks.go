package model

// Indeks daje szybki dostęp do danych po identyfikatorach. Budujemy go raz po wczytaniu.
// Wskaźniki prowadzą do elementów tabel w Dane, więc Dane nie wolno potem modyfikować.
type Indeks struct {
	Zajecia    map[string]*Zajecia
	Sale       map[string]*Sala
	Prowadzacy map[string]*Prowadzacy
	Grupy      map[string]*Grupa
	Roczniki   map[string]*Rocznik
	Przedmioty map[string]*Przedmiot

	// StudenciGrupy: grupa → zbiór jej studentów.
	StudenciGrupy map[string]map[string]bool
	// DezyderatyOgolne: prowadzący → jego dezyderat ogólny.
	DezyderatyOgolne map[string]*DezyderatOgolny
	// Powiazania: wartość Zajecia.Powiazanie → zajęcia w tym zestawie (w kolejności z danych).
	Powiazania map[string][]*Zajecia

	// kolizje: pary grup ze wspólnymi studentami, liczone ze StudenciGrupy.
	kolizje map[paraGrup]int
}

type paraGrup struct{ a, b string }

func nowaPara(a, b string) paraGrup {
	if a > b {
		a, b = b, a
	}
	return paraGrup{a, b}
}

// NowyIndeks buduje indeks. Kolizje grup liczy ze studenci_grupy, które są źródłem prawdy (data/README.md).
func NowyIndeks(d *Dane) *Indeks {
	ix := &Indeks{
		Zajecia:          make(map[string]*Zajecia, len(d.Zajecia)),
		Sale:             make(map[string]*Sala, len(d.Sale)),
		Prowadzacy:       make(map[string]*Prowadzacy, len(d.Prowadzacy)),
		Grupy:            make(map[string]*Grupa, len(d.Grupy)),
		Roczniki:         make(map[string]*Rocznik, len(d.Roczniki)),
		Przedmioty:       make(map[string]*Przedmiot, len(d.Przedmioty)),
		StudenciGrupy:    make(map[string]map[string]bool, len(d.Grupy)),
		DezyderatyOgolne: make(map[string]*DezyderatOgolny, len(d.DezyderatyOgolne)),
		Powiazania:       make(map[string][]*Zajecia),
	}
	for i := range d.Zajecia {
		z := &d.Zajecia[i]
		ix.Zajecia[z.ID] = z
		if z.Powiazanie != "" {
			ix.Powiazania[z.Powiazanie] = append(ix.Powiazania[z.Powiazanie], z)
		}
	}
	for i := range d.Sale {
		ix.Sale[d.Sale[i].ID] = &d.Sale[i]
	}
	for i := range d.Prowadzacy {
		ix.Prowadzacy[d.Prowadzacy[i].ID] = &d.Prowadzacy[i]
	}
	for i := range d.Grupy {
		ix.Grupy[d.Grupy[i].ID] = &d.Grupy[i]
	}
	for i := range d.Roczniki {
		ix.Roczniki[d.Roczniki[i].ID] = &d.Roczniki[i]
	}
	for i := range d.Przedmioty {
		ix.Przedmioty[d.Przedmioty[i].ID] = &d.Przedmioty[i]
	}
	for i := range d.DezyderatyOgolne {
		ix.DezyderatyOgolne[d.DezyderatyOgolne[i].Prowadzacy] = &d.DezyderatyOgolne[i]
	}
	for _, sg := range d.StudenciGrupy {
		if ix.StudenciGrupy[sg.Grupa] == nil {
			ix.StudenciGrupy[sg.Grupa] = make(map[string]bool)
		}
		ix.StudenciGrupy[sg.Grupa][sg.Student] = true
	}
	ix.kolizje = policzKolizje(ix.StudenciGrupy)
	return ix
}

func policzKolizje(studenci map[string]map[string]bool) map[paraGrup]int {
	grupyStudenta := make(map[string][]string)
	for g, ss := range studenci {
		for s := range ss {
			grupyStudenta[s] = append(grupyStudenta[s], g)
		}
	}
	kolizje := make(map[paraGrup]int)
	for _, gs := range grupyStudenta {
		for i := range gs {
			for j := i + 1; j < len(gs); j++ {
				kolizje[nowaPara(gs[i], gs[j])]++
			}
		}
	}
	return kolizje
}

// WspolniStudenci zwraca liczbę studentów należących do obu grup.
// Dla tej samej grupy zwraca jej liczebność.
func (ix *Indeks) WspolniStudenci(a, b string) int {
	if a == b {
		return len(ix.StudenciGrupy[a])
	}
	return ix.kolizje[nowaPara(a, b)]
}

// GrupyKoliduja mówi, czy dwie grupy mają choć jednego wspólnego studenta (reguła twarda 5).
func (ix *Indeks) GrupyKoliduja(a, b string) bool { return ix.WspolniStudenci(a, b) > 0 }

// ZajeciaKolidujaStudenci mówi, czy jakiś student uczestniczy w obu zajęciach.
func (ix *Indeks) ZajeciaKolidujaStudenci(z1, z2 *Zajecia) bool {
	for _, a := range z1.Grupy {
		for _, b := range z2.Grupy {
			if ix.GrupyKoliduja(a, b) {
				return true
			}
		}
	}
	return false
}

// WspolnyProwadzacy mówi, czy zajęcia mają choć jednego wspólnego prowadzącego.
func WspolnyProwadzacy(z1, z2 *Zajecia) bool {
	for _, a := range z1.Prowadzacy {
		for _, b := range z2.Prowadzacy {
			if a == b {
				return true
			}
		}
	}
	return false
}
