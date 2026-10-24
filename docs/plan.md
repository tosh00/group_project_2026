# Plan projektu (skrót)

Pełna wersja z diagramami: https://claude.ai/artifact/Y2bEygpkVhLJdbvjqdd3QU
Ten plik jest wersją roboczą w repo. Gdy coś się zmieni, aktualizujemy go tutaj, a decyzje zapisujemy w `decisions/`.

## Algorytm: trzy fazy

1. **Przygotowanie.** Każde zajęcia dostają warianty maski tygodni (`uint16`, 14 bitów) bez `tygodnie_niedostepne`.
   Domeny: dozwolone sale (typ, wyposażenie, pojemność, dostępność) i terminy (bez wykluczeń, Rady, blokad).
   Graf konfliktów: wspólny prowadzący lub kolidujące grupy. Pary `ten_sam_termin` łączymy w jeden węzeł.
   Nakładanie się: ten sam dzień ∧ wspólny slot ∧ `maskaA & maskaB != 0`.
2. **Konstrukcja zachłanna.** Kolejność: najmniejsza domena, potem najwyższy stopień w grafie, najdłuższe i największe.
   Wybór kandydata o najmniejszym przyroście kosztu, forward checking, naprawa ślepych zaułków.
3. **Symulowane wyżarzanie.** Ruchy: zmiana sali, przesunięcie terminu, zamiana terminów, zmiana parzystości.
   Tylko ruchy dopuszczalne, koszt liczony przyrostowo, kilka goroutines z różnymi seedami.

Interfejs: `Solve(ctx, dane, wagi) (Plan, error)`. Opcjonalnie drugi silnik CP-SAT (Python, podproces).

## Funkcja kosztu

`koszt(plan) = Σ waga_k × kara_k(plan)`. Twarde reguły 1–10 to filtr, a nie składnik kosztu.

| Kryterium | Kara | Waga startowa |
|---|---|---|
| Okienka studentów | wolne sloty między zajęciami grupy w dniu (× studenci/25) | 10 |
| Przeciążony dzień studenta | każdy slot powyżej 8 | 15 |
| Okienka prowadzących | wolne sloty między zajęciami prowadzącego | 5 |
| Zajęcia po 17:00 | slot 10–12 × studenci/25 | 4 |
| Zajęcia przed 9:00 | slot 1 × studenci/25 | 3 |
| Za duża sala | (pojemność − osoby) / pojemność | 6 |
| Skumulowane / nieskumulowane | odchylenie liczby dni | 8 |
| „Nie po 17:00” (inne_uwagi) | slot ≥ 10 | 20 |
| „Jeden dzień wolny” | prowadzący ma zajęcia w 5 dniach | 10 |
| „Przerwa ≤ 2 h” | sloty przerwy ponad 2 | 8 |
| Pora dnia (inne_uwagi) | sloty poza preferowaną połową dnia | 3 |
| „Budynek A” | zajęcia w budynku B | 3 |
| Zmiana budynku bez przerwy | kolejne zajęcia grupy w innym budynku | 4 |
| Wykłady rocznika w jednym dniu | ponad limit | 2 |

## Stack

React 18 + TS + Vite (TanStack Query, React Router), Go 1.23+ (chi, pgx v5, sqlc, goose, slog),
PostgreSQL 16, Docker Compose, GitHub Actions (go test, golangci-lint, Vitest).

## Etapy

| Etap | Tydzień | Zakres |
|---|---|---|
| 0 | 1 | Szkielet repo, Docker Compose, CI, `internal/model` + import JSON, kontrakty |
| 1 | 2 | Walidator 10 reguł z testami tabelarycznymi |
| 2 | 3–4 | Solver fazy 1–2: poprawny plan dla 174 zajęć z `cmd/solve` |
| 3 | 4–5 | Funkcja kosztu, wyżarzanie, benchmarki, strojenie wag |
| 4 | 5–6 | Migracje, sqlc, API, pula workerów, SSE |
| 5 | 6–7 | Frontend MVP: dane, wagi, postęp, siatka, raport |
| 6 | 8 | Dodatki: edycja ręczna, porównanie, eksport, CP-SAT |
