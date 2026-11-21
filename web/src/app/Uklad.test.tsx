import { describe, expect, it } from "vitest";
import { screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "../mocks/server.ts";
import { renderTrase } from "../test/render.tsx";

describe("Uklad", () => {
  it("ma w nagłówku logo wydziału prowadzące do pulpitu", async () => {
    renderTrase("/");

    const logo = await screen.findByRole("link", { name: "Pulpit" });
    expect(logo).toHaveAttribute("href", "/");
    expect(screen.getByAltText(/Wydział Fizyki i Informatyki Stosowanej/)).toBeInTheDocument();
  });

  it("nie pokazuje znacznika, gdy serwer działa", async () => {
    renderTrase("/");

    await screen.findByRole("link", { name: "Pulpit" });
    expect(screen.queryByText("Serwer niedostępny")).not.toBeInTheDocument();
  });

  it("pokazuje „Serwer niedostępny”, gdy baza nie odpowiada", async () => {
    server.use(
      http.get("/api/zdrowie", () =>
        HttpResponse.json({ status: "blad", baza: "blad", walidator: "atrapa" }, { status: 503 }),
      ),
    );

    renderTrase("/");

    expect(await screen.findByText("Serwer niedostępny")).toBeInTheDocument();
  });
});
