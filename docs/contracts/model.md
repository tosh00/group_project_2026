# Pakiet `internal/model`

Struktury danych wejściowych, ich import i format planu. Szczegóły są w komentarzach w kodzie (`go doc ./internal/model`).

## Jak używać

```go
dane, err := model.Wczytaj("data/dane_planu.json") // dekoduje i sprawdza spójność (Sprawdz)
ix := model.NowyIndeks(dane)                       // mapy po id, kolizje grup, powiązania

z := ix.Zajecia["Z105"]
z.LiczbaOsob()        // liczba_studentow + dodatkowi_studenci (reguła 7)
z.DozwoloneTygodnie() // tydzien_od–tydzien_do bez tygodnie_niedostepne

ix.ZajeciaKolidujaStudenci(z1, z2) // wspólny student (reguła 5), ze studenci_grupy
model.WspolnyProwadzacy(z1, z2)    // wspólny prowadzący (reguła 3)

plan, err := model.WczytajPlan("plan.json") // sprawdza wersję formatu, nie sprawdza reguł planu
err = model.ZapiszPlan(w, plan)
```

## Konwencje

- Pola, które w JSON mogą być `null`: tekst to `""`, liczba to `0`. Dni, sloty i tygodnie liczymy od 1, więc `0` oznacza „brak”.
- `Sala.Pojemnosc = 0` oznacza brak danych (zdarza się tylko w salach niedostępnych do planowania).
- Tygodnie to `model.Tygodnie`, maska `uint16` (bit 0 = tydzień 1), w JSON zapisywana jako lista numerów.
- Kolizje grup są liczone ze `studenci_grupy`. `Sprawdz` porównuje je z dostarczoną tabelą `kolizje_grup`.
- Nieznane pola w JSON są błędem, zarówno w danych, jak i w planie.
- `Sprawdz` sprawdza spójność danych wejściowych, a nie reguły planu. Reguły 1–10 sprawdza `internal/validator`.
