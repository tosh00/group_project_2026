import type { Schematy } from "../../../api/klient.ts";
import type { IndeksPlanu } from "./indeks.ts";
import type { StatusPlanu, ZajeciaWPlanie } from "./typy.ts";

type Raport = Schematy["Raport"];

// Nazwy reguł twardych do panelu (kolejność i identyfikatory z docs/contracts/api.md).
export const REGULY: { nazwa: string; etykieta: string }[] = [
  { nazwa: "przypisanie", etykieta: "przypisanie" },
  { nazwa: "tygodnie", etykieta: "tygodnie" },
  { nazwa: "kolizja_prowadzacego", etykieta: "kolizja prowadzącego" },
  { nazwa: "kolizja_sali", etykieta: "kolizja sali" },
  { nazwa: "kolizja_grup", etykieta: "kolizja grup" },
  { nazwa: "typ_sali", etykieta: "typ sali" },
  { nazwa: "pojemnosc", etykieta: "pojemność" },
  { nazwa: "dezyderaty", etykieta: "dezyderaty" },
  { nazwa: "blokady_stale", etykieta: "blokady stałe" },
  { nazwa: "powiazania", etykieta: "powiązania" },
];

const etykietaReguly = new Map(REGULY.map((r) => [r.nazwa, r.etykieta]));

export type Naruszenie = Schematy["Naruszenie"] & { indeks: number };

// Naruszenia z raportu walidatora, z odnośnikami do zajęć (podświetlanie kart, liczniki w filtrach).
export class NaruszeniaPlanu {
  readonly status: StatusPlanu;
  readonly lista: Naruszenie[];
  private readonly poZajeciach: Map<string, number[]>;

  constructor(raport: Raport | undefined) {
    this.status = NaruszeniaPlanu.statusZRaportu(raport);
    this.lista = (raport?.naruszenia ?? []).map((n, indeks) => ({ ...n, indeks }));
    this.poZajeciach = new Map();
    for (const n of this.lista) {
      for (const zajeciaId of n.zajecia_id) {
        this.poZajeciach.set(zajeciaId, [...(this.poZajeciach.get(zajeciaId) ?? []), n.indeks]);
      }
    }
  }

  static etykieta(nazwa: string): string {
    const tekst = etykietaReguly.get(nazwa) ?? nazwa;
    return tekst.charAt(0).toUpperCase() + tekst.slice(1);
  }

  dlaZajec(zajeciaId: string): Naruszenie[] {
    return (this.poZajeciach.get(zajeciaId) ?? []).map((i) => this.lista[i]!);
  }

  // Zajęcia, które razem łamią te same reguły co dane zajęcia (np. obie strony kolizji).
  zajeciaPowiazane(zajeciaId: string): Set<string> {
    return new Set(this.dlaZajec(zajeciaId).flatMap((n) => n.zajecia_id));
  }

  // Liczba różnych naruszeń, które dotyczą choć jednych z podanych zajęć (liczniki „INF1 (2)”).
  liczbaDla(zajecia: readonly ZajeciaWPlanie[]): number {
    return new Set(zajecia.flatMap((z) => this.poZajeciach.get(z.id) ?? [])).size;
  }

  wgReguly(): { nazwa: string; etykieta: string; naruszenia: Naruszenie[] }[] {
    return REGULY.map((r) => ({
      nazwa: r.nazwa,
      etykieta: NaruszeniaPlanu.etykieta(r.nazwa),
      naruszenia: this.lista.filter((n) => n.nazwa === r.nazwa),
    })).filter((g) => g.naruszenia.length > 0);
  }

  regulyBezNaruszen(): string[] {
    const zNaruszeniami = new Set(this.lista.map((n) => n.nazwa));
    return REGULY.filter((r) => !zNaruszeniami.has(r.nazwa)).map((r) => r.etykieta);
  }

  // Zajęcia naruszenia z planu: gdy walidator nie podał dnia, slotów czy sali, bierzemy je z przypisania.
  zajeciaNaruszenia(naruszenie: Naruszenie, indeks: IndeksPlanu): ZajeciaWPlanie[] {
    const ids = new Set(naruszenie.zajecia_id);
    return indeks.zajecia.filter((z) => ids.has(z.id));
  }

  private static statusZRaportu(raport: Raport | undefined): StatusPlanu {
    if (!raport || raport.walidator === "atrapa") {
      return "niesprawdzony";
    }

    return raport.status === "z_naruszeniami" ? "naruszenia" : "poprawny";
  }
}
