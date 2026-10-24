# Zespół i właściciele modułów

Właściciel modułu decyduje o jego wnętrzu i robi review zmian w nim.
Tabela odpowiada plikowi `.github/CODEOWNERS`; zmieniajcie oba razem.

| Moduł | Ścieżki | Właściciel | GitHub |
|---|---|---|---|
| Model danych + import | `internal/model/` | Bartosz | `@tosh00` |
| Solver (fazy 1–3, koszt) | `internal/solver/`, `cmd/solve/` | krzysiu-w | `@krzysiu-w` |
| Walidator + testy | `internal/validator/`, `cmd/waliduj/` | Bartosz | `@tosh00` |
| Backend + baza | `cmd/server/`, `internal/api/`, `internal/store/`, `migrations/`, `api/` | Kajetan | `@KajetanJanusz` |
| Frontend | `web/`, `docs/design/` | shokera | `@shokera` |
| Infrastruktura (Docker, CI) | `docker-compose.yml`, `Dockerfile*`, `.github/workflows/` | shokera | `@shokera` |

## Kanały

- Zadania i status: GitHub Issues / Projects.
- Dyskusje o kontraktach: komentarze w PR zmieniającym kontrakt.
- Decyzje: `docs/decisions/`.
