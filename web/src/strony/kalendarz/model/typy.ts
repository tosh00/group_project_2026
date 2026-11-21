import type { Schematy } from "../../../api/klient.ts";

export type Forma = Schematy["Forma"];
export type Wymiar = "rocznik" | "grupa" | "prowadzacy" | "sala";
export type KolorKart = "grupa" | "forma";

// Stan filtrów kalendarza; trzymany w URL (useFiltryKalendarza), żeby dało się podlinkować widok.
export type Filtry = {
  wymiar: Wymiar;
  id: string;
  tydzien: number; // 0 = wszystkie tygodnie
  ukryteFormy: ReadonlySet<Forma>;
  tylkoNaruszenia: boolean;
  kolor: KolorKart;
};

// Zajęcia z danych połączone z ich przypisaniem w planie.
export type ZajeciaWPlanie = {
  id: string;
  przedmiot: string;
  forma: Forma;
  salaId: string;
  prowadzacy: string[];
  rocznikId: string;
  grupaId: string;
  dzien: number; // 1–5
  slotStart: number; // 1–12
  dlugosc: number;
  tygodnie: number[];
};

export type StatusPlanu = "naruszenia" | "poprawny" | "niesprawdzony";

export const FORMY: readonly Forma[] = ["W", "LI", "K", "P", "LEK", "SEM"];
export const LICZBA_SLOTOW = 12;
export const LICZBA_DNI = 5;
