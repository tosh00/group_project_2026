import type { Schematy } from "../api/klient.ts";

// Przykładowe odpowiedzi API do pracy bez backendu i do testów. Typy z api/openapi.yaml,
// więc zmiana kontraktu bez poprawienia mocków nie przejdzie `pnpm typecheck`.

export const zdrowie: Schematy["Zdrowie"] = { status: "ok", baza: "ok", walidator: "atrapa" };

export const wagiDomyslne: Required<Schematy["Wagi"]> = {
  okienka_studentow: 10,
  przeciazony_dzien: 15,
  okienka_prowadzacych: 5,
  po_17: 4,
  przed_9: 3,
  za_duza_sala: 6,
  preferencja_dni: 8,
  nie_po_17: 20,
  dzien_wolny: 10,
  przerwa_max_2h: 8,
  pora_dnia: 3,
  sale_a: 3,
  wyklady_rocznika: 2,
};

export const zestaw: Schematy["Zestaw"] = {
  id: "7d2f0c1e-4b5a-4c8e-9f10-2a3b4c5d6e7f",
  nazwa: "Informatyka, semestr zimowy 2026/27",
  utworzono: "2026-11-21T09:30:00Z",
  podsumowanie: {
    zajecia: 174,
    sale: 55,
    sale_do_planowania: 30,
    prowadzacy: 50,
    roczniki: 6,
    grupy: 51,
    studenci: 372,
  },
  ostrzezenia: [
    { kod: "pojemnosc_szacowana", opis: "Sala B151: pojemność szacowana (16 miejsc).", sala_id: "B151" },
    { kod: "pojemnosc_szacowana", opis: "Sala B153: pojemność szacowana (16 miejsc).", sala_id: "B153" },
    { kod: "pojemnosc_szacowana", opis: "Sala B472: pojemność szacowana (16 miejsc).", sala_id: "B472" },
  ],
};

export const uruchomienie: Schematy["Uruchomienie"] = {
  id: "0b9a8c7d-6e5f-4a3b-8c2d-1e0f9a8b7c6d",
  zestaw_id: zestaw.id,
  status: "poprawny",
  wagi: wagiDomyslne,
  seed: 42,
  iteracje: 1_000_000,
  zatrzymane: false,
  liczba_naruszen: 0,
  koszt: 412.5,
  blad: null,
  utworzono: "2026-11-21T09:31:00Z",
  rozpoczeto: "2026-11-21T09:31:00Z",
  zakonczono: "2026-11-21T09:31:04Z",
};

// Mały wycinek data/dane_planu.json: jeden rocznik, jedne zajęcia z wszystkimi tabelami, do których się odwołują.
export const daneZestawu: Schematy["DaneWejsciowe"] = {
  meta: {
    opis: "Wycinek danych do mocków",
    wersja: "1.0",
    data: "2026-10-23",
    tabele: {},
  },
  kalendarz: [
    {
      data: "2026-10-05",
      dzien_tygodnia: "poniedziałek",
      typ_dnia: "zajęcia",
      plan_dnia: 1,
      nr_tygodnia: 1,
      opis: "",
    },
  ],
  sloty: [
    { slot: 1, od: "8:15", do: "9:00" },
    { slot: 2, od: "9:15", do: "10:00" },
  ],
  sale: [
    {
      id: "A281",
      nazwa: "Duża Aula",
      budynek: "A",
      typ: "aula",
      pojemnosc: 264,
      pojemnosc_szacowana: false,
      metraz_m2: 300,
      wyposazenie: ["projektor"],
      przeznaczenie: null,
      dostepna_do_planowania: true,
      uwagi: null,
    },
  ],
  prowadzacy: [
    { id: "P39", stopien: "prof.", nazwa: "prof. P39", jednostka: "WFiIS", czlonek_rady_wydzialu: true },
  ],
  roczniki: [
    {
      id: "INF1-R1",
      kierunek: "Informatyka",
      stopien_studiow: "I",
      rok: 1,
      semestr: 1,
      nazwa: "Informatyka I st., rok I, sem. I",
      liczba_studentow: 118,
      syntetyczny: false,
    },
  ],
  grupy: [
    {
      id: "INF1-R1-ALL",
      rocznik: "INF1-R1",
      nazwa: "Cały rocznik",
      typ: "rocznik",
      podzial: null,
      grupa_nadrzedna: null,
      liczba_studentow: 118,
    },
  ],
  studenci: [],
  studenci_grupy: [],
  kolizje_grup: [],
  przedmioty: [{ id: "PRZ-INF1-R1-01", rocznik: "INF1-R1", nazwa: "Analiza matematyczna I" }],
  zajecia: [
    {
      id: "Z001",
      rocznik: "INF1-R1",
      przedmiot_id: "PRZ-INF1-R1-01",
      przedmiot: "Analiza matematyczna I",
      forma: "W",
      forma_opis: "wykład",
      grupy: ["INF1-R1-ALL"],
      liczba_studentow: 118,
      laczone_z: null,
      dodatkowi_studenci: 0,
      prowadzacy: ["P39"],
      godziny_programowe: 28,
      liczba_spotkan: 14,
      sloty_na_spotkanie: 2,
      czestotliwosc: "co_tydzien",
      tydzien_od: 1,
      tydzien_do: 14,
      typy_sal: ["aula", "sala_dydaktyczna"],
      dozwolone_sale: [],
      wymagane_wyposazenie: [],
      tryb: "stacjonarny",
      powiazanie: null,
      typ_powiazania: null,
      tygodnie_niedostepne: [],
      uwagi: null,
    },
  ],
  blokady_stale: [],
  rada_wydzialu: [{ dzien: 2, slot_od: 9, slot_do: 10, opis: "Rada Wydziału" }],
  dezyderaty_terminy: [],
  dezyderaty_ogolne: [],
  dezyderaty_sale: [],
};
