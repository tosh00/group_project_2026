# Planer zajęć WFiIS: instrukcje dla agentów

Generator planu zajęć dla kierunku Informatyka (WFiIS UŁ, semestr zimowy 2026/27).
Projekt grupowy: kilka osób pracuje równolegle, każda z własnym agentem.
**Repozytorium to wspólna pamięć zespołu.** Ten plik czyta każdy agent na starcie sesji.

## Zanim zaczniesz

1. Przeczytaj `docs/team.md`: kto jest właścicielem którego modułu.
2. Przeczytaj `docs/decisions/`: przyjęte decyzje (ADR). **Nie podważaj ich i nie obchodź.**
   Jeśli uważasz, że decyzja jest zła, zaproponuj nowy ADR w PR.
3. Sprawdź `docs/open-questions.md`. Jeśli zadanie zależy od otwartej kwestii, zapytaj użytkownika,
   zamiast zgadywać.
4. Sprawdź zadanie w GitHub Issues (`gh issue view <nr>`). Jedno zadanie to jedna gałąź i jeden PR.

## Źródła prawdy

| Co | Gdzie |
|---|---|
| Treść zadania, 10 ograniczeń twardych, ograniczenia miękkie | `data/README.md` |
| Dane wejściowe (import z JSON) | `data/dane_planu.json` (to samo w `data/csv/` i `.xlsx`) |
| Plan architektury i algorytmu | `docs/plan.md` |
| Kontrakty między modułami | `docs/contracts/` (README opisuje, co jest kontraktem) |
| Decyzje | `docs/decisions/NNNN-*.md` |
| Zadania i status | GitHub Issues / Projects w `tosh00/group_project_2026` |

Katalog `data/` jest tylko do odczytu. Nie zmieniaj danych wejściowych.

## Zasady współpracy

- **Pracuj tylko w swoim module** (patrz `docs/team.md`). Zmiana w cudzym module wymaga PR
  z review właściciela.
- **Kontrakty zmieniamy tylko przez PR z review wszystkich zainteresowanych.** Kontraktami są:
  `internal/model/`, format `plan.json`, `api/openapi.yaml`, `migrations/`.
  Gdy zmieniasz kontrakt, zaktualizuj też `docs/contracts/`.
- Nowa decyzja architektoniczna (biblioteka, format, algorytm, waga) to nowy plik
  w `docs/decisions/` według `docs/decisions/0000-template.md`.
- Odkryłeś coś o danych lub problemie, co przyda się innym (np. pułapkę w danych)? Dopisz to
  do `docs/notes.md` w tym samym PR. Nie zostawiaj tego tylko w pamięci swojej sesji.
- Nie zapisuj ustaleń zespołu w prywatnej pamięci agenta (`~/.claude`), bo inni jej nie widzą.
- `internal/validator` **nie importuje** niczego z `internal/solver`. To niezależny dowód poprawności.
- Commity i PR-y opisuj po polsku albo po angielsku, ale spójnie w obrębie PR.

## Konwencje (wstępne, aż do przyjęcia ADR-ów)

- Backend i solver: Go 1.23+, `gofmt`, `golangci-lint`, testy tabelaryczne, `go test ./...`.
- Frontend: React + TypeScript + Vite w `web/`, Vitest.
- Baza: PostgreSQL 16, migracje goose w `migrations/`.
- Nazwy domenowe po polsku, tak jak w danych (`zajecia`, `sala`, `prowadzacy`), kod i identyfikatory bez polskich znaków.
- Czas: dzień 1–5 (pon–pt), slot 1–12, tydzień 1–14 jako maska bitowa `uint16` (bit 0 = tydzień 1).

## Struktura repozytorium (docelowa)

```
cmd/server/          serwer HTTP: API, SSE, pula workerów
cmd/solve/           CLI: solve -in data/dane_planu.json -out plan.json -seed 42
internal/model/      struktury danych wejściowych + import          [kontrakt]
internal/solver/     fazy 1–3, funkcja kosztu
internal/validator/  10 reguł twardych, bez importów z solver/
internal/api/        handlery, DTO, SSE
internal/store/      zapytania sqlc
migrations/          SQL dla goose                                   [kontrakt]
api/openapi.yaml     kontrakt backend ↔ frontend                     [kontrakt]
web/                 frontend React
docs/                plan, decyzje, kontrakty, notatki
data/                dane wejściowe (tylko do odczytu)
```
