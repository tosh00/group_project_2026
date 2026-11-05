# Interfejs solvera (`internal/solver`)

```go
plan, err := solver.Solve(ctx, dane, solver.DomyslneWagi(), solver.ZSeed(42))
```

| Element | Znaczenie |
|---|---|
| `ctx` | przerwanie uruchomienia (np. anulowanie w API) |
| `dane` | `*model.Dane` po `model.Wczytaj` (dane spójne wg `model.Sprawdz`) |
| `wagi` | `solver.Wagi`, wagi kryteriów miękkich; `DomyslneWagi()` to tabela startowa z `plan.md` (otwarta kwestia #4) |
| `opcje...` | opcjonalne ustawienia uruchomienia, np. `ZSeed(seed)` (trafia do `meta.seed`); bez opcji działa wywołanie `Solve(ctx, dane, wagi)` |
| wynik | `model.Plan` w formacie `plan.json` v1, z przypisaniem dla każdych zajęć (ADR 0006) |
| `err` | tylko niepoprawne dane albo przerwanie przez `ctx` przed końcem konstrukcji (ADR 0006) |

Opcje są zmiennoargumentowe, żeby kolejne ustawienia (np. liczba workerów, limit czasu, postęp dla SSE)
dało się dodać bez zmiany sygnatury. Nowa opcja nie łamie kontraktu, zmiana istniejącej tak.

Pola `Wagi` mają tagi JSON (`okienka_studentow`, `po_17`, …), więc backend może je przyjmować wprost z API.

Status: szkic. Fazy 1 i 2 są gotowe: `Solve` zwraca pełny plan (na prawdziwych danych bez naruszeń,
sprawdzone walidatorem). `meta.koszt` to na razie suma kryteriów lokalnych (pora dnia, sala, uwagi
prowadzących); pełna funkcja kosztu i wyżarzanie przyjdą z fazą 3.
