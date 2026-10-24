# Otwarte kwestie

Po rozstrzygnięciu kwestii: utwórz ADR w `decisions/`, usuń ją stąd i dopisz numer ADR-u.
Agent, którego zadanie zależy od otwartej kwestii, pyta użytkownika i nie zgaduje.

Rozstrzygnięte: #1 → ADR 0002, #5 → ADR 0004, #6 → ADR 0005, #9 → ADR 0003.

| # | Kwestia | Rekomendacja z planu | Blokuje |
|---|---|---|---|
| 2 | Router HTTP: chi, samo `net/http` (Go 1.22+), Gin/Echo? | chi | backend |
| 3 | Warstwa SQL: sqlc czy ORM (GORM)? | sqlc | backend |
| 4 | Wartości wag kryteriów miękkich i ich kolejność ważności | Tabela startowa w `plan.md` | solver, frontend |
| 7 | Ręczna edycja planu w MVP czy jako dodatek? | Dodatek | frontend, backend |
| 8 | Logowanie użytkowników? | Brak w MVP | backend, frontend |
