# API HTTP (`api/openapi.yaml`) — SZKIC 0.1

Źródłem prawdy jest `api/openapi.yaml`; tu jest skrót i to, czego nie widać w schematach.
Wszystkie ścieżki pod `/api`. Zakres MVP: bez edycji ręcznej i logowania (ADR 0010).

| Metoda i ścieżka | Co robi |
|---|---|
| `GET /zdrowie` | stan serwera, bazy i rodzaj walidatora (`atrapa` / `internal/validator`) |
| `GET /wagi/domyslne` | `solver.DomyslneWagi()` jako JSON |
| `POST /zestawy?nazwa=` | wgranie `dane_planu.json` (treść żądania = plik); 422 z listą niespójności |
| `GET /zestawy`, `GET /zestawy/{id}` | podsumowanie (liczby) i ostrzeżenia (`pojemnosc_szacowana`) |
| `GET /zestawy/{id}/dane` | oryginalny plik, z którego frontend bierze nazwy do kart |
| `POST /zestawy/{id}/uruchomienia` | `{wagi?, seed?, iteracje?}` → 202, uruchomienie w kolejce |
| `GET /uruchomienia?zestaw_id=`, `GET /uruchomienia/{id}` | stan uruchomienia |
| `POST /uruchomienia/{id}/zatrzymaj` | anulowanie; po konstrukcji zostaje najlepszy plan |
| `GET /uruchomienia/{id}/postep` | SSE: `stan`, `postep`, `koniec` |
| `GET /uruchomienia/{id}/plan` | `plan.json` v1 |
| `GET /uruchomienia/{id}/raport` | raport walidatora (format z `plan-format.md`) |

## Cykl życia uruchomienia

```
w_kolejce ──► trwa (faza: konstrukcja → optymalizacja → walidacja) ──► poprawny | z_naruszeniami
    │            │
    └────────────┴──► blad   (zatrzymanie przed końcem konstrukcji, restart serwera)
```

- Status końcowy `poprawny` / `z_naruszeniami` ustala walidator, nie solver (ADR 0006).
- Zatrzymanie w trakcie optymalizacji kończy się planem i `zatrzymane: true`.
- Po restarcie serwera uruchomienia `w_kolejce` i `trwa` dostają status `blad` („przerwane restartem serwera”).
- Pula workerów ma stałą liczbę miejsc (`WORKERY`, domyślnie 1, bo jeden solver zajmuje 4 rdzenie)
  i ograniczoną kolejkę; pełna kolejka to 503 `kolejka_pelna`.

## Walidator: atrapa do czasu merge'a `internal/validator`

Dopóki `internal/validator` nie jest w `main`, backend używa atrapy: zwraca 0 naruszeń, a raport
i `/zdrowie` mają `walidator: "atrapa"`. Frontend powinien wtedy pokazywać, że status nie jest sprawdzony.
Po merge'u podmieniamy atrapę na adapter `validator.Nowy(dane).Sprawdz(plan)` w `internal/api/walidacja.go`;
kontrakt API się nie zmienia (pole `walidator` zmienia wartość).

Rozjazd do uzgodnienia z @tosh00: `validator.Naruszenie` na gałęzi `validator/reguly-twarde` ma pola
`regula`, `zajecia`, `opis`, a szkic raportu w `plan-format.md` także `nazwa`, `dzien`, `sloty`, `sala_id`,
`tygodnie`. W API są one opcjonalne (poza `nazwa`, którą backend uzupełnia z numeru reguły).

## Błędy

Każdy błąd to `{ "kod": "...", "komunikat": "...", "szczegoly": [...] }`. Kody: `zle_zadanie` (400),
`za_duze` (413), `niespojne_dane` (422), `nie_znaleziono` / `brak_planu` (404), `zakonczone` (409),
`kolejka_pelna` (503), `blad_serwera` (500).

## Nazwy reguł w raporcie (`Naruszenie.nazwa`)

Numeracja z `data/README.md`, sekcja 5 (tak samo jak stałe w `internal/validator`):

| Nr | `nazwa` | Nr | `nazwa` |
|---|---|---|---|
| 1 | `przypisanie` | 6 | `typ_sali` |
| 2 | `tygodnie` | 7 | `pojemnosc` |
| 3 | `kolizja_prowadzacego` | 8 | `dezyderaty` |
| 4 | `kolizja_sali` | 9 | `blokady_stale` |
| 5 | `kolizja_grup` | 10 | `powiazania` |

Uwaga: przykład w szkicu raportu w `plan-format.md` ma `"regula": 3, "nazwa": "kolizja_grup"`,
a kolizja grup (studentów) to reguła 5. Do poprawienia przy przyjmowaniu formatu raportu.

## Uruchomienie lokalne

```sh
docker run -d --name planer-pg -e POSTGRES_USER=planer -e POSTGRES_PASSWORD=planer -p 55432:5432 postgres:16
cp .env.example .env          # raz; DATABASE_URL, TEST_DATABASE_URL, ADDR, WORKERY, KOLEJKA
go run ./cmd/server
curl -X POST --data-binary @data/dane_planu.json localhost:8080/api/zestawy
```

`cmd/server` i testy wczytują `.env` z katalogu repo same; zmienna ustawiona w środowisku wygrywa z plikiem.
Testy z bazą (`go test ./...`) potrzebują `TEST_DATABASE_URL`; każdy test dostaje własną, tymczasową bazę.
Bez tej zmiennej testy z bazą są pomijane.
