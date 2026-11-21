import { describe, expect, it } from "vitest";
import { screen } from "@testing-library/react";
import { renderTrase } from "../test/render.tsx";

describe("trasy", () => {
  it.each([
    ["/", "Zestawy danych"],
    ["/zestawy/abc", "Zestaw danych"],
    ["/uruchomienia/abc", "Kalendarz planu"],
    ["/uruchomienia/abc/raport", "Raport jakości"],
    ["/nie/ma/takiej", "Nie ma takiej strony"],
  ])("%s pokazuje stronę „%s”", async (sciezka, tytul) => {
    renderTrase(sciezka);

    expect(await screen.findByRole("heading", { name: tytul })).toBeInTheDocument();
  });
});
