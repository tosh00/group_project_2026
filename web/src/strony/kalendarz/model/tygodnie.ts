import { Format } from "../../../lib/format.ts";

export type OpisTygodni = { pelny: string; krotki: string };

const WSZYSTKIE = 14;

// Opis tygodni zajęć na kartę: pusty dla całego semestru, inaczej np. „co 2 tyg. (parz.)”, „tyg. 6–14”.
export class Tygodnie {
  static opis(tygodnie: readonly number[]): OpisTygodni {
    const lista = [...new Set(tygodnie)].sort((a, b) => a - b);
    if (lista.length === WSZYSTKIE) {
      return { pelny: "", krotki: "" };
    }
    if (lista.length === 7 && lista.every((t) => t % 2 === 0)) {
      return { pelny: "co 2 tyg. (parz.)", krotki: "parz." };
    }
    if (lista.length === 7 && lista.every((t) => t % 2 === 1)) {
      return { pelny: "co 2 tyg. (nieparz.)", krotki: "np." };
    }

    const pierwszy = lista[0]!;
    const ostatni = lista.at(-1)!;
    if (ostatni - pierwszy + 1 === lista.length) {
      return { pelny: `tyg. ${pierwszy}–${ostatni}`, krotki: `${pierwszy}–${ostatni}` };
    }
    if (lista.length === WSZYSTKIE - 1) {
      const brakujacy = Array.from({ length: WSZYSTKIE }, (_, i) => i + 1).find((t) => !lista.includes(t))!;
      return { pelny: `bez tyg. ${brakujacy}`, krotki: `bez ${brakujacy}` };
    }

    return {
      pelny: `${Format.ilosc(lista.length, "wybrany tydz.", "wybrane tyg.", "wybranych tyg.")}`,
      krotki: `${lista.length} tyg.`,
    };
  }

  // Do dymka: pełne słowa zamiast skrótów („Co drugi tydzień (parzyste)”, „Cały semestr”).
  static opisDlaLudzi(tygodnie: readonly number[]): string {
    const lista = [...new Set(tygodnie)].sort((a, b) => a - b);
    const { pelny } = Tygodnie.opis(lista);
    if (!pelny) {
      return "Cały semestr";
    }
    if (pelny.startsWith("co 2 tyg.")) {
      return `Co drugi tydzień (${pelny.includes("nieparz") ? "nieparzyste" : "parzyste"})`;
    }
    if (pelny.startsWith("tyg. ")) {
      return `Tygodnie ${pelny.slice(5)}`;
    }
    if (pelny.startsWith("bez tyg. ")) {
      return `Bez tygodnia ${pelny.slice(9)}`;
    }

    return `Tygodnie ${lista.join(", ")}`;
  }

  // Do panelu naruszeń: „tyg. 1–14” zamiast pustego opisu.
  static opisZawsze(tygodnie: readonly number[]): string {
    return Tygodnie.opis(tygodnie).pelny || `tyg. 1–${WSZYSTKIE}`;
  }
}
