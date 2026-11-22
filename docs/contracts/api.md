# API HTTP (`api/openapi.yaml`) — SZKIC 0.2

Źródłem prawdy jest `api/openapi.yaml`; tu jest skrót i to, czego nie widać w schematach.
Wszystkie ścieżki pod `/api`. Zakres MVP: bez edycji ręcznej i logowania (ADR 0010).

Frontend generuje typy z tego pliku (`web/src/api/schema.gen.ts`, `pnpm gen:api`, ADR 0011). Zmiana
`openapi.yaml` bez ponownego wygenerowania typów nie przejdzie CI frontendu, więc PR zmieniający kontrakt
regeneruje też `schema.gen.ts`.

Test `internal/api/kontrakt_test.go` (bez bazy, część `go test ./...` i CI) pilnuje, żeby schematy
odpowiedzi i `DaneWejsciowe` miały dokładnie te pola, co typy Go (`internal/api`, `internal/model`), oraz żeby
pola, które w `data/dane_planu.json` bywają `null`, miały w schemacie typ z `"null"`. Zmiana DTO albo modelu
bez zmiany `openapi.yaml` kończy się czerwonym testem. Frontend (Vite w dev, nginx w Dockerze) jest pod tym samym originem co API,
więc backend nie obsługuje CORS.

| Metoda i ścieżka | Co robi |
|---|---|
| `GET /zdrowie` | stan serwera, bazy i rodzaj walidatora (`atrapa` / `internal/validator`) |
| `GET /wagi/domyslne` | `solver.DomyslneWagi()` jako JSON |
| `POST /zestawy?nazwa=` | wgranie `dane_planu.json` (treść żądania = plik); 422 z listą niespójności |
| `GET /zestawy`, `GET /zestawy/{id}` | podsumowanie (liczby) i ostrzeżenia (`pojemnosc_szacowana`) |
| `DELETE /zestawy/{id}` | usunięcie zestawu z uruchomieniami i planami (204); 409 `trwa_generowanie`, gdy generowanie trwa |
| `GET /zestawy/{id}/dane` | dane wejściowe (`DaneWejsciowe`, 17 tabel), z których frontend bierze nazwy do kart i filtrów |
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

## Walidator

Serwer ocenia każdy plan przez `internal/validator` (adapter `api.PelnyWalidator` w `internal/api/walidacja.go`),
więc raport i `/zdrowie` mają `walidator: "internal/validator"`. `api.Atrapa` (0 naruszeń, `walidator: "atrapa"`)
została tylko do testów; frontend nadal powinien pokazywać, że raport `atrapa` nie jest sprawdzony.
Naruszenia mają wszystkie pola z `plan-format.md` (`regula`, `nazwa`, `zajecia_id`, `dzien`, `sloty`, `sala_id`,
`tygodnie`, `opis`); pola położenia są pomijane, gdy nie dotyczą naruszenia (tabela w `plan-format.md`).

## Błędy

Każdy błąd to `{ "kod": "...", "komunikat": "...", "szczegoly": [...] }`. Kody: `zle_zadanie` (400),
`za_duze` (413), `niespojne_dane` (422), `nie_znaleziono` / `brak_planu` (404), `zakonczone` / `trwa_generowanie` (409),
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

Przykład w szkicu raportu w `plan-format.md` miał `"regula": 3` przy `kolizja_grup`; poprawione na 5 (2026-11-18).

## Różnice względem briefu frontendu

`docs/design/claude-design-brief.md` używa ścieżki `POST /api/runs/{id}/sprawdz-ruch`. W API ścieżki są po polsku
(`/api/uruchomienia/...`, konwencja z CLAUDE.md), a sprawdzanie ruchu wraca jako dodatek po MVP (ADR 0010),
docelowo jako `POST /api/uruchomienia/{id}/sprawdz-ruch`.

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
