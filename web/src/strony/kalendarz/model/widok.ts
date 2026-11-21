import { Format } from "../../../lib/format.ts";
import type { IndeksPlanu } from "./indeks.ts";
import { KOLORY_FORM } from "./kolory.ts";
import type { NaruszeniaPlanu } from "./naruszenia.ts";
import type { DzienTygodnia, TydzienNaSiatce } from "./semestr.ts";
import { UkladTorow } from "./uklad.ts";
import { LICZBA_DNI, LICZBA_SLOTOW, type Filtry, type Wymiar, type ZajeciaWPlanie } from "./typy.ts";

export const NAZWY_DNI = ["Poniedziałek", "Wtorek", "Środa", "Czwartek", "Piątek"];
export const SKROTY_DNI = ["pon", "wt", "śr", "czw", "pt"];

export const NAZWY_TYPOW_SAL: Record<string, string> = {
  aula: "aula",
  sala_dydaktyczna: "sala dydaktyczna",
  pracownia_komputerowa: "pracownia komputerowa",
  pracownia_robotyki: "pracownia robotyki",
  pracownia_specjalistyczna: "pracownia specjalistyczna",
};

export type KartaNaSiatce = {
  zajecia: ZajeciaWPlanie;
  tor: number;
  torow: number;
  kolor: string;
  etykietaGrupy: string; // „INF1-R1-G2” albo „INF1-R1 · G2” poza widokami studentów
  skrotGrupy: string;
  naruszona: boolean;
  wygaszona: boolean;
};

export type BlokNaSiatce = {
  rodzaj: "wykluczone" | "okienko" | "wolny" | "bezZajec";
  slotOd: number;
  slotDo: number;
  tytul: string;
  opis: string;
};

export type DzienNaSiatce = {
  dzien: number;
  karty: KartaNaSiatce[];
  bloki: BlokNaSiatce[];
  wolny: boolean;
  kalendarz: DzienTygodnia | null; // data i numer wystąpienia w wybranym tygodniu
};

export type OpisWidoku = {
  okruszek: string;
  nazwa: string;
  podpis: string;
  tytul: string;
  podtytul: string;
};

export type StatystykiProwadzacego = { slotow: number; dniWolne: string[]; okienek: number };

// Co pokazuje siatka dla danego filtra. Czysta logika, testowana bez Reacta (widok.test.ts).
export class WidokPlanu {
  readonly zajeciaWidoku: ZajeciaWPlanie[];
  readonly widoczne: ZajeciaWPlanie[];
  readonly widokStudentow: boolean;

  constructor(
    private readonly indeks: IndeksPlanu,
    private readonly naruszenia: NaruszeniaPlanu,
    private readonly filtry: Filtry,
    // Wybrany tydzień z kalendarza semestru; bez niego siatka pokazuje zajęcia ze wszystkich tygodni.
    private readonly tydzien?: TydzienNaSiatce,
  ) {
    this.widokStudentow = filtry.wymiar === "rocznik" || filtry.wymiar === "grupa";
    this.zajeciaWidoku = WidokPlanu.zajeciaDla(indeks, filtry.wymiar, filtry.id);
    this.widoczne = this.zajeciaWidoku.filter((z) => !filtry.ukryteFormy.has(z.forma) && this.wTygodniu(z));
  }

  // Każdy dzień planu ma w danym tygodniu własny numer wystąpienia (KalendarzSemestru).
  private wTygodniu(z: ZajeciaWPlanie): boolean {
    if (!this.tydzien) {
      return true;
    }

    const numer = this.tydzien.dni[z.dzien - 1]?.numer ?? null;
    return numer !== null && z.tygodnie.includes(numer);
  }

  static zajeciaDla(indeks: IndeksPlanu, wymiar: Wymiar, id: string): ZajeciaWPlanie[] {
    switch (wymiar) {
      case "rocznik":
        return indeks.zajecia.filter((z) => z.rocznikId === id);
      case "grupa": {
        const grupy = indeks.grupyPowiazane(id);
        return indeks.zajecia.filter((z) => grupy.has(z.grupaId));
      }
      case "prowadzacy":
        return indeks.zajecia.filter((z) => z.prowadzacy.includes(id));
      case "sala":
        return indeks.zajecia.filter((z) => z.salaId === id);
    }
  }

  dni(): DzienNaSiatce[] {
    const wolne = this.filtry.wymiar === "prowadzacy" ? this.dniWolneProwadzacego() : new Set<number>();
    return Array.from({ length: LICZBA_DNI }, (_, i) => i + 1).map((dzien) => {
      const zajeciaDnia = this.widoczne.filter((z) => z.dzien === dzien);
      const karty = UkladTorow.rozloz(zajeciaDnia, (z) => ({ start: z.slotStart, dlugosc: z.dlugosc })).map(
        ({ element, tor, torow }) => this.karta(element, tor, torow),
      );
      const kalendarz = this.tydzien?.dni[dzien - 1] ?? null;
      if (kalendarz && kalendarz.numer === null) {
        const bezZajec: BlokNaSiatce = {
          rodzaj: "bezZajec",
          slotOd: 1,
          slotDo: LICZBA_SLOTOW,
          tytul: "Brak zajęć",
          opis: kalendarz.uwaga,
        };
        return { dzien, karty, bloki: [bezZajec], wolny: false, kalendarz };
      }

      const bloki =
        this.filtry.wymiar === "prowadzacy" ? this.blokiProwadzacego(dzien, wolne.has(dzien)) : [];
      return { dzien, karty, bloki, wolny: wolne.has(dzien), kalendarz };
    });
  }

  opis(): OpisWidoku {
    const { wymiar, id } = this.filtry;
    const liczba = Format.ilosc(this.zajeciaWidoku.length, "zajęcia", "zajęcia", "zajęć");
    switch (wymiar) {
      case "rocznik": {
        const r = this.indeks.roczniki.find((x) => x.id === id);
        const grup = (this.indeks.grupyRocznika.get(id) ?? []).filter((g) => g.typ !== "rocznik").length;
        return {
          okruszek: `Rocznik ${id}`,
          nazwa: id,
          podpis: `${r?.nazwa ?? ""} · ${r?.liczba_studentow ?? 0} os.`,
          tytul: `Plan rocznika ${id}`,
          podtytul: `${r?.nazwa ?? ""} · ${Format.ilosc(r?.liczba_studentow ?? 0, "student", "studentów", "studentów")} · ${Format.ilosc(grup, "grupa", "grupy", "grup")} · ${liczba}`,
        };
      }
      case "grupa": {
        const g = this.indeks.grupy.get(id);
        return {
          okruszek: `Grupa ${id}`,
          nazwa: id,
          podpis: `${g?.podzial ?? g?.typ ?? "grupa"} · ${g?.liczba_studentow ?? 0} os.`,
          tytul: `Plan grupy ${id}`,
          podtytul: `Zajęcia grupy i grup ze wspólnymi studentami · ${liczba}`,
        };
      }
      case "prowadzacy": {
        const p = this.indeks.prowadzacy.get(id);
        const nazwa = p?.nazwa ?? id;
        const przedmiotow = new Set(this.zajeciaWidoku.map((z) => z.przedmiot)).size;
        const rocznikow = new Set(this.zajeciaWidoku.map((z) => z.rocznikId)).size;
        return {
          okruszek: `Prowadzący ${nazwa}`,
          nazwa,
          podpis: `${p?.jednostka ?? ""} · ${Format.ilosc(przedmiotow, "przedmiot", "przedmioty", "przedmiotów")}`,
          tytul: `Plan prowadzącego ${nazwa}`,
          podtytul: `${liczba} w ${Format.ilosc(rocznikow, "roczniku", "rocznikach", "rocznikach")}`,
        };
      }
      case "sala": {
        const s = this.indeks.sale.get(id);
        const opisSali = WidokPlanu.opisSali(this.indeks, id);
        return {
          okruszek: `Sala ${id}`,
          nazwa: id,
          podpis: opisSali,
          tytul: `Plan sali ${id}${s?.nazwa ? ` · ${s.nazwa}` : ""}`,
          podtytul: `${NAZWY_TYPOW_SAL[s?.typ ?? ""] ?? ""} · ${Format.ilosc(this.zajeciaWidoku.length, "zajęcia", "zajęcia", "zajęć")} w tygodniu`,
        };
      }
    }
  }

  static opisSali(indeks: IndeksPlanu, salaId: string): string {
    const s = indeks.sale.get(salaId);
    if (!s) {
      return "";
    }

    const miejsca =
      s.pojemnosc === null
        ? "brak danych o pojemności"
        : Format.ilosc(s.pojemnosc, "miejsce", "miejsca", "miejsc");
    const typ = NAZWY_TYPOW_SAL[s.typ] ?? s.typ;
    return [s.nazwa, typ, `${miejsca}${s.pojemnosc_szacowana ? " (szacowana)" : ""}`]
      .filter(Boolean)
      .join(" · ");
  }

  // Zajętość sali: sloty zajęte w tygodniu z 60 możliwych (niezależnie od filtrów zawężających).
  zajetoscSali(): { slotow: number; procent: number } {
    const slotow = this.zajeciaWidoku.reduce((suma, z) => suma + z.dlugosc, 0);
    return { slotow, procent: Math.round((slotow / (LICZBA_DNI * LICZBA_SLOTOW)) * 100) };
  }

  statystykiProwadzacego(): StatystykiProwadzacego {
    const dniWolne = [...this.dniWolneProwadzacego()].map((d) => SKROTY_DNI[d - 1]!);
    const okienek = Array.from({ length: LICZBA_DNI }, (_, i) => this.okienka(i + 1)).reduce(
      (suma, o) => suma + o.reduce((s, b) => s + b.slotDo - b.slotOd + 1, 0),
      0,
    );
    return { slotow: this.zajeciaWidoku.reduce((suma, z) => suma + z.dlugosc, 0), dniWolne, okienek };
  }

  private karta(zajecia: ZajeciaWPlanie, tor: number, torow: number): KartaNaSiatce {
    const naruszona = this.naruszenia.dlaZajec(zajecia.id).length > 0;
    const skrot = this.indeks.skrotGrupy(zajecia.grupaId);
    const calyRocznik = this.indeks.czyCalyRocznik(zajecia.grupaId);
    return {
      zajecia,
      tor,
      torow,
      kolor: this.kolor(zajecia),
      etykietaGrupy: this.widokStudentow
        ? calyRocznik
          ? "cały rocznik"
          : zajecia.grupaId
        : `${zajecia.rocznikId} · ${calyRocznik ? "cały rocznik" : skrot}`,
      skrotGrupy: this.widokStudentow ? skrot : `${zajecia.rocznikId}${calyRocznik ? "" : ` ${skrot}`}`,
      naruszona,
      wygaszona: this.filtry.tylkoNaruszenia && !naruszona,
    };
  }

  private kolor(zajecia: ZajeciaWPlanie): string {
    if (this.filtry.kolor === "forma") {
      return KOLORY_FORM[zajecia.forma];
    }

    return this.widokStudentow
      ? (this.indeks.kolorGrupy.get(zajecia.grupaId) ?? KOLORY_FORM[zajecia.forma])
      : (this.indeks.kolorRocznika.get(zajecia.rocznikId) ?? KOLORY_FORM[zajecia.forma]);
  }

  // Dzień wolny prowadzącego: brak zajęć w tym dniu w całym planie (nie tylko po filtrach).
  private dniWolneProwadzacego(): Set<number> {
    const zajete = new Set(this.zajeciaWidoku.map((z) => z.dzien));
    return new Set(Array.from({ length: LICZBA_DNI }, (_, i) => i + 1).filter((d) => !zajete.has(d)));
  }

  private blokiProwadzacego(dzien: number, wolny: boolean): BlokNaSiatce[] {
    const wykluczone: BlokNaSiatce[] = this.indeks
      .blokadyProwadzacego(this.filtry.id)
      .filter((b) => b.dzien === dzien)
      .map((b) => ({
        rodzaj: "wykluczone",
        slotOd: b.slotOd,
        slotDo: b.slotDo,
        tytul: b.rada ? "Rada Wydziału" : "Wykluczone",
        opis: b.rada ? "godziny wykluczone" : "dezyderat prowadzącego",
      }));
    const tlo = wolny ? WidokPlanu.dzienWolny(wykluczone) : this.okienka(dzien);
    return [...tlo, ...wykluczone];
  }

  // Dzień wolny zajmuje sloty poza godzinami wykluczonymi (makieta: „poza wykluczonymi slotami”).
  // Napis „Dzień wolny” dostaje najdłuższy wolny odcinek, żeby nie nachodził na blok wykluczeń.
  private static dzienWolny(wykluczone: BlokNaSiatce[]): BlokNaSiatce[] {
    const zajete = new Set(
      wykluczone.flatMap((b) => Array.from({ length: b.slotDo - b.slotOd + 1 }, (_, i) => b.slotOd + i)),
    );
    const odcinki: { slotOd: number; slotDo: number }[] = [];
    for (let slot = 1; slot <= LICZBA_SLOTOW; slot++) {
      if (zajete.has(slot)) {
        continue;
      }

      const ostatni = odcinki.at(-1);
      if (ostatni && ostatni.slotDo === slot - 1) {
        ostatni.slotDo = slot;
      } else {
        odcinki.push({ slotOd: slot, slotDo: slot });
      }
    }
    const najdluzszy = odcinki.reduce<(typeof odcinki)[number] | undefined>(
      (max, o) => (!max || o.slotDo - o.slotOd > max.slotDo - max.slotOd ? o : max),
      undefined,
    );
    return odcinki.map((o) => ({
      rodzaj: "wolny",
      ...o,
      tytul: o === najdluzszy ? "Dzień wolny" : "",
      opis: o === najdluzszy && wykluczone.length ? "poza wykluczonymi slotami" : "",
    }));
  }

  // Okienka prowadzącego: wolne sloty między jego zajęciami w danym dniu.
  private okienka(dzien: number): BlokNaSiatce[] {
    const zajete = new Set(
      this.zajeciaWidoku
        .filter((z) => z.dzien === dzien)
        .flatMap((z) => Array.from({ length: z.dlugosc }, (_, i) => z.slotStart + i)),
    );
    if (zajete.size === 0) {
      return [];
    }

    const pierwszy = Math.min(...zajete);
    const ostatni = Math.max(...zajete);
    const bloki: BlokNaSiatce[] = [];
    for (let slot = pierwszy; slot <= ostatni; slot++) {
      if (zajete.has(slot)) {
        continue;
      }

      const poprzedni = bloki.at(-1);
      if (poprzedni && poprzedni.slotDo === slot - 1) {
        poprzedni.slotDo = slot;
        poprzedni.opis = Format.ilosc(
          slot - poprzedni.slotOd + 1,
          "slot przerwy",
          "sloty przerwy",
          "slotów przerwy",
        );
      } else {
        bloki.push({
          rodzaj: "okienko",
          slotOd: slot,
          slotDo: slot,
          tytul: "Okienko",
          opis: "1 slot przerwy",
        });
      }
    }
    return bloki;
  }
}
