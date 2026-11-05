# Interfejs solvera (`internal/solver`)

```go
plan, err := solver.Solve(ctx, dane, solver.DomyslneWagi(), solver.ZSeed(42))
```

| Element | Znaczenie |
|---|---|
| `ctx` | przerwanie uruchomienia (np. anulowanie w API) |
| `dane` | `*model.Dane` po `model.Wczytaj` (dane spójne wg `model.Sprawdz`) |
| `wagi` | `solver.Wagi`, wagi kryteriów miękkich; `DomyslneWagi()` to tabela startowa z `plan.md` (otwarta kwestia #4) |
| `opcje...` | opcjonalne ustawienia uruchomienia (niżej); bez opcji działa wywołanie `Solve(ctx, dane, wagi)` |
| wynik | `model.Plan` w formacie `plan.json` v1, z przypisaniem dla każdych zajęć (ADR 0006) |
| `err` | tylko przerwanie przez `ctx` przed końcem konstrukcji (ADR 0006); przerwanie w trakcie wyżarzania zwraca najlepszy plan do tej pory |

Opcje są zmiennoargumentowe, żeby kolejne ustawienia dało się dodać bez zmiany sygnatury.
Nowa opcja nie łamie kontraktu, zmiana istniejącej tak.

| Opcja | Domyślnie | Znaczenie |
|---|---|---|
| `ZSeed(seed)` | 1 | ziarno losowania, trafia do `meta.seed`; ten sam seed i te same opcje dają ten sam plan |
| `ZIteracjami(n)` | 1 000 000 | iteracje wyżarzania na workera; 0 = sama konstrukcja |
| `ZWorkerami(n)` | 4 | równoległe workery (każdy z własnym seedem pochodnym); stała liczba, żeby wynik nie zależał od maszyny |
| `ZPostepem(f)` | brak | `f(Postep)` wołane co 10 000 iteracji przez każdego workera (np. do SSE); musi być bezpieczna współbieżnie |

Domyślnie solver działa ok. 1,5 s na 4 rdzeniach.

`solver.RozbicieKosztu(dane, plan, wagi)` zwraca koszt miękki planu z `Solve` z podziałem na kryteria
(klucze jak tagi JSON w `Wagi`). Suma kryteriów to `meta.koszt`.

Pola `Wagi` mają tagi JSON (`okienka_studentow`, `po_17`, …), więc backend może je przyjmować wprost z API.

## Funkcja kosztu miękkiego

`meta.koszt = Σ waga × kara`, bez kary za naruszenia twarde (ADR 0006). Kary:

| Kryterium (`Wagi`) | Kara |
|---|---|
| `okienka_studentow` | dla każdego studenta, dnia i tygodnia: wolne sloty między pierwszym a ostatnim zajętym; suma / 14 tygodni / 25 |
| `przeciazony_dzien` | jak wyżej, sloty zajęte ponad 8 |
| `okienka_prowadzacych` | dla każdego prowadzącego, dnia i tygodnia: wolne sloty między zajęciami; suma / 14 |
| `przerwa_max_2h` | prowadzący z uwagą „przerwa ≤ 2 h”: każdy slot przerwy ponad 2, suma / 14 |
| `po_17` | sloty 10–12 spotkania × `liczba_studentow` / 25 |
| `przed_9` | slot 1 spotkania × `liczba_studentow` / 25 |
| `za_duza_sala` | (pojemność − `liczba_studentow` − `dodatkowi_studenci`) / pojemność, na zajęcia |
| `preferencja_dni` | `skumulowane`: dni z zajęciami ponad ⌈sloty tygodniowo / 8⌉; `nieskumulowane`: dni poniżej min(5, liczba zajęć prowadzącego) |
| `dzien_wolny` | prowadzący z uwagą „jeden dzień wolny” ma zajęcia we wszystkich 5 dniach: 1 |
| `nie_po_17` | prowadzący z uwagą „nie po 17:00”: sloty 10–12 spotkania |
| `pora_dnia` | uwaga „rano”: sloty 7–12 spotkania; uwaga „popołudnie”: sloty 1–6 |
| `sale_a` | uwaga „sale A”: 1 za każde zajęcia w sali z oznaczeniem B |
| `wyklady_rocznika` | wykłady (forma `W`) rocznika w jednym dniu ponad 2 |

Tygodnie liczymy osobno, więc zajęcia `co_2_tygodnie` i `ten_sam_termin` (Z037 tyg. 1–7, Z038 tyg. 8–14)
wpływają na okienka tylko w swoich tygodniach. Uwagi z `inne_uwagi` są rozpoznawane po słowach kluczowych;
test sprawdza, że rozpoznane są wszystkie uwagi z danych.

Status: fazy 1–3 gotowe. Na prawdziwych danych plan bez naruszeń (sprawdzone walidatorem z
`validator/reguly-twarde` dla kilku seedów), koszt ok. 400–480 przy ustawieniach domyślnych.
