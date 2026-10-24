# Kontrakty między modułami

Kontrakt to wszystko, od czego zależy więcej niż jeden moduł. Dzięki kontraktom ustalonym
na początku każda osoba (i jej agent) może pracować niezależnie.

| Kontrakt | Plik | Używają | Status |
|---|---|---|---|
| Dane wejściowe | `data/dane_planu.json` (format: `data/README.md`) | wszyscy | stały |
| Struktury danych w Go | `internal/model/`, opis: [model.md](model.md) | solver, walidator, backend | gotowe (etap 0) |
| Plan wyjściowy | `plan.json`, opis: [plan-format.md](plan-format.md) | solver, walidator, backend, frontend | przyjęty, wersja 1 (ADR 0003) |
| Interfejs solvera | `Solve(ctx, dane, wagi) (Plan, error)` w `internal/solver` | backend | do zrobienia (etap 0) |
| Naruszenia reguł twardych | `validator.Naruszenie` w `internal/validator` (JSON: `regula`, `zajecia`, `opis`) | backend, frontend | gotowe |
| API HTTP | `api/openapi.yaml` | backend, frontend | szkic w etapie 0, rozwijany równolegle z API |
| Schemat bazy | `migrations/` | backend | do zrobienia (ścieżka API i baza, tygodnie 1–3) |

## Zasady

- Zmiana kontraktu to osobny PR, z review właścicieli wszystkich modułów, które go używają.
- Zmiany wstecznie niekompatybilne opisujemy w opisie PR.
- Typy TypeScript generujemy z `api/openapi.yaml` (openapi-typescript), nie piszemy ich ręcznie.
