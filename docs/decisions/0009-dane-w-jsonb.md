# 0009. Dane wejściowe, plan i raport w kolumnach `jsonb`

- Status: proponowana
- Data: 2026-11-18
- Decydują: @KajetanJanusz, @krzysiu-w

## Kontekst

Backend przechowuje wgrane pliki `dane_planu.json` (kilkanaście tabel: sale, zajęcia, prowadzący…),
wygenerowane plany (`plan.json` v1) i raporty walidatora. Solver i walidator i tak dostają całe dane
jako `*model.Dane` (z `model.Dekoduj`), a plan jako `model.Plan`. W MVP nikt nie edytuje danych w UI.

## Decyzja

Wgrany plik zapisujemy w całości w kolumnie `jsonb` tabeli `zestawy_danych`. Plan i raport walidatora też
trzymamy jako `jsonb` (tabela `plany`). Tym, czym zarządza backend (uruchomienia: status, wagi, seed, czasy,
koszt, liczba naruszeń), są zwykłe kolumny w tabeli `uruchomienia`.

## Rozważane opcje

- **A. `jsonb`:** jedna migracja na start, `internal/model` zostaje jedynym opisem struktury danych,
  zmiana formatu danych nie wymaga migracji. Minus: zapytania „po sali” czy „po prowadzącym” przez `jsonb`
  (da się, ale nie jest to wygodne).
- **B. Tabele znormalizowane (`sala`, `zajecia`, `prowadzacy`, …):** zapytania SQL po każdym polu.
  Minus: kilkanaście tabel i zapytań, druga kopia struktury z `internal/model`, a w MVP nikt tych zapytań nie potrzebuje.

## Konsekwencje

- Import: `model.Dekoduj` sprawdza plik, a do bazy trafia oryginalna treść żądania (bez zmian).
- Frontend dostaje dane zestawu przez API w oryginalnym formacie (`GET /api/zestawy/{id}/dane`).
- Gdy pojawi się edycja danych w UI (np. własne ograniczenia użytkownika z `plan.md`), można przejść
  na tabele osobną migracją, bez zmiany API.
