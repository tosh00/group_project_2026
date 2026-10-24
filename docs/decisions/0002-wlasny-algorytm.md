# 0002. Własny algorytm zamiast gotowego solvera

- Status: przyjęta
- Data: 2026-10-24
- Decydują: zespół

## Kontekst

Plan można generować własnym algorytmem albo opisać problem dla gotowego solvera ograniczeń
(np. Google OR-Tools CP-SAT). OR-Tools nie ma oficjalnych bindingów do Go, więc wymagałby
osobnego skryptu w Pythonie uruchamianego jako podproces.

## Decyzja

Piszemy własny algorytm w Go w `internal/solver`, w trzech fazach z `docs/plan.md`:
przygotowanie domen, konstrukcja zachłanna z naprawą, symulowane wyżarzanie.
Nie używamy gotowych solverów ograniczeń ani optymalizacji (OR-Tools, CP-SAT, solvery MIP/SAT),
także jako drugiego silnika do porównania.

## Rozważane opcje

- Własny algorytm w Go: pełna kontrola, jeden język, solver bez zależności zewnętrznych; więcej pracy i ryzyko jakości.
- CP-SAT jako główny silnik: mniej pracy nad przeszukiwaniem; drugi język, podproces, mniejsza wartość dydaktyczna projektu.
- Własny + CP-SAT do porównania: dodatkowa praca bez wpływu na wynik.

## Konsekwencje

- Solver to czysty Go, bez zależności zewnętrznych poza biblioteką standardową.
- Poprawność potwierdza niezależny walidator (`internal/validator`), a nie solver.
- Format `plan.json` (ADR 0003) zostaje niezależny od silnika, więc decyzję da się w przyszłości zmienić nowym ADR-em.
