# 0003. Format pliku `plan.json`

- Status: przyjęta
- Data: 2026-10-24
- Decydują: zespół

## Kontekst

Solver, walidator, backend i frontend wymieniają się gotowym planem. Potrzebny jest jeden format,
ustalony przed startem prac, żeby moduły dało się rozwijać równolegle.

## Decyzja

Przyjmujemy format opisany w [`docs/contracts/plan-format.md`](../contracts/plan-format.md), wersja 1.
Względem szkicu doszły dwa pola:

- `meta.wersja_formatu`: liczba całkowita, obecnie `1`. Zmiana niekompatybilna podnosi wersję.
- `przypisania[].zablokowane`: opcjonalne, domyślnie `false`. Oznacza zajęcia przypięte ręcznie,
  których solver nie przesuwa. Przygotowuje kontrakt pod ręczną edycję bez późniejszej zmiany formatu.

## Rozważane opcje

- Tygodnie jako maska bitowa w JSON: zwięźle, ale nieczytelnie i zależnie od konwencji bitów. Odrzucone.
- Tygodnie jako jawna lista numerów 1–14: czytelne, łatwe do sprawdzenia. Przyjęte.

## Konsekwencje

- `cmd/solve` zapisuje ten format, `internal/validator` go czyta, backend go przechowuje i zwraca.
- Typy Go dla planu leżą w `internal/model` obok danych wejściowych.
- Zmiana formatu to nowy ADR i podniesienie `wersja_formatu`.
