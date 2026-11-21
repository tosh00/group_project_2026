import { describe, expect, it } from "vitest";
import { IndeksPlanu } from "./indeks.ts";
import { NaruszeniaPlanu } from "./naruszenia.ts";
import { Tygodnie } from "./tygodnie.ts";
import { UkladTorow } from "./uklad.ts";
import { WidokPlanu } from "./widok.ts";
import { Wyszukiwarka } from "./wyszukiwanie.ts";
import { ParametryFiltrow } from "../useFiltryKalendarza.ts";
import type { Filtry } from "./typy.ts";
import { dane, plan, raportZNaruszeniami } from "./dane.testowe.ts";

const indeks = new IndeksPlanu(dane, plan);
const bezNaruszen = new NaruszeniaPlanu(undefined);
const filtry = (zmiana: Partial<Filtry> = {}): Filtry => ({
  wymiar: "rocznik",
  id: "INF1-R1",
  tydzien: 0,
  ukryteFormy: new Set(),
  tylkoNaruszenia: false,
  kolor: "grupa",
  ...zmiana,
});

describe("Tygodnie.opis", () => {
  it.each([
    [[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14], ""],
    [[2, 4, 6, 8, 10, 12, 14], "co 2 tyg. (parz.)"],
    [[1, 3, 5, 7, 9, 11, 13], "co 2 tyg. (nieparz.)"],
    [[6, 7, 8, 9, 10, 11, 12, 13, 14], "tyg. 6–14"],
    [[1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 13, 14], "bez tyg. 6"],
    [[1, 2, 3, 5, 7, 8, 10, 12, 14], "9 wybranych tyg."],
  ])("%j → „%s”", (tygodnie, opis) => {
    expect(Tygodnie.opis(tygodnie).pelny).toBe(opis);
  });
});

describe("UkladTorow", () => {
  it("rozkłada nachodzące zajęcia na tory, a rozłączne zostawia w jednym", () => {
    const wynik = UkladTorow.rozloz(
      [
        { id: "a", start: 5, dlugosc: 2 },
        { id: "b", start: 5, dlugosc: 3 },
        { id: "c", start: 6, dlugosc: 2 },
        { id: "d", start: 10, dlugosc: 1 },
      ],
      (x) => x,
    );
    const po = Object.fromEntries(wynik.map((w) => [w.element.id, [w.tor, w.torow]]));

    expect(po).toEqual({ b: [0, 3], a: [1, 3], c: [2, 3], d: [0, 1] });
  });

  it("zwalnia tor, gdy poprzednie zajęcia się skończyły", () => {
    const wynik = UkladTorow.rozloz(
      [
        { id: "a", start: 1, dlugosc: 4 },
        { id: "b", start: 1, dlugosc: 2 },
        { id: "c", start: 3, dlugosc: 2 },
      ],
      (x) => x,
    );

    expect(wynik.find((w) => w.element.id === "c")).toMatchObject({ tor: 1, torow: 2 });
  });
});

describe("WidokPlanu", () => {
  it("widok rocznika ma zajęcia rocznika, a równoległe grupy dostają osobne tory", () => {
    const widok = new WidokPlanu(indeks, bezNaruszen, filtry());
    const poniedzialek = widok.dni()[0]!;

    expect(widok.zajeciaWidoku.map((z) => z.id).sort()).toEqual(["Z001", "Z002", "Z003"]);
    expect(poniedzialek.karty.map((k) => [k.zajecia.id, k.torow])).toEqual([
      ["Z001", 1],
      ["Z002", 2],
      ["Z003", 2],
    ]);
    expect(widok.opis().tytul).toBe("Plan rocznika INF1-R1");
  });

  it("widok grupy pokazuje jej zajęcia i zajęcia całego rocznika, ale nie innej grupy", () => {
    const widok = new WidokPlanu(indeks, bezNaruszen, filtry({ wymiar: "grupa", id: "INF1-R1-G1" }));

    expect(widok.zajeciaWidoku.map((z) => z.id).sort()).toEqual(["Z001", "Z002"]);
  });

  it("filtr tygodnia i formy zawęża siatkę, ale nie opis widoku", () => {
    const widok = new WidokPlanu(indeks, bezNaruszen, filtry({ tydzien: 3, ukryteFormy: new Set(["W"]) }));

    expect(widok.widoczne).toHaveLength(0);
    expect(widok.zajeciaWidoku).toHaveLength(3);
  });

  it("widok prowadzącego: wykluczenia, dzień wolny poza nimi i statystyki", () => {
    const widok = new WidokPlanu(indeks, bezNaruszen, filtry({ wymiar: "prowadzacy", id: "P07" }));
    const piatek = widok.dni()[4]!;

    expect(piatek.wolny).toBe(true);
    expect(piatek.bloki).toEqual([
      expect.objectContaining({ rodzaj: "wolny", slotOd: 3, slotDo: 12, tytul: "Dzień wolny" }),
      expect.objectContaining({ rodzaj: "wykluczone", slotOd: 1, slotDo: 2, tytul: "Wykluczone" }),
    ]);
    expect(widok.statystykiProwadzacego()).toEqual({ slotow: 6, dniWolne: ["śr", "czw", "pt"], okienek: 0 });
  });

  it("w widoku sali liczy zajętość z 60 slotów", () => {
    const widok = new WidokPlanu(indeks, bezNaruszen, filtry({ wymiar: "sala", id: "A281" }));

    expect(widok.zajetoscSali()).toEqual({ slotow: 6, procent: 10 });
  });

  it("„tylko naruszenia” wygasza karty bez naruszeń", () => {
    const naruszenia = new NaruszeniaPlanu(raportZNaruszeniami);
    const karty = new WidokPlanu(indeks, naruszenia, filtry({ tylkoNaruszenia: true })).dni()[0]!.karty;

    expect(karty.map((k) => [k.zajecia.id, k.naruszona, k.wygaszona])).toEqual([
      ["Z001", false, true],
      ["Z002", true, false],
      ["Z003", false, true],
    ]);
  });
});

describe("NaruszeniaPlanu", () => {
  it("atrapa walidatora to status niesprawdzony, nawet gdy raport mówi „poprawny”", () => {
    expect(
      new NaruszeniaPlanu({ ...raportZNaruszeniami, status: "poprawny", walidator: "atrapa" }).status,
    ).toBe("niesprawdzony");
    expect(new NaruszeniaPlanu(raportZNaruszeniami).status).toBe("naruszenia");
  });

  it("liczy naruszenia dla zajęć widoku i grupuje je według reguły", () => {
    const naruszenia = new NaruszeniaPlanu(raportZNaruszeniami);

    expect(naruszenia.liczbaDla(WidokPlanu.zajeciaDla(indeks, "grupa", "INF1-R1-G1"))).toBe(1);
    expect(naruszenia.liczbaDla(WidokPlanu.zajeciaDla(indeks, "grupa", "INF1-R1-G2"))).toBe(0);
    expect(naruszenia.wgReguly().map((g) => [g.etykieta, g.naruszenia.length])).toEqual([["Pojemność", 1]]);
    expect(naruszenia.regulyBezNaruszen()).toHaveLength(9);
  });
});

describe("Wyszukiwarka", () => {
  const wyszukiwarka = new Wyszukiwarka(indeks, new NaruszeniaPlanu(raportZNaruszeniami));

  it("bez frazy pokazuje roczniki, prowadzących i sale, a przedmioty dopiero z frazą", () => {
    expect(wyszukiwarka.szukaj("").map((s) => s.tytul)).toEqual(["Roczniki i grupy", "Prowadzący", "Sale"]);
    expect(wyszukiwarka.szukaj("analiza").map((s) => s.tytul)).toContain("Przedmioty");
  });

  it("znajduje po kodzie i pokazuje liczbę naruszeń", () => {
    const sale = wyszukiwarka.szukaj("b15").find((s) => s.tytul === "Sale")!;

    expect(sale.wyniki).toEqual([
      expect.objectContaining({ etykieta: "B151", naruszen: 1, cel: { wymiar: "sala", id: "B151" } }),
    ]);
  });

  it("dzieli etykietę na część przed trafieniem, trafienie i resztę", () => {
    expect(Wyszukiwarka.podswietl("A281", "28")).toEqual(["A", "28", "1"]);
    expect(Wyszukiwarka.podswietl("A281", "x")).toEqual(["A281", "", ""]);
  });
});

describe("ParametryFiltrow", () => {
  it("zapisuje filtry w URL i odczytuje je bez zmian", () => {
    const zapisane = ParametryFiltrow.zapisz(
      filtry({
        wymiar: "prowadzacy",
        id: "P07",
        tydzien: 6,
        ukryteFormy: new Set(["LEK", "W"]),
        tylkoNaruszenia: true,
        kolor: "forma",
      }),
    );

    expect(zapisane.toString()).toBe(
      "widok=prowadzacy&id=P07&tydzien=6&bez=W%2CLEK&naruszenia=1&kolor=forma",
    );
    expect(ParametryFiltrow.odczytaj(zapisane, "INF1-R1")).toEqual(
      filtry({
        wymiar: "prowadzacy",
        id: "P07",
        tydzien: 6,
        ukryteFormy: new Set(["W", "LEK"]),
        tylkoNaruszenia: true,
        kolor: "forma",
      }),
    );
  });

  it("odrzuca nieznane wartości i bierze domyślny rocznik", () => {
    expect(ParametryFiltrow.odczytaj(new URLSearchParams("widok=xyz&tydzien=99&bez=Q"), "INF1-R1")).toEqual(
      filtry(),
    );
  });
});
