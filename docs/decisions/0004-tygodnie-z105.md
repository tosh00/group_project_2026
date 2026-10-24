# 0004. Wybór tygodni dla zajęć `wybrane_tygodnie`

- Status: przyjęta
- Data: 2026-10-24
- Decydują: zespół

## Kontekst

Dwa zajęcia mają częstotliwość `wybrane_tygodnie`. Z169 (13 z 14) ma tydzień 6 w
`tygodnie_niedostepne`, więc jego tygodnie są wyznaczone przez dane. Wybór dotyczy tylko
Z105: 9 spotkań w dowolnych 9 z tygodni 1–14, zawsze ten sam dzień, godzina i sala.

## Decyzja

- Tygodnie dla `wybrane_tygodnie` wybiera solver (np. tak, żeby najmniej kolidowały w sali i u studentów).
- Walidator nie wymaga konkretnych tygodni. Sprawdza tylko, czy jest dokładnie `liczba_spotkan` różnych
  tygodni z zakresu `tydzien_od`–`tydzien_do` i żaden nie należy do `tygodnie_niedostepne`.

## Konsekwencje

- Ogólna reguła działa dla obu zajęć; Z169 nie wymaga specjalnego przypadku.
- Solver generuje dla Z105 kilka wariantów maski tygodni (faza 1) i może je zmieniać w fazie 3.
