import { describe, expect, it } from "vitest";
import { screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "../mocks/server.ts";
import { renderTrase } from "../test/render.tsx";

describe("Uklad", () => {
  it("pokazuje, że walidator nie jest podłączony, gdy backend używa atrapy", async () => {
    renderTrase("/");

    expect(await screen.findByText("Walidator niepodłączony")).toBeInTheDocument();
  });

  it("pokazuje „Serwer działa” z prawdziwym walidatorem", async () => {
    server.use(
      http.get("/api/zdrowie", () =>
        HttpResponse.json({ status: "ok", baza: "ok", walidator: "internal/validator" }),
      ),
    );

    renderTrase("/");

    expect(await screen.findByText("Serwer działa")).toBeInTheDocument();
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
