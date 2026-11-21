import type { Schematy } from "../../../api/klient.ts";
import { daneZestawu } from "../../../mocks/dane.ts";

type Zajecia = Schematy["DaneWejsciowe"]["zajecia"][number];

const WSZYSTKIE = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14];

const zajecia = (id: string, zmiany: Partial<Zajecia>): Zajecia => ({
  ...daneZestawu.zajecia[0]!,
  id,
  ...zmiany,
});

// Mały, ale kompletny zestaw: rocznik z dwiema grupami ćwiczeniowymi, drugi rocznik, prowadzący z dezyderatami.
export const dane: Schematy["DaneWejsciowe"] = {
  ...daneZestawu,
  sale: [
    ...daneZestawu.sale,
    {
      ...daneZestawu.sale[0]!,
      id: "B151",
      nazwa: null,
      typ: "sala_dydaktyczna",
      pojemnosc: 16,
      pojemnosc_szacowana: true,
    },
  ],
  prowadzacy: [
    ...daneZestawu.prowadzacy,
    { id: "P07", stopien: "dr", nazwa: "dr P07", jednostka: "WFiIS", czlonek_rady_wydzialu: false },
  ],
  roczniki: [
    ...daneZestawu.roczniki,
    { ...daneZestawu.roczniki[0]!, id: "INF1-R2", nazwa: "Informatyka I st., rok II", liczba_studentow: 72 },
  ],
  grupy: [
    ...daneZestawu.grupy,
    {
      id: "INF1-R1-G1",
      rocznik: "INF1-R1",
      nazwa: "Grupa 1",
      typ: "grupa",
      podzial: "grupy ćwiczeniowe",
      grupa_nadrzedna: "INF1-R1-ALL",
      liczba_studentow: 24,
    },
    {
      id: "INF1-R1-G2",
      rocznik: "INF1-R1",
      nazwa: "Grupa 2",
      typ: "grupa",
      podzial: "grupy ćwiczeniowe",
      grupa_nadrzedna: "INF1-R1-ALL",
      liczba_studentow: 24,
    },
    {
      id: "INF1-R2-ALL",
      rocznik: "INF1-R2",
      nazwa: "Cały rocznik",
      typ: "rocznik",
      podzial: null,
      grupa_nadrzedna: null,
      liczba_studentow: 72,
    },
  ],
  kolizje_grup: [
    { grupa_a: "INF1-R1-ALL", grupa_b: "INF1-R1-G1", wspolni_studenci: 24 },
    { grupa_a: "INF1-R1-ALL", grupa_b: "INF1-R1-G2", wspolni_studenci: 24 },
  ],
  przedmioty: [
    ...daneZestawu.przedmioty,
    { id: "PRZ-INF1-R2-01", rocznik: "INF1-R2", nazwa: "Analiza matematyczna II" },
  ],
  zajecia: [
    daneZestawu.zajecia[0]!,
    zajecia("Z002", {
      forma: "K",
      grupy: ["INF1-R1-G1"],
      prowadzacy: ["P07"],
      przedmiot: "Analiza matematyczna I",
    }),
    zajecia("Z003", {
      forma: "K",
      grupy: ["INF1-R1-G2"],
      prowadzacy: ["P07"],
      przedmiot: "Analiza matematyczna I",
    }),
    zajecia("Z101", {
      rocznik: "INF1-R2",
      przedmiot_id: "PRZ-INF1-R2-01",
      przedmiot: "Analiza matematyczna II",
      grupy: ["INF1-R2-ALL"],
      prowadzacy: ["P07"],
    }),
  ],
  dezyderaty_terminy: [
    { prowadzacy: "P07", dzien: 5, slot: 1, typ: "wykluczony" },
    { prowadzacy: "P07", dzien: 5, slot: 2, typ: "wykluczony" },
    { prowadzacy: "P39", dzien: 2, slot: 9, typ: "rada_wydzialu" },
    { prowadzacy: "P39", dzien: 2, slot: 10, typ: "rada_wydzialu" },
  ],
};

// Poniedziałek: wykład 3–4 (cały rocznik) i dwie grupy równolegle 5–6; P07 ma okienko 7 we wtorek.
export const plan: Schematy["Plan"] = {
  meta: { wersja_formatu: 1, silnik: "test", seed: 1, koszt: 100, utworzono: "2026-11-21T10:00:00Z" },
  przypisania: [
    { zajecia_id: "Z001", dzien: 1, slot_start: 3, sala_id: "A281", tygodnie: WSZYSTKIE },
    { zajecia_id: "Z002", dzien: 1, slot_start: 5, sala_id: "B151", tygodnie: [2, 4, 6, 8, 10, 12, 14] },
    {
      zajecia_id: "Z003",
      dzien: 1,
      slot_start: 5,
      sala_id: "A281",
      tygodnie: [6, 7, 8, 9, 10, 11, 12, 13, 14],
    },
    { zajecia_id: "Z101", dzien: 2, slot_start: 8, sala_id: "A281", tygodnie: WSZYSTKIE },
  ],
};

export const raportZNaruszeniami: Schematy["Raport"] = {
  status: "z_naruszeniami",
  liczba_naruszen: 1,
  naruszenia: [
    {
      regula: 7,
      nazwa: "pojemnosc",
      zajecia_id: ["Z002"],
      opis: "Sala B151 ma 16 miejsc, potrzeba 24",
    },
  ],
  koszt_miekki: 100,
  koszt_kryteria: {},
  walidator: "internal/validator",
};
