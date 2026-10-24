# Otwarte kwestie

Po rozstrzygnięciu kwestii: utwórz ADR w `decisions/`, usuń ją stąd i dopisz numer ADR-u.
Agent, którego zadanie zależy od otwartej kwestii, pyta użytkownika i nie zgaduje.

| # | Kwestia | Rekomendacja z planu | Blokuje |
|---|---|---|---|
| 1 | Własny algorytm czy gotowy solver (OR-Tools CP-SAT)? Czy prowadzący dopuszcza gotowe solvery? | Własny hybrydowy w Go jako główny, CP-SAT w Pythonie do porównania | solver |
| 2 | Router HTTP: chi, samo `net/http` (Go 1.22+), Gin/Echo? | chi | backend |
| 3 | Warstwa SQL: sqlc czy ORM (GORM)? | sqlc | backend |
| 4 | Wartości wag kryteriów miękkich i ich kolejność ważności | Tabela startowa w `plan.md` | solver, frontend |
| 5 | Wybór tygodni dla `wybrane_tygodnie` (Z105, Z169) | Z169 = wszystkie oprócz tyg. 6; Z105 wybiera solver | solver, walidator |
| 6 | Sale z `pojemnosc_szacowana = true`: margines bezpieczeństwa? | np. 10% | solver, walidator |
| 7 | Ręczna edycja planu w MVP czy jako dodatek? | Dodatek | frontend, backend |
| 8 | Logowanie użytkowników? | Brak w MVP | backend, frontend |
| 9 | Format pliku wyjściowego `plan.json` | Szkic w `contracts/plan-format.md` | solver, walidator, backend |
