# Plan projektu (skrót)

Pełna wersja z diagramami: https://claude.ai/artifact/Fp7h19TgAp6TyBp3NraJ94
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

Interfejs: `Solve(ctx, dane, wagi) (Plan, error)`. Własny algorytm, bez gotowych solverów (ADR 0002).

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
| „Sale A, jeśli możliwe” | zajęcia prowadzącego w sali z oznaczeniem B | 3 |
| Wykłady rocznika w jednym dniu | ponad limit | 2 |

A i B to oznaczenia sal (np. `A123`, `B657`) w jednym budynku WFiIS, a nie osobne budynki.
Przejście między salami A i B nie wymaga przerwy, więc nie ma kary za „zmianę budynku”.

## Stack

React 18 + TS + Vite (TanStack Query, React Router), Go 1.23+ (chi, pgx v5, sqlc, goose, slog),
PostgreSQL 16, Docker Compose, GitHub Actions (go test, golangci-lint, Vitest).

## Etapy

Solver i walidator idą pierwsze, bo to najtrudniejsza i najbardziej ryzykowna część; da się je budować
i testować na samym JSON przez CLI, bez bazy. API, baza i frontend idą równolegle: w dniu 1 ustalamy
format `plan.json`, interfejs `Solve` i szkic OpenAPI, więc backend pracuje z atrapą solvera, a frontend z mockiem API.
Funkcja kosztu nie jest osobnym etapem: powstaje razem z solverem, bo konstrukcja zachłanna wybiera według kosztu.

| Etap | Kiedy | Zakres |
|---|---|---|
| 0 | dzień 1 | Szkielet repo (moduł Go + `web/`), Docker Compose, CI, `internal/model` + import JSON; kontrakty: `plan.json`, `Solve(ctx, dane, wagi)`, szkic OpenAPI |
| 1 | tydzień 1 | Walidator 10 reguł z testami tabelarycznymi |
| 2 | tygodnie 1–4 | Solver z funkcją kosztu: fazy 1–3, benchmarki, strojenie wag. Poprawny plan dla 174 zajęć z `cmd/solve` po ok. 2 tygodniach, potem jakość |
| równolegle | tygodnie 1–3 | API i baza: migracje goose, sqlc, import, endpointy, pula workerów, SSE (z atrapą solvera) |
| równolegle | tygodnie 2–5 | Frontend MVP: dane, wagi, postęp, siatka, raport (najpierw na mocku z OpenAPI) |
| 3 | tydzień 5 | Integracja: prawdziwy solver w API, postęp przez SSE, walidator po każdym uruchomieniu, testy e2e |
| 4 | tydzień 6 | Dodatki i prezentacja: edycja ręczna, porównanie, eksport |
