# Otwarte kwestie

Po rozstrzygnięciu kwestii: utwórz ADR w `decisions/`, usuń ją stąd i dopisz numer ADR-u.
Agent, którego zadanie zależy od otwartej kwestii, pyta użytkownika i nie zgaduje.

Rozstrzygnięte: #1 → ADR 0002, #9 → ADR 0003.

| # | Kwestia | Rekomendacja z planu | Blokuje |
|---|---|---|---|
| 2 | Router HTTP: chi, samo `net/http` (Go 1.22+), Gin/Echo? | chi | backend |
| 3 | Warstwa SQL: sqlc czy ORM (GORM)? | sqlc | backend |
| 4 | Wartości wag kryteriów miękkich i ich kolejność ważności | Tabela startowa w `plan.md` | solver, frontend |
| 5 | Wybór tygodni dla `wybrane_tygodnie`. Z169 wynika z danych (tyg. 6 w `tygodnie_niedostepne`), więc dotyczy tylko Z105 (9 z 14) | Z105 wybiera solver; walidator sprawdza tylko liczbę i zakres tygodni | solver, walidator |
| 6 | Sale z `pojemnosc_szacowana = true`: margines bezpieczeństwa? Twardy margines 10% czyni problem nierozwiązywalnym (Z075–Z078: 15 osób, jedyne sale B153/B472 po 16 miejsc) | Reguła 7 dokładnie jak w README; szacowana pojemność tylko jako ostrzeżenie w raporcie | solver, walidator |
| 7 | Ręczna edycja planu w MVP czy jako dodatek? | Dodatek | frontend, backend |
| 8 | Logowanie użytkowników? | Brak w MVP | backend, frontend |
