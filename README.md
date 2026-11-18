# Planer zajęć WFiIS

Generator planu zajęć dla kierunku Informatyka (WFiIS UŁ, semestr zimowy 2026/27).

- Zadanie i dane: [`data/README.md`](data/README.md)
- Plan projektu: [`docs/plan.md`](docs/plan.md)
- Zespół i moduły: [`docs/team.md`](docs/team.md)
- Decyzje: [`docs/decisions/`](docs/decisions/), otwarte kwestie: [`docs/open-questions.md`](docs/open-questions.md)
- Kontrakty: [`docs/contracts/`](docs/contracts/)
- Zasady dla agentów (i ludzi): [`CLAUDE.md`](CLAUDE.md)

## Stan modułów

| Moduł | Ścieżki | Stan |
|---|---|---|
| Model danych + import | `internal/model/` | gotowe |
| Solver | `internal/solver/`, `cmd/solve/` | fazy 1–3 gotowe |
| Walidator | `internal/validator/` | gałąź `validator/reguly-twarde`, jeszcze nie w `main`; backend używa atrapy |
| Backend + baza | `cmd/server/`, `internal/api/`, `internal/store/`, `migrations/`, `api/` | MVP (szkic kontraktu 0.1) |
| Frontend | `web/` | do zrobienia |
| Docker Compose, CI | `docker-compose.yml`, `.github/workflows/` | do zrobienia |

## Wymagania

- Go 1.23+
- Docker (PostgreSQL 16 dla backendu; opcjonalnie sqlc i lint OpenAPI bez instalacji)

## Szybki start: backend

```sh
# 1. Baza (raz; potem: docker start planer-pg)
docker run -d --name planer-pg -e POSTGRES_USER=planer -e POSTGRES_PASSWORD=planer \
  -p 55432:5432 postgres:16

# 2. Ustawienia lokalne (raz)
cp .env.example .env

# 3. Serwer na :8080 (migracje stosują się przy starcie)
go run ./cmd/server
```

`cmd/server` i testy same wczytują `.env` z katalogu repo; zmienna ustawiona w środowisku wygrywa z plikiem.

| Zmienna | Flaga | Domyślnie | Znaczenie |
|---|---|---|---|
| `DATABASE_URL` | `-db` | (wymagana) | adres PostgreSQL |
| `ADDR` | `-addr` | `:8080` | adres nasłuchu |
| `WORKERY` | `-workery` | `1` | równoległe uruchomienia solvera (każde zajmuje 4 rdzenie) |
| `KOLEJKA` | `-kolejka` | `16` | ile uruchomień może czekać; pełna kolejka → 503 |
| `TEST_DATABASE_URL` | — | — | baza dla testów z bazą (bez niej są pomijane) |

### Przykład: od pliku do planu

```sh
curl -s localhost:8080/api/zdrowie

Z=$(curl -s -X POST --data-binary @data/dane_planu.json localhost:8080/api/zestawy | jq -r .id)
U=$(curl -s -X POST -d '{"seed":42,"wagi":{"po_17":8}}' \
  localhost:8080/api/zestawy/$Z/uruchomienia | jq -r .id)

curl -N localhost:8080/api/uruchomienia/$U/postep            # SSE: stan → postep… → koniec
curl -s localhost:8080/api/uruchomienia/$U/plan | jq '.przypisania | length'
curl -s localhost:8080/api/uruchomienia/$U/raport | jq '{status, walidator, koszt_miekki}'
```

Zatrzymanie: `curl -X POST localhost:8080/api/uruchomienia/$U/zatrzymaj`. Po konstrukcji solver oddaje
najlepszy dotychczasowy plan (`zatrzymane: true`).

### Endpointy

Kontrakt: [`api/openapi.yaml`](api/openapi.yaml), opis i szczegóły (cykl życia, SSE, błędy):
[`docs/contracts/api.md`](docs/contracts/api.md). Wszystko pod `/api`, bez logowania (ADR 0010).

| Metoda | Ścieżka | Co robi |
|---|---|---|
| GET | `/zdrowie` | stan serwera, bazy i rodzaj walidatora (`atrapa` / `internal/validator`) |
| GET | `/wagi/domyslne` | domyślne wagi kryteriów miękkich |
| GET | `/zestawy` | lista wgranych zestawów danych |
| POST | `/zestawy?nazwa=` | wgranie `dane_planu.json` (treść żądania = plik, do 10 MB); 422 z listą niespójności |
| GET | `/zestawy/{id}` | podsumowanie (liczby tabel) i ostrzeżenia |
| GET | `/zestawy/{id}/dane` | dane zestawu w formacie `dane_planu.json` |
| POST | `/zestawy/{id}/uruchomienia` | nowe uruchomienie: `{wagi?, seed?, iteracje?}` → 202 |
| GET | `/uruchomienia?zestaw_id=` | lista uruchomień |
| GET | `/uruchomienia/{id}` | stan uruchomienia (`w_kolejce` → `trwa` → `poprawny` / `z_naruszeniami` / `blad`) |
| POST | `/uruchomienia/{id}/zatrzymaj` | zatrzymanie (409, jeśli już zakończone) |
| GET | `/uruchomienia/{id}/postep` | postęp na żywo (SSE): `stan`, `postep`, `koniec` |
| GET | `/uruchomienia/{id}/plan` | plan w formacie `plan.json` v1 |
| GET | `/uruchomienia/{id}/raport` | raport walidatora z rozbiciem kosztu na kryteria |

## Solver z linii poleceń

Bez bazy i serwera:

```sh
go run ./cmd/solve -in data/dane_planu.json -out plan.json -seed 42 -v
```

Flagi: `-iteracje` (na workera, domyślnie 1 000 000; 0 = sama konstrukcja), `-workery` (domyślnie 4),
`-out -` (wynik na stdout), `-v` (czas i rozbicie kosztu na stderr). Opis interfejsu: [`docs/contracts/solver.md`](docs/contracts/solver.md).

## Testy

```sh
go test ./...            # z .env: także testy z bazą (każdy test tworzy własną tymczasową bazę)
go test -race ./...
go test -v ./internal/api/ | grep -E '^--- (PASS|SKIP|FAIL)'   # które testy z bazą się uruchomiły
```

Bez `TEST_DATABASE_URL` testy `internal/store` i `internal/api`, które potrzebują bazy, są pomijane (SKIP),
a `go test` i tak kończy się `ok`.

## Komendy dla backendu

```sh
# Kod Go z zapytań SQL po zmianie internal/store/zapytania/*.sql albo migrations/*.sql (ADR 0008)
docker run --rm -v "$PWD":/src -w /src/internal/store sqlc/sqlc:1.27.0 generate

# Lint kontraktu API
docker run --rm -v "$PWD":/spec redocly/cli:1.25.11 lint --extends=minimal /spec/api/openapi.yaml

# Podgląd kontraktu API w przeglądarce: http://localhost:8081
docker run --rm -p 8081:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v "$PWD/api":/spec swaggerapi/swagger-ui

# Usunięcie lokalnej bazy razem z danymi
docker rm -f planer-pg
```

Nowa migracja to nowy plik `migrations/NNNNN_opis.sql` z sekcjami `-- +goose Up` / `-- +goose Down`.
Serwer stosuje ją przy starcie. Migracje to kontrakt, więc zmiana idzie przez PR z review (`docs/contracts/`).
