import type { Schematy } from "../../../api/klient.ts";
import { LICZBA_DNI } from "./typy.ts";

type DzienKalendarza = Schematy["DzienKalendarza"];

// Kolumna siatki w wybranym tygodniu: który numer wystąpienia tego dnia planu pokazujemy i kiedy to jest.
// `numer = null` oznacza brak zajęć w tym dniu (święto, przerwa), wtedy `uwaga` mówi dlaczego.
export type DzienTygodnia = {
  numer: number | null;
  data: string | null; // RRRR-MM-DD
  odrabianie: boolean;
  uwaga: string;
};

export type TydzienNaSiatce = { dni: DzienTygodnia[]; zakres: string };

const krotkaData = new Intl.DateTimeFormat("pl-PL", { day: "numeric", month: "2-digit" });
const nazwaDniaTygodnia = new Intl.DateTimeFormat("pl-PL", { weekday: "short" });

// Kalendarz semestru z danych (`kalendarz`). Tydzień w planie to n-te wystąpienie danego dnia planu,
// a nie tydzień kalendarzowy (notes.md): po świętach poniedziałek i piątek tego samego tygodnia
// kalendarzowego mogą mieć różne numery, a dzień odrabiania prowadzi zajęcia według planu innego dnia.
export class KalendarzSemestru {
  private readonly dni: DzienKalendarza[];

  constructor(kalendarz: DzienKalendarza[]) {
    this.dni = [...kalendarz].sort((a, b) => a.data.localeCompare(b.data));
  }

  static data(iso: string): string {
    return krotkaData.format(new Date(`${iso}T12:00:00`));
  }

  // Bieżący tydzień kalendarzowy (w weekend: najbliższy), ograniczony do semestru.
  obecny(dzis: Date = new Date()): TydzienNaSiatce {
    const poniedzialek = KalendarzSemestru.poniedzialek(dzis);
    const zajecia = this.dni.filter((d) => d.typ_dnia !== "wolne");
    const pierwszy = zajecia[0];
    const ostatni = zajecia.at(-1);
    let start = KalendarzSemestru.iso(poniedzialek);
    if (
      pierwszy &&
      start < KalendarzSemestru.iso(KalendarzSemestru.poniedzialek(new Date(`${pierwszy.data}T12:00:00`)))
    ) {
      start = KalendarzSemestru.iso(KalendarzSemestru.poniedzialek(new Date(`${pierwszy.data}T12:00:00`)));
    }
    if (ostatni && start > ostatni.data) {
      start = KalendarzSemestru.iso(KalendarzSemestru.poniedzialek(new Date(`${ostatni.data}T12:00:00`)));
    }

    return this.tydzienKalendarzowy(start);
  }

  // Tydzień o numerze n: n-te wystąpienie każdego dnia planu.
  numer(n: number): TydzienNaSiatce {
    const dni = Array.from({ length: LICZBA_DNI }, (_, i): DzienTygodnia => {
      const dzien = this.dni.find((d) => d.plan_dnia === i + 1 && d.nr_tygodnia === n);
      return {
        numer: n,
        data: dzien?.data ?? null,
        odrabianie: dzien?.typ_dnia === "odpracowanie",
        uwaga: dzien?.typ_dnia === "odpracowanie" ? dzien.opis : "",
      };
    });
    return { dni, zakres: KalendarzSemestru.zakres(dni) };
  }

  private tydzienKalendarzowy(poniedzialek: string): TydzienNaSiatce {
    const piatek = KalendarzSemestru.iso(
      new Date(new Date(`${poniedzialek}T12:00:00`).getTime() + 4 * 86_400_000),
    );
    const wTygodniu = this.dni.filter((d) => d.data >= poniedzialek && d.data <= piatek);
    const dni = Array.from({ length: LICZBA_DNI }, (_, i): DzienTygodnia => {
      const prowadzone = wTygodniu.find((d) => d.plan_dnia === i + 1 && d.typ_dnia !== "wolne");
      if (prowadzone) {
        const odrabianie = prowadzone.typ_dnia === "odpracowanie";
        return {
          numer: prowadzone.nr_tygodnia,
          data: prowadzone.data,
          odrabianie,
          uwaga: odrabianie
            ? `w ${nazwaDniaTygodnia.format(new Date(`${prowadzone.data}T12:00:00`))} ${KalendarzSemestru.data(prowadzone.data)}`
            : "",
        };
      }

      const wolny = wTygodniu.find((d) => new Date(`${d.data}T12:00:00`).getDay() === i + 1);
      return {
        numer: null,
        data: wolny?.data ?? null,
        odrabianie: false,
        uwaga: wolny?.typ_dnia === "wolne" ? wolny.opis : "brak zajęć w tym tygodniu",
      };
    });
    return { dni, zakres: KalendarzSemestru.zakres(dni) };
  }

  private static zakres(dni: DzienTygodnia[]): string {
    const daty = dni
      .map((d) => d.data)
      .filter((d): d is string => d !== null)
      .sort();
    if (daty.length === 0) {
      return "";
    }

    const od = KalendarzSemestru.data(daty[0]!);
    const doDnia = KalendarzSemestru.data(daty.at(-1)!);
    return od === doDnia ? od : `${od}–${doDnia}`;
  }

  // Poniedziałek tygodnia z datą; sobota i niedziela należą do następnego tygodnia.
  private static poniedzialek(data: Date): Date {
    const dzien = data.getDay(); // 0 = niedziela
    const przesuniecie = dzien === 0 ? 1 : dzien === 6 ? 2 : 1 - dzien;
    const wynik = new Date(data);
    wynik.setHours(12, 0, 0, 0);
    wynik.setDate(wynik.getDate() + przesuniecie);
    return wynik;
  }

  private static iso(data: Date): string {
    const r = data.getFullYear();
    const m = String(data.getMonth() + 1).padStart(2, "0");
    const d = String(data.getDate()).padStart(2, "0");
    return `${r}-${m}-${d}`;
  }
}
