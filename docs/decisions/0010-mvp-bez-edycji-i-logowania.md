# 0010. MVP bez ręcznej edycji planu i bez logowania

- Status: proponowana
- Data: 2026-11-18
- Decydują: @KajetanJanusz, @shokera, @krzysiu-w

## Kontekst

Otwarte kwestie #7 (ręczna edycja planu w MVP czy jako dodatek) i #8 (logowanie użytkowników) blokują
zakres API i frontendu. Etapy w `plan.md` stawiają edycję ręczną w etapie 4 (dodatki).

## Decyzja

W MVP API nie ma endpointów edycji planu (`sprawdz-ruch`, zapis przesunięć, przypinanie) ani logowania.
Aplikacja działa dla jednego, niezalogowanego użytkownika. Obie rzeczy wracają jako dodatki po integracji (etap 4).

## Rozważane opcje

- **Edycja w MVP:** zgodne z briefem frontendu, ale wymaga szybkiego sprawdzania ruchu (walidator przyrostowo)
  i wersjonowania planu, zanim działa ścieżka podstawowa.
- **Edycja jako dodatek (wybrane):** najpierw działający przepływ wgranie → generowanie → kalendarz → raport.
- **Logowanie:** niepotrzebne dla jednego koordynatora na prezentacji; dokłada sesje, hasła i tabelę użytkowników.

## Konsekwencje

- `api/openapi.yaml` w wersji 0.1 nie ma tras edycji; brief w `docs/design/` dalej je opisuje jako cel docelowy.
- Pole `zablokowane` w `plan.json` zostaje w formacie (ADR 0003), ale w MVP zawsze ma wartość `false`.
- Z `docs/open-questions.md` znikają kwestie #7 i #8.
