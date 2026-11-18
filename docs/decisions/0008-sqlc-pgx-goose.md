# 0008. Warstwa SQL: sqlc + pgx, migracje goose

- Status: proponowana
- Data: 2026-11-18
- Decydują: @KajetanJanusz, @krzysiu-w

## Kontekst

Backend zapisuje zestawy danych, uruchomienia solvera i plany w PostgreSQL 16. Potrzebujemy warstwy
dostępu do bazy i migracji schematu. Otwarta kwestia #3.

## Decyzja

Zapytania piszemy w SQL i generujemy z nich kod Go przez **sqlc** (sterownik **pgx v5**).
Schemat zmieniamy migracjami **goose** w `migrations/`, uruchamianymi przez serwer przy starcie
(pliki osadzone w binarce przez `embed`).

## Rozważane opcje

- **sqlc:** SQL jest jawny i sprawdzany przy generowaniu względem schematu z migracji, kod typowany,
  bez refleksji. Minus: krok generowania (`sqlc generate`) i wygenerowane pliki w repo.
- **GORM:** mniej SQL na start. Minus: zapytania ukryte w ORM, błędy dopiero w czasie działania,
  słaba obsługa `jsonb` i typów Postgresa, automigracje kłócą się z goose.
- **Ręczny pgx:** zero generowania. Minus: ręczne skanowanie wierszy, błędy literówek w SQL dopiero w testach.

## Konsekwencje

- Zapytania: `internal/store/zapytania/*.sql`, konfiguracja: `internal/store/sqlc.yaml`,
  wygenerowany kod: `internal/store/db/` (commitowany, nie edytujemy ręcznie).
- sqlc czyta schemat bezpośrednio z `migrations/`, więc migracja i zapytania nie mogą się rozjechać.
- `sqlc` nie musi być zainstalowany: `docker run --rm -v "$PWD":/src -w /src/internal/store sqlc/sqlc:1.27.0 generate`.
- Z `docs/open-questions.md` znika kwestia #3.
