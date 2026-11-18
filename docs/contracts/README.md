# Kontrakty między modułami

Kontrakt to wszystko, od czego zależy więcej niż jeden moduł. Dzięki kontraktom ustalonym
na początku każda osoba (i jej agent) może pracować niezależnie.

| Kontrakt | Plik | Używają | Status |
|---|---|---|---|
| Dane wejściowe | `data/dane_planu.json` (format: `data/README.md`) | wszyscy | stały |
| Struktury danych w Go | `internal/model/`, opis: [model.md](model.md) | solver, walidator, backend | gotowe (etap 0) |
| Plan wyjściowy | `plan.json`, opis: [plan-format.md](plan-format.md) | solver, walidator, backend, frontend | przyjęty, wersja 1 (ADR 0003) |
| Interfejs solvera | `Solve(ctx, dane, wagi, opcje...) (Plan, error)` w `internal/solver`, opis: [solver.md](solver.md) | backend | gotowe (fazy 1–3) |
| API HTTP | `api/openapi.yaml`, opis: [api.md](api.md) | backend, frontend | szkic 0.1 (2026-11-18), rozwijany równolegle z API |
| Schemat bazy | `migrations/` (ADR 0008, 0009) | backend | szkic (2026-11-18) |

## Zasady

- Zmiana kontraktu to osobny PR, z review właścicieli wszystkich modułów, które go używają.
- Zmiany wstecznie niekompatybilne opisujemy w opisie PR.
- Typy TypeScript generujemy z `api/openapi.yaml` (openapi-typescript), nie piszemy ich ręcznie.
