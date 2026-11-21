import { Format } from "../../../lib/format.ts";
import type { IndeksPlanu } from "./indeks.ts";
import type { NaruszeniaPlanu } from "./naruszenia.ts";
import type { Wymiar } from "./typy.ts";
import { NAZWY_TYPOW_SAL, WidokPlanu } from "./widok.ts";

export type Cel = { wymiar: Wymiar; id: string };

export type Wynik = {
  klucz: string;
  etykieta: string;
  opis: string;
  cel: Cel;
  naruszen: number;
  kolor?: string;
  wciecie?: boolean;
};

export type SekcjaWynikow = { tytul: string; notka: string; wyniki: Wynik[] };

// Limity z makiety: bez frazy krótka lista na sekcję, z frazą do 8 wyników.
const LIMIT_BEZ_FRAZY = 8;
const LIMIT_PROWADZACYCH_SAL = 4;
const LIMIT_Z_FRAZA = 8;

// Źródło wyszukiwarki „Szukaj: P07, A28, Analiza…”: roczniki z grupami, prowadzący, sale, przedmioty.
export class Wyszukiwarka {
  private readonly sekcje: { tytul: string; notka: string; tylkoZFraza: boolean; wyniki: Wynik[] }[];

  constructor(indeks: IndeksPlanu, naruszenia: NaruszeniaPlanu) {
    const liczba = (wymiar: Wymiar, id: string) =>
      naruszenia.liczbaDla(WidokPlanu.zajeciaDla(indeks, wymiar, id));

    const rocznikiIGrupy = indeks.roczniki.flatMap((r): Wynik[] => {
      const grupy = (indeks.grupyRocznika.get(r.id) ?? []).filter((g) => g.typ !== "rocznik");
      return [
        {
          klucz: `rocznik:${r.id}`,
          etykieta: r.id,
          opis: `${r.nazwa} · ${r.liczba_studentow} os. · ${Format.ilosc(grupy.length, "grupa", "grupy", "grup")}`,
          cel: { wymiar: "rocznik", id: r.id },
          naruszen: liczba("rocznik", r.id),
          kolor: indeks.kolorRocznika.get(r.id),
        },
        ...grupy.map((g): Wynik => ({
          klucz: `grupa:${g.id}`,
          etykieta: g.id,
          opis: `${g.podzial ?? g.typ} · ${g.liczba_studentow} os.`,
          cel: { wymiar: "grupa", id: g.id },
          naruszen: liczba("grupa", g.id),
          kolor: indeks.kolorGrupy.get(g.id),
          wciecie: true,
        })),
      ];
    });

    const prowadzacy = [...indeks.prowadzacy.values()]
      .sort((a, b) => a.id.localeCompare(b.id, "pl", { numeric: true }))
      .map((p): Wynik => {
        const zajecia = WidokPlanu.zajeciaDla(indeks, "prowadzacy", p.id);
        const przedmioty = [...new Set(zajecia.map((z) => z.przedmiot))];
        return {
          klucz: `prowadzacy:${p.id}`,
          etykieta: p.nazwa,
          opis: `${przedmioty.slice(0, 2).join(", ") || "bez zajęć"} · ${Format.ilosc(zajecia.length, "zajęcia", "zajęcia", "zajęć")}`,
          cel: { wymiar: "prowadzacy", id: p.id },
          naruszen: naruszenia.liczbaDla(zajecia),
        };
      });

    const sale = [...indeks.sale.values()]
      .filter((s) => s.dostepna_do_planowania)
      .sort((a, b) => a.id.localeCompare(b.id))
      .map((s): Wynik => ({
        klucz: `sala:${s.id}`,
        etykieta: s.id,
        opis: WidokPlanu.opisSali(indeks, s.id) || (NAZWY_TYPOW_SAL[s.typ] ?? s.typ),
        cel: { wymiar: "sala", id: s.id },
        naruszen: liczba("sala", s.id),
      }));

    const przedmioty = indeks.przedmioty.map((p): Wynik => {
      const zajecia = indeks.zajecia.filter((z) => z.rocznikId === p.rocznik && z.przedmiot === p.nazwa);
      const formy = [...new Set(zajecia.map((z) => z.forma))].join(", ");
      const prowadzacyNazwy = [...new Set(zajecia.flatMap((z) => z.prowadzacy))]
        .map((id) => indeks.nazwaProwadzacego(id))
        .slice(0, 3)
        .join(", ");
      return {
        klucz: `przedmiot:${p.id}`,
        etykieta: p.nazwa,
        opis: [p.rocznik, formy, prowadzacyNazwy].filter(Boolean).join(" · "),
        cel: { wymiar: "rocznik", id: p.rocznik },
        naruszen: naruszenia.liczbaDla(zajecia),
      };
    });

    const grup = indeks.grupy.size - indeks.roczniki.length;
    this.sekcje = [
      {
        tytul: "Roczniki i grupy",
        notka: `${Format.ilosc(indeks.roczniki.length, "rocznik", "roczniki", "roczników")} · ${Format.ilosc(grup, "grupa", "grupy", "grup")}`,
        tylkoZFraza: false,
        wyniki: rocznikiIGrupy,
      },
      {
        tytul: "Prowadzący",
        notka: `${prowadzacy.length} · według kodu`,
        tylkoZFraza: false,
        wyniki: prowadzacy,
      },
      { tytul: "Sale", notka: `${sale.length} do planowania`, tylkoZFraza: false, wyniki: sale },
      { tytul: "Przedmioty", notka: "otwiera plan rocznika", tylkoZFraza: true, wyniki: przedmioty },
    ];
  }

  szukaj(fraza: string): SekcjaWynikow[] {
    const q = fraza.trim().toLowerCase();
    return this.sekcje
      .filter((s) => !s.tylkoZFraza || q)
      .map((s) => {
        const limit = q
          ? LIMIT_Z_FRAZA
          : s.tytul === "Roczniki i grupy"
            ? LIMIT_BEZ_FRAZY
            : LIMIT_PROWADZACYCH_SAL;
        const wyniki = s.wyniki
          .filter((w) => !q || `${w.etykieta} ${w.opis} ${w.cel.id}`.toLowerCase().includes(q))
          .slice(0, limit);
        return { tytul: s.tytul, notka: q ? `${wyniki.length} wyn.` : s.notka, wyniki };
      })
      .filter((s) => s.wyniki.length > 0);
  }

  // Podział etykiety na część przed trafieniem, trafienie i resztę (podświetlenie <mark>).
  static podswietl(etykieta: string, fraza: string): [string, string, string] {
    const q = fraza.trim().toLowerCase();
    const i = q ? etykieta.toLowerCase().indexOf(q) : -1;
    if (i < 0) {
      return [etykieta, "", ""];
    }

    return [etykieta.slice(0, i), etykieta.slice(i, i + q.length), etykieta.slice(i + q.length)];
  }
}
