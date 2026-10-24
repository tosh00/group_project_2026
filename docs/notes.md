# Notatki o danych i problemie

Wiedza, która przyda się innym. Dopisujcie krótko, z datą. Źródło liczb: `data/`.

## Dane (analiza z planu, 2026-10-24)

- 174 zajęcia, 55 sal (30 dostępnych do planowania), 50 prowadzących, 6 roczników, 51 grup, 372 studentów.
- Import z `dane_planu.json`: listy są w nim tablicami, nie trzeba rozdzielać po `;`.
- **Kolizje grup bierzemy z `kolizje_grup` / `studenci_grupy`, nie z hierarchii grup.**
  Grupy z różnych podziałów nie zawsze mają wspólnych studentów (np. `INF1-R2`).
- `nr_tygodnia` to kolejne wystąpienie dnia planu, nie tydzień kalendarzowy (dni odpracowania: 8 I, 25 I, 26 I).
- Pracownie komputerowe: 68 zajęć, 11 sal po 20–30 miejsc; tłok w godzinach 9–16.
- Sale jednoznaczne (planować najpierw): B153, B472, robotyka B571/572 i B573/574, B273 (CISCO), A234 (MAC).
- Wykłady łączone z Fizyką (do 198 osób) mieszczą się tylko w A281 (264 miejsca); A283 ma 156.
- 14 zajęć `co_2_tygodnie`, 2 zajęcia `wybrane_tygodnie` (Z105: 9 z 14, Z169: 13 z 14 bez tyg. 6).
- 30 wierszy `rozne_dni`, 2 wiersze `ten_sam_termin`.
- Rada Wydziału: wtorek, sloty 9–10. WF rocznika INF1-R2: czwartek, sloty 1–4.
- 5 zajęć ma kilku prowadzących, wszyscy muszą być wolni.
- Litera A/B w nazwie sali (`A281`, `B153`) to część oznaczenia sali w jednym budynku WFiIS, a nie osobne budynki.
  Przejście między salami A i B nie wymaga przerwy (ustalenie zespołu, 2026-10-24).
