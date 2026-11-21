import type { Forma } from "./typy.ts";

// Kolory z makiety „2 · Kalendarz”. Karta ma tło z koloru (13%), ramkę (40%) i pasek u góry (100%).
export const KOLORY_FORM: Record<Forma, string> = {
  W: "#4338CA",
  LI: "#0369A1",
  K: "#7E22CE",
  P: "#A21CAF",
  LEK: "#475569",
  SEM: "#0F766E",
};

export const NAZWY_FORM: Record<Forma, string> = {
  W: "wykład",
  LI: "laboratorium",
  K: "konwersatorium",
  P: "projekt",
  LEK: "lektorat",
  SEM: "seminarium",
};

// Paleta grup i roczników: pierwsze kolory jak w makiecie, dalej kolejne dla roczników z wieloma grupami.
const PALETA = [
  "#4338CA",
  "#0284C7",
  "#7C3AED",
  "#0D9488",
  "#C026D3",
  "#475569",
  "#B45309",
  "#BE123C",
  "#15803D",
  "#1D4ED8",
  "#9333EA",
  "#0E7490",
];

export class Kolory {
  static zPalety(indeks: number): string {
    return PALETA[indeks % PALETA.length]!;
  }

  static rgba(hex: string, alfa: number): string {
    const n = parseInt(hex.slice(1), 16);
    return `rgba(${n >> 16}, ${(n >> 8) & 255}, ${n & 255}, ${alfa})`;
  }
}
