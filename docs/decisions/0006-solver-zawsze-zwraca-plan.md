# 0006. Solver zawsze zwraca plan, walidator wypisuje naruszenia

- Status: przyjęta
- Data: 2026-10-24
- Decydują: zespół

## Kontekst

Dane mogą być sprzeczne albo zbyt ciasne (np. tłok w pracowniach 9–16, jedna sala dla wykładów łączonych).
Wtedy solver, który akceptuje tylko plany zgodne ze wszystkimi regułami twardymi, kończy się komunikatem
„nie udało się” i użytkownik nie wie, co poprawić. Wolimy dostać plan i listę konkretnych problemów,
które można naprawić ręcznie albo poprawić w danych.

## Decyzja

Solver zawsze zwraca plan, w którym każde zajęcia mają przypisanie. Jeśli nie da się spełnić wszystkich
reguł twardych, przechodzi w tryb awaryjny: wstawia zajęcia tam, gdzie łamią najmniej reguł, a wyżarzanie
najpierw usuwa naruszenia (kara 10 000 za każde), potem poprawia jakość. Status planu (`poprawny` albo
`z_naruszeniami`) ustala walidator, nie solver. Frontend pokazuje naruszenia na siatce i w raporcie.

## Rozważane opcje

- Zajęcia wstawione z naruszeniami (wybrana): na siatce są wszystkie zajęcia, widać, gdzie jest problem,
  i można go poprawić przeciąganiem. Minus: plan z naruszeniami wygląda jak kompletny, więc UI musi
  wyraźnie pokazywać status.
- Zajęcia nieprzypisane, wypisane osobno: plan zawiera tylko poprawne przypisania, czytelniej. Minus:
  na siatce brakuje zajęć i trudniej zobaczyć, gdzie mogłyby trafić.
- Oba warianty jako opcja przy uruchomieniu: więcej kodu i testów w solverze, walidatorze i UI,
  bez wyraźnej korzyści w MVP.

## Konsekwencje

- Solver: tryb awaryjny w fazie 2, kara za naruszenia w koszcie fazy 3. Format `plan.json` v1 (ADR 0003)
  się nie zmienia; `meta.koszt` to koszt miękki bez kary za naruszenia.
- Walidator: zwraca pełną listę naruszeń (reguła, zajęcia, termin, sala), a nie tylko pierwsze.
  Format raportu: `contracts/plan-format.md`.
- API: status uruchomienia `poprawny` / `z_naruszeniami` z liczbą naruszeń. Błąd solvera tylko dla
  niepoprawnych danych lub przerwania przed końcem konstrukcji.
- Frontend: znacznik statusu, podświetlenie zajęć z naruszeniami, lista naruszeń, filtr „tylko naruszenia”.
  Ruch ręczny jest dozwolony, jeśli nie zwiększa liczby naruszeń twardych.
- Raport zawsze rozdziela naruszenia twarde (plan niedopuszczalny) i koszt miękki (jakość planu).
