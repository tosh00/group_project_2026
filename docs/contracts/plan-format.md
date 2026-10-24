# Format `plan.json` (wersja 1)

Przyjęty w [ADR 0003](../decisions/0003-format-plan-json.md).
Wynik solvera i wejście walidatora. Ten sam format zwraca CLI `cmd/solve`, przechowuje i zwraca backend.

```json
{
  "meta": {
    "wersja_formatu": 1,
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
      "tygodnie": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14],
      "zablokowane": false
    }
  ]
}
```

## `meta`

| Pole | Znaczenie |
|---|---|
| `wersja_formatu` | wersja tego formatu, obecnie `1`; czytający odrzuca nieznaną wersję |
| `silnik` | nazwa silnika, który wygenerował plan |
| `seed` | ziarno losowania (powtarzalność) |
| `koszt` | koszt miękki planu, bez kary 10 000 za naruszenia twarde (ADR 0006) |
| `utworzono` | czas utworzenia, RFC 3339 w UTC |

## `przypisania[]`

| Pole | Znaczenie |
|---|---|
| `zajecia_id` | id z tabeli `zajecia`; każde zajęcia występują dokładnie raz |
| `dzien` | 1–5 (dzień planu, pon–pt) |
| `slot_start` | 1–12; zajęcia zajmują sloty `slot_start … slot_start + sloty_na_spotkanie − 1` |
| `sala_id` | id z tabeli `sale` |
| `tygodnie` | rosnąca lista numerów tygodni 1–14, w których odbywa się spotkanie (w kodzie Go: maska `uint16`, bit 0 = tydzień 1) |
| `zablokowane` | opcjonalne, domyślnie `false`; `true` = przypięte ręcznie, solver tego nie przesuwa |

W JSON tygodnie są jawną listą, a nie maską, żeby plik był czytelny i nie zależał od konwencji bitów.

Solver zawsze zwraca przypisanie dla każdych zajęć (ADR 0006). Status planu nie jest częścią `plan.json`:
ustala go walidator.

## Raport walidatora (SZKIC)

Wynik `internal/validator` dla danego `plan.json`. Backend zapisuje go przy uruchomieniu, frontend
wyświetla status, podświetla zajęcia z naruszeniami i pokazuje listę naruszeń.

```json
{
  "status": "z_naruszeniami",
  "liczba_naruszen": 2,
  "naruszenia": [
    {
      "regula": 3,
      "nazwa": "kolizja_grup",
      "zajecia_id": ["Z014", "Z087"],
      "dzien": 2,
      "sloty": [5, 6],
      "sala_id": "A121",
      "tygodnie": [2, 4, 6, 8, 10, 12, 14],
      "opis": "Grupy INF1-R1-G2 i INF1-R1 mają jednocześnie Z014 i Z087"
    }
  ],
  "koszt_miekki": 1234.5,
  "koszt_kryteria": { "okienka_studentow": 420.0, "po_17": 96.0 }
}
```

| Pole | Znaczenie |
|---|---|
| `status` | `poprawny` (0 naruszeń) albo `z_naruszeniami` |
| `naruszenia` | **wszystkie** naruszenia, nie tylko pierwsze; jedno naruszenie na parę (regułę i zestaw zajęć) |
| `regula` | numer reguły twardej 1–10 z `data/README.md`; `nazwa` to stały identyfikator reguły |
| `zajecia_id` | zajęcia biorące udział w naruszeniu (1 lub więcej) |
| `dzien`, `sloty`, `sala_id`, `tygodnie` | gdzie występuje naruszenie: dzień, wspólne sloty, sala (jeśli dotyczy), wspólne tygodnie |
| `opis` | jedno zdanie po polsku do wyświetlenia w UI |
| `koszt_miekki`, `koszt_kryteria` | koszt miękki całkowity i rozbicie na kryteria; liczone zawsze, także przy naruszeniach |
