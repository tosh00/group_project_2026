# Brief dla Claude Design: Planer zajęć WFiIS

Wklej treść poniżej kreski do Claude Design jako pierwszą wiadomość.

---

## Produkt

Aplikacja webowa, która generuje tygodniowy plan zajęć dla kierunku Informatyka (Wydział Fizyki i Informatyki
Stosowanej UŁ, semestr zimowy 2026/27). Użytkownik wgrywa plik JSON z danymi, backend układa plan algorytmem,
a aplikacja pokazuje go jako czytelny kalendarz, który można poprawiać ręcznie przez przeciąganie.

Użytkownik: osoba układająca plan na wydziale (pracownik dziekanatu / koordynator). Zna dane, nie zna algorytmu.
Pracuje na laptopie lub dużym monitorze; telefon tylko do podglądu planu. Język interfejsu: polski.

## Główny przepływ (3 kroki)

1. **Wgranie danych.** Przeciągnij lub wybierz plik `.json` (tylko ten format). Aplikacja pokazuje podsumowanie
   (174 zajęcia, 55 sal (30 do planowania), 50 prowadzących, 6 roczników, 51 grup) i ostrzeżenia z walidacji pliku
   (np. „sala B151: pojemność szacowana”). Błędy blokujące (niepoprawny JSON, brak wymaganej tabeli lub pola) wyraźnie, z nazwą pola.
2. **Generowanie.** Przycisk „Generuj plan”. Opcjonalnie panel wag: suwaki dla kryteriów jakości
   (okienka studentów, zajęcia po 17:00, za duża sala, preferencje prowadzących…). W trakcie: postęp na żywo
   (faza: przygotowanie → konstrukcja → optymalizacja, wykres malejącego kosztu, czas), przycisk „Zatrzymaj”
   (zachowuje najlepszy dotychczasowy plan). Trwa od kilku sekund do ok. minuty.
3. **Kalendarz.** Wynik jako siatka tygodnia z możliwością ręcznej edycji przez drag & drop.

## Kalendarz: najważniejszy ekran

- Siatka **5 dni (pon–pt) × 12 slotów**. Sloty 45 min z 15 min przerwy: 1 = 8:15–9:00, 2 = 9:15–10:00, …,
  12 = 19:15–20:00. Sloty 1 i 10–12 są „niechciane” (przed 9:00, po 17:00): delikatnie wyróżnione tło.
- Zajęcia trwają 1–4 slotów (bloki różnej wysokości). Karta zajęć: przedmiot, forma (W wykład, LI laboratorium,
  K konwersatorium, P projekt, LEK lektorat, SEM seminarium), sala, prowadzący, grupa. Kolor według formy
  albo rocznika (do wyboru).
- **Widok według filtra:** rocznik, grupa, prowadzący lub sala. Plan jednego rocznika na raz, bo cały wydział
  w jednej siatce jest nieczytelny. Kilka zajęć w tym samym slocie w widoku rocznika (różne grupy
  laboratoryjne równolegle) to norma: karty obok siebie w kolumnie dnia.
- **Tygodnie 1–14.** Zajęcia nie zawsze trwają cały semestr: „co 2 tygodnie” (tyg. parzyste / nieparzyste),
  „tyg. 6–14”, „bez tyg. 6”. Na karcie mała etykieta tygodni. Przełącznik tygodnia 1–14 (z datą) pokazuje
  tylko zajęcia z danego tygodnia; domyślnie widok „wszystkie tygodnie”.
- Pasek nad siatką: koszt planu (jedna liczba + trend po edycji), znacznik „0 naruszeń” (zielony), eksport.

## Drag & drop: zasady (kluczowe)

Reguły planu dzielą się na dwie grupy i UI musi je traktować różnie:

- **Reguły twarde (10 sztuk): nie wolno ich złamać.** Np. prowadzący w dwóch miejscach naraz, dwa zajęcia
  w jednej sali, kolizja studentów, sala za mała / złego typu, godzina wykluczona przez prowadzącego,
  Rada Wydziału, stałe blokady (WF). Przeciągnięcie, które by je złamało, **jest niemożliwe**: pole jest
  zablokowane, karta wraca na miejsce.
- **Kryteria miękkie (wagi): mogą się pogorszyć.** Ruch dozwolony nawet jeśli plan staje się gorszy
  (więcej okienek, zajęcia po 17:00). Użytkownik widzi, o ile zmienił się koszt.

Zachowanie podczas przeciągania:

- Po chwyceniu karty siatka od razu pokazuje mapę dozwolonych pól: dozwolone (neutralne / zielonkawe),
  zabronione (wyszarzone, przekreślone). Najechanie na zabronione pole pokazuje **powód** w jednym zdaniu
  („dr P12 ma wtedy zajęcia w B153”, „sala A225 ma 12 miejsc, potrzeba 30”).
- Na dozwolonym polu podgląd zmiany kosztu: „+12 (okienko grupy G2, zajęcia po 17:00)” albo „−4”.
- Zmiana sali bez zmiany terminu: po upuszczeniu lub z menu karty, lista dozwolonych sal z pojemnością.
- Po upuszczeniu: toast z „Cofnij”, historia zmian, możliwość **przypięcia** zajęć (ikona pinezki),
  żeby ponowne „Generuj plan” ich nie ruszało.
- Powiązane zajęcia przesuwają się razem (te, które muszą mieć ten sam termin), a zajęcia z zestawu
  „różne dni” nie mogą trafić w ten sam dzień: to też jest reguła twarda.

## Pozostałe ekrany

- **Raport jakości:** 0 naruszeń twardych, koszt całkowity, rozbicie kosztu na kryteria (wykres słupkowy),
  najgorsze grupy i prowadzący (lista z linkiem do kalendarza z filtrem).
- **Porównanie (dodatkowo):** dwa wygenerowane plany obok siebie, różnice zaznaczone.
- **Eksport (dodatkowo):** CSV, XLSX, ICS.

## Stany do zaprojektowania

Pusty stan (brak danych), wgrywanie, błąd pliku, generowanie w toku, zatrzymane, brak planu (błąd solvera),
kalendarz bez zajęć dla filtra, przeciąganie (dozwolone / zabronione / podgląd kosztu), konflikt przy cofnięciu.

## Kierunek wizualny

Narzędzie robocze, nie strona marketingowa: gęste, spokojne, czytelne przy 30+ kartach na ekranie.
Jasny i ciemny motyw. Kolory semantyczne: czerwony tylko dla reguł twardych, zielony dla „0 naruszeń”,
koszt neutralnie. Wysoki kontrast tekstu na kartach, dostępność kolorów (forma rozróżniona nie tylko kolorem).

## Stack (dla realizacji)

React 18 + TypeScript + Vite, TanStack Query, React Router, dnd-kit do drag & drop. Backend w Go zwraca
plan i na żywo sprawdza ruch (`POST /api/runs/{id}/sprawdz-ruch` → dozwolony?, powód, zmiana kosztu).

## Przykładowe dane do makiet

- Rocznik „Informatyka I st., rok I” (118 studentów), grupy INF1-R1-G1…G5.
- „Analiza matematyczna I”, W, Duża Aula A281, dr P07, pon 3–4, tyg. 1–14.
- „Programowanie”, LI, pracownia komputerowa A121, dr P21, grupa G2, śr 5–7, co 2 tyg. (parzyste).
- „Fizyka”, K, A233, dr hab. P03, grupa G1, czw 8–9, tyg. 6–14.
- Rada Wydziału: wt 9–10 (blokada dla członków rady); WF rocznika II: czw 1–4.
