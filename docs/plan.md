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
   Jeśli naprawa nie znajdzie opcji dla zajęć w limicie iteracji, solver przechodzi w **tryb awaryjny**
   (patrz niżej): wstawia te zajęcia tam, gdzie łamią najmniej reguł twardych, i jedzie dalej.
3. **Symulowane wyżarzanie.** Ruchy: zmiana sali, przesunięcie terminu, zamiana terminów, zmiana parzystości.
   Koszt liczony przyrostowo, kilka goroutines z różnymi seedami. W trybie normalnym tylko ruchy dopuszczalne;
   w trybie awaryjnym także niedopuszczalne, ale z karą za naruszenia w koszcie.

Interfejs: `Solve(ctx, dane, wagi) (Plan, error)`. Własny algorytm, bez gotowych solverów (ADR 0002).

## Tryb awaryjny i walidator (ADR 0006)

Solver **zawsze zwraca plan**, w którym każde zajęcia mają przypisanie. Zamiast „nie udało się”
użytkownik dostaje plan i listę konkretnych problemów do naprawienia ręcznie albo w danych.

- **Tryb domyślny bez zmian:** najpierw próba planu zgodnego ze wszystkimi regułami twardymi.
- **Tryb awaryjny:** gdy naprawa ślepego zaułka nie znajdzie opcji w limicie iteracji, zajęcia nie są pomijane.
  Trafiają do kandydata z najmniejszą liczbą naruszeń twardych (remis: mniejszy przyrost kosztu miękkiego).
  Format `plan.json` v1 się nie zmienia: status planu wynika z raportu walidatora, nie z pola w planie.
- **Wyżarzanie usuwa naruszenia:** w trybie awaryjnym `koszt = 10 000 × liczba_naruszeń + koszt_miękki`.
  Waga 10 000 jest większa niż każdy realny koszt miękki, więc wyżarzanie najpierw usuwa naruszenia,
  a dopiero potem poprawia jakość. Gdy liczba naruszeń spadnie do 0, solver wraca do ruchów tylko dopuszczalnych.
- **Walidator ma ostatnie słowo:** po każdym uruchomieniu sprawdza plan niezależnie od solvera i ustala status:
  `poprawny` (0 naruszeń) albo `z_naruszeniami` (N). Wypisuje **wszystkie** naruszenia, nie tylko pierwsze:
  regułę, zajęcia, dzień, sloty, salę i tygodnie. Format raportu: `contracts/plan-format.md`.
- **Dwa rodzaje „nieidealności” są rozdzielone:** naruszenia twarde (plan niedopuszczalny) i koszt miękki
  (plan dopuszczalny, ale z okienkami itd.). Raport zawsze pokazuje oba; koszt miękki w `meta.koszt` nie zawiera kary 10 000.
- **API:** status uruchomienia `poprawny` / `z_naruszeniami` z liczbą naruszeń. Status `blad` tylko dla
  niepoprawnych danych wejściowych albo przerwania przed końcem konstrukcji.
- **Frontend:** znacznik statusu nad siatką, zajęcia z naruszeniami podświetlone na siatce, lista naruszeń
  z przejściem do zajęć, filtr „tylko naruszenia”. Ruch ręczny jest dozwolony, jeśli nie zwiększa liczby
  naruszeń twardych, więc plan z naruszeniami da się naprawiać przeciąganiem. Szczegóły: `design/claude-design-brief.md`.

## Funkcja kosztu

`koszt(plan) = Σ waga_k × kara_k(plan)`. W trybie normalnym twarde reguły 1–10 to filtr, a nie składnik kosztu.
W trybie awaryjnym dochodzi składnik `10 000 × liczba_naruszeń_twardych` (patrz wyżej).

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
| 1 | tydzień 1 | Walidator 10 reguł z testami tabelarycznymi; pełna lista naruszeń w formacie raportu |
| 2 | tygodnie 1–4 | Solver z funkcją kosztu: fazy 1–3, tryb awaryjny, benchmarki, strojenie wag. Poprawny plan dla 174 zajęć z `cmd/solve` po ok. 2 tygodniach, potem jakość |
| równolegle | tygodnie 1–3 | API i baza: migracje goose, sqlc, import, endpointy, pula workerów, SSE (z atrapą solvera) |
| równolegle | tygodnie 2–5 | Frontend MVP: dane, wagi, postęp, siatka, raport z naruszeniami i podświetleniem na siatce (najpierw na mocku z OpenAPI) |
| 3 | tydzień 5 | Integracja: prawdziwy solver w API, postęp przez SSE, walidator po każdym uruchomieniu, testy e2e |
| 4 | tydzień 6 | Dodatki i prezentacja: edycja ręczna, porównanie, eksport |

## Pomysły na kolejne iteracje (poza MVP)

Nie planujemy ich w etapach 0–4. Zapisane, żeby nie zamknąć sobie drogi w kontraktach.

### Własne ograniczenia użytkownika

Użytkownik dodaje w UI ograniczenia, których nie ma w danych wejściowych, i generuje plan od nowa.
Przykład: remont w części A, więc dostępne są tylko sale z oznaczeniem B (w naszych danych A i B
to oznaczenia sal w jednym budynku, patrz `notes.md`, ale mechanizm ma działać tak samo dla osobnych budynków).

- **Gotowe szablony zamiast dowolnych reguł:** niedostępność sal (wybór pojedynczo, według oznaczenia A/B,
  typu albo wyposażenia), niedostępność prowadzącego, blokada terminu dla rocznika lub grupy,
  zajęcia przypięte do sali lub terminu. Każdy szablon ma zakres: dni, sloty, tygodnie 1–14.
- **Twarde albo miękkie:** użytkownik wybiera, czy ograniczenie jest twarde (filtr w fazie 1 i reguła
  w walidatorze), czy miękkie (kryterium kosztu z wagą, np. „unikaj sal A”).
- **Kompilacja do istniejących mechanizmów:** szablony zmieniają domeny w fazie 1 (dostępność sal,
  wykluczenia prowadzących, blokady), więc solver nie potrzebuje nowego kodu dla każdego typu.
  Walidator dostaje te same ograniczenia jako dodatkowe reguły i wypisuje ich naruszenia w raporcie,
  jak dla reguł 1–10.
- **Podgląd skutków przed uruchomieniem:** np. „bez sal A zostaje 14 sal, 2 wykłady nie mają sali z ≥ 198 miejscami”.
  Przy niemożliwych ograniczeniach działa tryb awaryjny (ADR 0006).
- **Kontrakty:** ograniczenia użytkownika jako osobna lista obok danych wejściowych (nie zmieniamy
  `dane_planu.json`), wersjonowana razem z uruchomieniem, żeby plan dało się odtworzyć. Wymaga ADR
  przed implementacją.
