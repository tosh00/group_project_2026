# Dane wejściowe: generator planu zajęć

Wydział Fizyki i Informatyki Stosowanej UŁ, kierunek Informatyka, studia stacjonarne, semestr zimowy 2026/27.

## 1. Zadanie

Zaprojektujcie i zbudujcie system, który z tych danych wygeneruje plan zajęć na semestr. Każde zajęcia z tabeli `zajecia` muszą dostać dzień, godzinę, salę i (dla zajęć co 2 tygodnie) tydzień parzysty lub nieparzysty. Plan musi spełniać wszystkie ograniczenia twarde (sekcja 5). Ograniczenia miękkie (sekcja 6) służą do oceny jakości planu.

Dane są spójne: istnieje co najmniej jeden plan, który spełnia wszystkie ograniczenia twarde.

## 2. Pliki

| Plik | Zawartość |
|---|---|
| `dane_planu.xlsx` | wszystkie tabele, jeden arkusz na tabelę; arkusz `_spis` opisuje każdą tabelę |
| `csv/*.csv` | te same tabele jako CSV (UTF-8, separator `,`; listy w komórce rozdzielone `;`) |
| `dane_planu.json` | te same tabele w jednym pliku JSON (listy jako tablice JSON) |

We wszystkich trzech formatach dane są identyczne. Wybierzcie format, który Wam pasuje.

## 3. Czas: sloty, dni i tygodnie

**Sloty.** Dzień ma 12 slotów po 45 minut, z 15-minutową przerwą (`sloty`):
slot 1 = 8:15–9:00, slot 2 = 9:15–10:00, …, slot 12 = 19:15–20:00.
Zajęcia trwające 2 sloty zajmują np. sloty 3–4 (10:15–12:00).

**Dni.** Dni 1–5 to poniedziałek–piątek.

**Tygodnie.** Każdy dzień planu odbywa się dokładnie 14 razy w semestrze. `nr_tygodnia` (1–14) to numer kolejnego wystąpienia danego dnia planu, a nie numer tygodnia w kalendarzu. Szczegóły są w tabeli `kalendarz`:
- dni wolne: 1 X (inauguracja), 2 XI (dzień rektorski), 11 XI (święto), 23 XII–6 I (przerwa świąteczna),
- dni odpracowania: 8 I (piątek) i 26 I (wtorek) mają plan **środy**, 25 I to dodatkowy **poniedziałek**.

Przykład: dla środy tydzień 11 to 8 I 2027 (piątek), a tydzień 14 to 26 I 2027 (wtorek).

Plan jest tygodniowy. Jeden wiersz `zajecia` to cykl spotkań, które odbywają się zawsze w tym samym dniu planu, o tej samej godzinie i w tej samej sali.

## 4. Tabele

### `zajecia`: co trzeba zaplanować (174 wiersze)

| Kolumna | Znaczenie |
|---|---|
| `id` | identyfikator, np. `Z042` |
| `rocznik`, `przedmiot_id`, `przedmiot` | do czego należą zajęcia |
| `forma` | W wykład, K konwersatorium/ćwiczenia, LI laboratorium informatyczne, P pracownia, LEK lektorat, SEM seminarium |
| `grupy` | grupy studentów, które uczestniczą (zwykle jedna) |
| `liczba_studentow` | liczba studentów w tych grupach |
| `laczone_z`, `dodatkowi_studenci` | zajęcia wspólne z innym kierunkiem (np. Fizyka): dodatkowe osoby, które też muszą zmieścić się w sali |
| `prowadzacy` | jedna lub więcej osób; **wszystkie** muszą być wolne |
| `godziny_programowe` | liczba godzin w programie studiów (informacyjnie) |
| `liczba_spotkan` | ile spotkań w cyklu |
| `sloty_na_spotkanie` | długość jednego spotkania w slotach |
| `czestotliwosc` | `co_tydzien`: każdy tydzień od `tydzien_od` do `tydzien_do`; `co_2_tygodnie`: co drugi tydzień w tym zakresie (sami wybieracie: parzyste lub nieparzyste); `wybrane_tygodnie`: sami wybieracie `liczba_spotkan` tygodni z zakresu |
| `tydzien_od`, `tydzien_do` | dozwolony zakres tygodni (1–14) |
| `typy_sal` | dozwolone typy sal (gdy `dozwolone_sale` jest puste) |
| `dozwolone_sale` | jeśli niepuste, tylko te sale są dozwolone |
| `wymagane_wyposazenie` | sala musi mieć to wyposażenie (np. `MAC`, `CISCO`) |
| `tryb` | `stacjonarny` lub `hybrydowy` (hybrydowe też potrzebują sali) |
| `powiazanie`, `typ_powiazania` | wiersze z tym samym `powiazanie` tworzą zestaw: `rozne_dni`: spotkania muszą być w różnych dniach; `ten_sam_termin`: ten sam dzień, godzina i sala (np. wykład, który przejmuje drugi prowadzący w połowie semestru) |
| `tygodnie_niedostepne` | tygodnie, w których te zajęcia nie mogą się odbyć |
| `uwagi` | wyjaśnienia |

Przykład: `liczba_spotkan = 9`, `sloty_na_spotkanie = 2`, `co_tydzien`, tygodnie `6–14` oznacza 9 spotkań po 2 sloty, co tydzień od tygodnia 6.

### Studenci i grupy

- `roczniki`: 6 roczników (I st.: sem. I, III, V, VII; II st.: sem. I, III).
- `grupy`: grupy studentów. Każdy rocznik ma grupę `…-ALL` (cały rocznik). `podzial` to sposób podziału rocznika (np. grupy laboratoryjne, specjalność, przedmiot do wyboru). Grupy z tego samego podziału są rozłączne. `grupa_nadrzedna` pokazuje zagnieżdżenie (np. grupa A1 jest w specjalności A).
- `studenci`, `studenci_grupy`: każdy student i jego grupy. **To jest źródło prawdy o kolizjach.** Dwie grupy kolidują, jeśli mają choć jednego wspólnego studenta.
- `kolizje_grup`: pary grup ze wspólnymi studentami, gotowe wyliczenie ze `studenci_grupy`.

Uwaga: grupy z różnych podziałów nie zawsze kolidują. Na przykład na roczniku `INF1-R2` część grup pracowni (`GR…`) nie ma wspólnych studentów z niektórymi grupami laboratoryjnymi. Sprawdzajcie zawsze `studenci_grupy` albo `kolizje_grup`.

Przedmioty do wyboru (`typ = opcja`): każdy student wybiera dokładnie jedną opcję.

### Sale (`sale`)

- `typ`: `aula`, `sala_dydaktyczna`, `pracownia_komputerowa`, `pracownia_robotyki`, `pracownia_specjalistyczna`.
- `pojemnosc`: liczba miejsc. `pojemnosc_szacowana = true` oznacza, że w źródle nie było tej liczby.
- `wyposazenie`: lista, np. `projektor`, `MAC`, `CISCO`, `robotyka`.
- `dostepna_do_planowania = false`: tej sali nie używamy (pracownie fizyczne, sala dla studentów), chyba że jest w `dozwolone_sale`.

Duża Aula to `A281`, Mała Aula to `A283`.

### Prowadzący i dezyderaty

- `prowadzacy`: pseudonimy (`P01` …). `czlonek_rady_wydzialu = true` oznacza, że osoba ma zablokowany termin Rady Wydziału.
- `rada_wydzialu`: stały termin Rady Wydziału i Senatu.
- `dezyderaty_terminy`: z druku dezyderatów; jeden wiersz = jedna godzina (dzień + slot), w której osoba nie może prowadzić zajęć. `typ` = `wykluczony` (max 12 godzin na osobę) lub `rada_wydzialu`.
- `dezyderaty_ogolne`: powód wykluczeń, preferencja (`skumulowane`: zajęcia w jak najmniejszej liczbie dni; `nieskumulowane`: rozłożone na kilka dni; `brak`) oraz inne uwagi.
- `dezyderaty_sale`: wymagania co do sali i sprzętu (te same wymagania są już w `zajecia.dozwolone_sale` / `wymagane_wyposazenie`).

### Inne

- `blokady_stale`: terminy ustalone poza wydziałem (np. WF). Grupa jest wtedy zajęta.
- `przedmioty`: lista przedmiotów.

## 5. Ograniczenia twarde (muszą być spełnione)

Dwa spotkania **nakładają się**, jeśli są w tym samym dniu planu, mają wspólny slot i wspólny tydzień.

1. Każde zajęcia mają dokładnie jeden dzień, slot startowy i salę. Wszystkie sloty mieszczą się w zakresie 1–12.
2. Liczba spotkań, zakres tygodni, częstotliwość i `tygodnie_niedostepne` są zachowane.
3. Prowadzący nie ma dwóch nakładających się zajęć.
4. Sala nie ma dwóch nakładających się zajęć.
5. Student nie ma dwóch nakładających się zajęć (grupy ze wspólnymi studentami).
6. Sala ma dozwolony typ (lub jest w `dozwolone_sale`), ma wymagane wyposażenie i jest dostępna do planowania.
7. Pojemność sali ≥ `liczba_studentow` + `dodatkowi_studenci`.
8. Zajęcia nie trafiają w godziny wykluczone przez prowadzącego ani w Radę Wydziału (dla członków rady).
9. Zajęcia nie nakładają się na `blokady_stale` swoich studentów.
10. Powiązania (`rozne_dni`, `ten_sam_termin`) są zachowane.

## 6. Ograniczenia miękkie (jakość planu)

Propozycje. Możecie dodać własne i ustalić wagi.

- preferencja prowadzącego `skumulowane` / `nieskumulowane`,
- mało „okienek” (przerw) w dniu studenta i prowadzącego,
- mało zajęć po 17:00 i przed 9:00,
- rozsądne obciążenie dnia studenta (np. nie więcej niż 8 slotów),
- sala niezbyt duża dla grupy (dobre wykorzystanie sal),
- uwagi z `dezyderaty_ogolne.inne_uwagi`.

## 7. Skąd są dane i co uproszczono

- Zajęcia, prowadzący i sale pochodzą z prawdziwego planu wydziału na semestr zimowy 2026/27 oraz z zestawienia sal z 20.10.2026.
- **Nazwiska zastąpiono pseudonimami.**
- **Rocznik `INF2-R1` (II st., sem. I) jest syntetyczny**: przedmioty i obsada są wymyślone, ale realistyczne.
- **Liczby studentów są szacunkowe**, dobrane do pojemności sal w obecnym planie. Przynależność do grup jest wygenerowana.
- **Dezyderaty są syntetyczne**, ale zgodne z formatem druku (wykluczenia max 12 godzin, Rada Wydziału, preferencje, sale).
- Godziny z planu zaokrąglono do siatki slotów (np. 8:30 → slot 1, 15:30 → slot 8).
- Dla przedmiotów bez podanej liczby godzin (Ekonomia, Seminarium dyplomowe) przyjęto 28 h.
- Kalendarz pochodzi z kalendarza akademickiego UŁ 2026/27. Dni odpracowania (8 I, 25 I, 26 I) odtworzono z planu wydziału.
- Dane obejmują tylko kierunek Informatyka. Inne kierunki też korzystają z sal, ale tu tego nie uwzględniono. Wyjątek: zajęcia łączone z Fizyką (`laczone_z`).
- Studia niestacjonarne nie wchodzą w zakres zadania.
