import type { Schematy } from "../../../api/klient.ts";
import { Kolory } from "./kolory.ts";
import type { ZajeciaWPlanie } from "./typy.ts";

type Dane = Schematy["DaneWejsciowe"];

export type BlokadaProwadzacego = { dzien: number; slotOd: number; slotDo: number; rada: boolean };

// Dane zestawu i plan w postaci wygodnej dla kalendarza: mapy po id, grupy roczników, kolizje grup.
// Budowany raz na uruchomienie (useMemo w stronie), bo dane i plan się nie zmieniają.
export class IndeksPlanu {
  readonly zajecia: ZajeciaWPlanie[];
  readonly sale: Map<string, Schematy["Sala"]>;
  readonly prowadzacy: Map<string, Schematy["Prowadzacy"]>;
  readonly roczniki: Schematy["Rocznik"][];
  readonly grupy: Map<string, Schematy["Grupa"]>;
  readonly grupyRocznika: Map<string, Schematy["Grupa"][]>;
  readonly przedmioty: Schematy["Przedmiot"][];
  readonly sloty: Map<number, Schematy["Slot"]>;
  readonly kolorGrupy: Map<string, string>;
  readonly kolorRocznika: Map<string, string>;
  private readonly kolizjeGrup: Map<string, Set<string>>;
  private readonly blokady: Map<string, BlokadaProwadzacego[]>;

  constructor(dane: Dane, plan: Schematy["Plan"]) {
    this.sale = new Map(dane.sale.map((s) => [s.id, s]));
    this.prowadzacy = new Map(dane.prowadzacy.map((p) => [p.id, p]));
    this.roczniki = [...dane.roczniki].sort((a, b) => a.id.localeCompare(b.id));
    this.grupy = new Map(dane.grupy.map((g) => [g.id, g]));
    this.przedmioty = [...dane.przedmioty].sort((a, b) => a.nazwa.localeCompare(b.nazwa, "pl"));
    this.sloty = new Map(dane.sloty.map((s) => [s.slot, s]));

    this.grupyRocznika = new Map(this.roczniki.map((r) => [r.id, [] as Schematy["Grupa"][]]));
    for (const grupa of dane.grupy) {
      this.grupyRocznika.get(grupa.rocznik)?.push(grupa);
    }
    this.kolorGrupy = new Map();
    for (const grupy of this.grupyRocznika.values()) {
      grupy.forEach((g, i) => this.kolorGrupy.set(g.id, Kolory.zPalety(i)));
    }
    this.kolorRocznika = new Map(this.roczniki.map((r, i) => [r.id, Kolory.zPalety(i)]));

    this.kolizjeGrup = new Map();
    for (const { grupa_a, grupa_b, wspolni_studenci } of dane.kolizje_grup) {
      if (wspolni_studenci > 0) {
        this.dodajKolizje(grupa_a, grupa_b);
        this.dodajKolizje(grupa_b, grupa_a);
      }
    }

    this.blokady = IndeksPlanu.scalBlokady(dane.dezyderaty_terminy);

    const zajeciaZDanych = new Map(dane.zajecia.map((z) => [z.id, z]));
    this.zajecia = plan.przypisania.flatMap((p) => {
      const z = zajeciaZDanych.get(p.zajecia_id);
      if (!z) {
        return [];
      }

      return [
        {
          id: z.id,
          przedmiot: z.przedmiot,
          forma: z.forma,
          salaId: p.sala_id,
          prowadzacy: z.prowadzacy,
          rocznikId: z.rocznik,
          grupaId: z.grupy[0] ?? `${z.rocznik}-ALL`,
          dzien: p.dzien,
          slotStart: p.slot_start,
          dlugosc: z.sloty_na_spotkanie,
          tygodnie: p.tygodnie,
        },
      ];
    });
  }

  // Grupy, których zajęcia dotyczą studentów danej grupy: ona sama i grupy ze wspólnymi studentami.
  grupyPowiazane(grupaId: string): Set<string> {
    return new Set([grupaId, ...(this.kolizjeGrup.get(grupaId) ?? [])]);
  }

  czyCalyRocznik(grupaId: string): boolean {
    return this.grupy.get(grupaId)?.typ === "rocznik";
  }

  // „G2” dla INF1-R1-G2; „rocz.” dla grupy całego rocznika.
  skrotGrupy(grupaId: string): string {
    if (this.czyCalyRocznik(grupaId)) {
      return "rocz.";
    }

    const rocznik = this.grupy.get(grupaId)?.rocznik;
    return rocznik && grupaId.startsWith(`${rocznik}-`) ? grupaId.slice(rocznik.length + 1) : grupaId;
  }

  nazwaProwadzacego(prowadzacyId: string): string {
    return this.prowadzacy.get(prowadzacyId)?.nazwa ?? prowadzacyId;
  }

  godziny(slotOd: number, slotDo: number): string {
    return `${this.sloty.get(slotOd)?.od ?? ""}–${this.sloty.get(slotDo)?.do ?? ""}`;
  }

  blokadyProwadzacego(prowadzacyId: string): BlokadaProwadzacego[] {
    return this.blokady.get(prowadzacyId) ?? [];
  }

  private dodajKolizje(grupa: string, inna: string) {
    const zbior = this.kolizjeGrup.get(grupa) ?? new Set<string>();
    zbior.add(inna);
    this.kolizjeGrup.set(grupa, zbior);
  }

  // Dezyderaty to pojedyncze sloty; na siatce pokazujemy ciągłe bloki tego samego rodzaju.
  private static scalBlokady(terminy: Schematy["DezyderatTermin"][]): Map<string, BlokadaProwadzacego[]> {
    const posortowane = [...terminy].sort(
      (a, b) => a.prowadzacy.localeCompare(b.prowadzacy) || a.dzien - b.dzien || a.slot - b.slot,
    );
    const wynik = new Map<string, BlokadaProwadzacego[]>();
    for (const termin of posortowane) {
      const lista = wynik.get(termin.prowadzacy) ?? [];
      const ostatnia = lista.at(-1);
      const rada = termin.typ === "rada_wydzialu";
      if (
        ostatnia &&
        ostatnia.dzien === termin.dzien &&
        ostatnia.slotDo + 1 === termin.slot &&
        ostatnia.rada === rada
      ) {
        ostatnia.slotDo = termin.slot;
      } else {
        lista.push({ dzien: termin.dzien, slotOd: termin.slot, slotDo: termin.slot, rada });
      }
      wynik.set(termin.prowadzacy, lista);
    }
    return wynik;
  }
}
