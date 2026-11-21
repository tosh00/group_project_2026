package api

import (
	"encoding/json"
	"maps"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/tosh00/group_project_2026/internal/model"
)

// Test kontraktu: api/openapi.yaml (z którego frontend generuje typy) musi opisywać dokładnie te pola,
// które backend wysyła i przyjmuje. Bez bazy i bez serwera, więc działa też bez TEST_DATABASE_URL.

const (
	plikOpenAPI = "../../api/openapi.yaml"
	plikDanych  = "../../data/dane_planu.json"
	prefiksRef  = "#/components/schemas/"
)

type schemat struct {
	Type       any                 `yaml:"type"`
	Default    any                 `yaml:"default"`
	Required   []string            `yaml:"required"`
	Ref        string              `yaml:"$ref"`
	Properties map[string]*schemat `yaml:"properties"`
	Items      *schemat            `yaml:"items"`
}

// dopuszczaNull mówi, czy `type` schematu zawiera "null" (np. `type: [string, "null"]`).
func (s *schemat) dopuszczaNull() bool {
	switch t := s.Type.(type) {
	case string:
		return t == "null"
	case []any:
		return slices.Contains(t, any("null"))
	}
	return false
}

type schematy map[string]*schemat

func wczytajSchematy(t *testing.T) schematy {
	t.Helper()
	b, err := os.ReadFile(plikOpenAPI)
	if err != nil {
		t.Fatal(err)
	}
	var dok struct {
		Components struct {
			Schemas schematy `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(b, &dok); err != nil {
		t.Fatalf("%s: %v", plikOpenAPI, err)
	}
	return dok.Components.Schemas
}

// rozwin zamienia `$ref` na wskazywany schemat.
func (sch schematy) rozwin(t *testing.T, s *schemat) *schemat {
	t.Helper()
	if s == nil || s.Ref == "" {
		return s
	}
	cel, ok := sch[strings.TrimPrefix(s.Ref, prefiksRef)]
	if !ok {
		t.Fatalf("nieznany $ref %q", s.Ref)
	}
	return cel
}

// porownaj sprawdza, że pola struktury Go (tagi json) i `properties` schematu to te same zbiory,
// i schodzi rekurencyjnie do zagnieżdżonych struktur oraz list struktur.
func (sch schematy) porownaj(t *testing.T, sciezka string, typ reflect.Type, s *schemat) {
	t.Helper()
	s = sch.rozwin(t, s)
	pola := polaJSON(typ)

	var wGo, wSchemacie []string
	for nazwa := range pola {
		if _, ok := s.Properties[nazwa]; !ok {
			wGo = append(wGo, nazwa)
		}
	}
	for nazwa := range s.Properties {
		if _, ok := pola[nazwa]; !ok {
			wSchemacie = append(wSchemacie, nazwa)
		}
	}
	sort.Strings(wGo)
	sort.Strings(wSchemacie)
	if len(wGo) > 0 {
		t.Errorf("%s: pola z Go brakujące w openapi.yaml: %v", sciezka, wGo)
	}
	if len(wSchemacie) > 0 {
		t.Errorf("%s: pola z openapi.yaml, których nie ma w Go: %v", sciezka, wSchemacie)
	}

	// Pole z `omitempty` może nie przyjść, więc nie może być wymagane. Nie może też mieć `default`:
	// openapi-typescript robi z takiego pola wymagane, a frontend zakładałby, że zawsze jest.
	for nazwa := range polaOmitempty(typ) {
		if slices.Contains(s.Required, nazwa) {
			t.Errorf("%s.%s: Go pomija puste pole (omitempty), a openapi.yaml ma je w required", sciezka, nazwa)
		}
		if podschemat, ok := s.Properties[nazwa]; ok && sch.rozwin(t, podschemat).Default != nil {
			t.Errorf("%s.%s: Go pomija puste pole (omitempty), a openapi.yaml daje mu default", sciezka, nazwa)
		}
	}

	for nazwa, poleTyp := range pola {
		podschemat, ok := s.Properties[nazwa]
		if !ok {
			continue
		}
		switch {
		case czyStruktura(poleTyp):
			sch.porownaj(t, sciezka+"."+nazwa, poleTyp, podschemat)
		case poleTyp.Kind() == reflect.Slice && czyStruktura(poleTyp.Elem()):
			sch.porownaj(t, sciezka+"."+nazwa+"[]", poleTyp.Elem(), sch.rozwin(t, podschemat).Items)
		}
	}
}

// polaJSON zwraca nazwy pól JSON struktury i ich typy (wskaźniki rozwinięte).
func polaJSON(typ reflect.Type) map[string]reflect.Type {
	pola := map[string]reflect.Type{}
	for i := range typ.NumField() {
		f := typ.Field(i)
		nazwa, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if nazwa == "" || nazwa == "-" {
			continue
		}
		ft := f.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		pola[nazwa] = ft
	}
	return pola
}

// polaOmitempty to pola, których Go nie wysyła przy wartości zerowej (tag `omitempty`).
func polaOmitempty(typ reflect.Type) map[string]bool {
	pola := map[string]bool{}
	for i := range typ.NumField() {
		nazwa, opcje, _ := strings.Cut(typ.Field(i).Tag.Get("json"), ",")
		if strings.Contains(","+opcje+",", ",omitempty,") {
			pola[nazwa] = true
		}
	}
	return pola
}

// czyStruktura: struktura z własnymi polami JSON (time.Time jest w JSON tekstem).
func czyStruktura(typ reflect.Type) bool {
	return typ.Kind() == reflect.Struct && typ != reflect.TypeFor[time.Time]()
}

func TestKontraktTypyGoZgodneZOpenAPI(t *testing.T) {
	sch := wczytajSchematy(t)
	pary := []struct {
		schemat string
		typ     reflect.Type
	}{
		{"Zdrowie", reflect.TypeFor[Zdrowie]()},
		{"Zestaw", reflect.TypeFor[Zestaw]()},
		{"Podsumowanie", reflect.TypeFor[Podsumowanie]()},
		{"Ostrzezenie", reflect.TypeFor[Ostrzezenie]()},
		{"NoweUruchomienie", reflect.TypeFor[NoweUruchomienie]()},
		{"Uruchomienie", reflect.TypeFor[Uruchomienie]()},
		{"Postep", reflect.TypeFor[Postep]()},
		{"Blad", reflect.TypeFor[Blad]()},
		{"Raport", reflect.TypeFor[Raport]()},
		{"Plan", reflect.TypeFor[model.Plan]()},
		{"DaneWejsciowe", reflect.TypeFor[model.Dane]()},
	}
	for _, p := range pary {
		t.Run(p.schemat, func(t *testing.T) {
			s, ok := sch[p.schemat]
			if !ok {
				t.Fatalf("brak schematu %s w openapi.yaml", p.schemat)
			}
			sch.porownaj(t, p.schemat, p.typ, s)
		})
	}
}

// Pola, które w data/dane_planu.json mają wartość null, muszą mieć "null" w typie schematu,
// inaczej typy frontendu obiecują wartość, której może nie być.
func TestKontraktNullZDanychDozwolony(t *testing.T) {
	sch := wczytajSchematy(t)
	b, err := os.ReadFile(plikDanych)
	if err != nil {
		t.Fatal(err)
	}
	var dane any
	if err := json.Unmarshal(b, &dane); err != nil {
		t.Fatal(err)
	}

	bledy := map[string]bool{}
	sch.sprawdzNull(t, "DaneWejsciowe", dane, sch["DaneWejsciowe"], bledy)
	for _, b := range slices.Sorted(maps.Keys(bledy)) {
		t.Error(b)
	}
}

// sprawdzNull przechodzi wartość z danych równolegle ze schematem. Błędy zbiera w mapie,
// żeby to samo pole z 174 wierszy było zgłoszone raz.
func (sch schematy) sprawdzNull(t *testing.T, sciezka string, wartosc any, s *schemat, bledy map[string]bool) {
	t.Helper()
	s = sch.rozwin(t, s)
	switch w := wartosc.(type) {
	case map[string]any:
		if s.Properties == nil {
			return // np. meta.tabele: słownik bez stałych pól
		}
		for klucz, v := range w {
			podschemat, ok := s.Properties[klucz]
			if !ok {
				bledy[sciezka+"."+klucz+": pole z danych brakujące w openapi.yaml"] = true
				continue
			}
			podschemat = sch.rozwin(t, podschemat)
			if v == nil {
				if !podschemat.dopuszczaNull() {
					bledy[sciezka+"."+klucz+": w danych jest null, a schemat go nie dopuszcza"] = true
				}
				continue
			}
			sch.sprawdzNull(t, sciezka+"."+klucz, v, podschemat, bledy)
		}
	case []any:
		for _, el := range w {
			sch.sprawdzNull(t, sciezka+"[]", el, s.Items, bledy)
		}
	}
}
