# 0005. Sale z szacowaną pojemnością bez marginesu

- Status: przyjęta
- Data: 2026-10-24
- Decydują: zespół

## Kontekst

Sale B151, B153 i B472 mają `pojemnosc_szacowana = true` (16 miejsc). Zajęcia Z071–Z080, Z136, Z137
mogą odbyć się tylko w B153 lub B472, a Z075–Z078 mają po 15 osób. Twardy margines 10% wykluczyłby
dla nich wszystkie sale i problem nie miałby rozwiązania. To samo dotyczy Z115 (20 osób, tylko A234 z 20 miejscami).

## Decyzja

- Reguła twarda 7 obowiązuje dokładnie jak w `data/README.md`:
  `pojemnosc ≥ liczba_studentow + dodatkowi_studenci`, bez marginesu, także dla pojemności szacowanych.
- Raport jakości (backend i frontend) pokazuje ostrzeżenie, gdy zajęcia trafiają do sali z szacowaną pojemnością.
  Ostrzeżenie nie wpływa na koszt.

## Konsekwencje

- Solver i walidator traktują wszystkie pojemności tak samo.
- Ostrzeżenie o szacowanej pojemności to informacja dla człowieka, nie kryterium optymalizacji.
