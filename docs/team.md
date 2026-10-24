# Zespół i właściciele modułów

Właściciel modułu decyduje o jego wnętrzu i robi review zmian w nim.
Uzupełnijcie tabelę i odpowiadający jej plik `.github/CODEOWNERS`.

| Moduł | Ścieżki | Właściciel | GitHub |
|---|---|---|---|
| Model danych + import | `internal/model/` | _do ustalenia_ | `@?` |
| Solver (fazy 1–3, koszt) | `internal/solver/`, `cmd/solve/` | _do ustalenia_ | `@?` |
| Walidator + testy | `internal/validator/` | _do ustalenia (inna osoba niż autor solvera)_ | `@?` |
| Backend + baza | `cmd/server/`, `internal/api/`, `internal/store/`, `migrations/`, `api/` | _do ustalenia_ | `@?` |
| Frontend | `web/` | _do ustalenia_ | `@?` |
| Infrastruktura (Docker, CI) | `docker-compose.yml`, `.github/workflows/` | _do ustalenia_ | `@?` |

## Kanały

- Zadania i status: GitHub Issues / Projects.
- Dyskusje o kontraktach: komentarze w PR zmieniającym kontrakt.
- Decyzje: `docs/decisions/`.
