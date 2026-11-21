import { describe, expect, it } from "vitest";
import { uruchomienie } from "../../mocks/dane.ts";
import { BladApi } from "../../api/klient.ts";
import { BledyPliku } from "./bledyPliku.ts";
import { StatusyPlanow } from "./statusPlanu.ts";

describe("StatusyPlanow", () => {
  it("nie pokazuje zielonego statusu, gdy walidator to atrapa", () => {
    expect(StatusyPlanow.zUruchomienia(uruchomienie, "atrapa")).toMatchObject({
      rodzaj: "neutralny",
      tekst: "Status niesprawdzony",
    });
    expect(StatusyPlanow.zUruchomienia(uruchomienie, "internal/validator")).toMatchObject({ rodzaj: "ok" });
  });

  it("pokazuje liczbę naruszeń na czerwono", () => {
    const zNaruszeniami = { ...uruchomienie, status: "z_naruszeniami" as const, liczba_naruszen: 2 };

    expect(StatusyPlanow.zUruchomienia(zNaruszeniami, "internal/validator")).toMatchObject({
      rodzaj: "blad",
      tekst: "Plan z naruszeniami · 2",
    });
  });

  it("bierze najnowsze uruchomienie każdego zestawu", () => {
    const nowsze = { ...uruchomienie, id: "nowsze" };
    const ostatnie = StatusyPlanow.ostatnieWgZestawu([nowsze, uruchomienie]);

    expect(ostatnie.get(uruchomienie.zestaw_id)?.id).toBe("nowsze");
  });
});

describe("BledyPliku", () => {
  it("rozbija błąd modelu na miejsce w danych i opis", () => {
    expect(BledyPliku.zTekstu('zajecia Z004: dozwolone_sale: nieznany identyfikator "X999"')).toEqual({
      miejsce: "zajecia[Z004].dozwolone_sale",
      opis: 'Nieznany identyfikator "X999"',
    });
  });

  it("używa komunikatu, gdy backend nie podał szczegółów", () => {
    const blad = new BladApi({
      status: 400,
      kod: "zle_zadanie",
      komunikat: "Plik nie jest poprawnym JSON-em.",
    });

    expect(BledyPliku.zBleduApi(blad)).toEqual([
      { miejsce: "zle_zadanie", opis: "Plik nie jest poprawnym JSON-em." },
    ]);
  });
});
