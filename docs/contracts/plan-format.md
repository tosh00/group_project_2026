# Format `plan.json` (SZKIC, do przyjęcia, patrz open-questions #9)

Wynik solvera i wejście walidatora. Ten sam format zwraca CLI `cmd/solve`, zapisuje backend
i może zwracać zewnętrzny silnik (np. CP-SAT w Pythonie).

```json
{
  "meta": {
    "silnik": "go-hybrid",
    "seed": 42,
    "koszt": 1234.5,
    "utworzono": "2026-10-24T14:00:00Z"
  },
  "przypisania": [
    {
      "zajecia_id": "Z001",
      "dzien": 1,
      "slot_start": 3,
      "sala_id": "A281",
      "tygodnie": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14]
    }
  ]
}
```

| Pole | Znaczenie |
|---|---|
| `zajecia_id` | id z tabeli `zajecia`; każde zajęcia występują dokładnie raz |
| `dzien` | 1–5 (dzień planu, pon–pt) |
| `slot_start` | 1–12; zajęcia zajmują sloty `slot_start … slot_start + sloty_na_spotkanie − 1` |
| `sala_id` | id z tabeli `sale` |
| `tygodnie` | numery tygodni 1–14, w których odbywa się spotkanie (w kodzie Go: maska `uint16`, bit 0 = tydzień 1) |

W JSON tygodnie są jawną listą, a nie maską, żeby plik był czytelny i nie zależał od konwencji bitów.
