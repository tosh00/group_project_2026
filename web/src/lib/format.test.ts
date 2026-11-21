import { describe, expect, it } from "vitest";
import { Format } from "./format.ts";

describe("Format", () => {
  it.each([
    [1, "zajęcie"],
    [2, "zajęcia"],
    [4, "zajęcia"],
    [5, "zajęć"],
    [12, "zajęć"],
    [22, "zajęcia"],
    [174, "zajęcia"],
    [168, "zajęć"],
  ])("odmienia %i → %s", (liczba, oczekiwane) => {
    expect(Format.odmiana(liczba, "zajęcie", "zajęcia", "zajęć")).toBe(oczekiwane);
  });

  it("grupuje tysiące i zaokrągla koszt", () => {
    expect(Format.liczba(1284.4).replace(/\s/g, " ")).toBe("1 284");
  });

  it("pisze „dziś” dla dzisiejszej daty i pełną datę dla starszej", () => {
    const teraz = new Date("2026-11-21T12:00:00");

    expect(Format.dataWzgledna("2026-11-21T11:05:00", teraz)).toBe("dziś, 11:05");
    expect(Format.dataWzgledna("2026-11-14T09:00:00", teraz)).toBe("14.11.2026");
  });
});
