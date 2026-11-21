import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "../mocks/server.ts";
import { zestaw } from "../mocks/dane.ts";
import { api, BladApi } from "./klient.ts";

describe("BladApi.dane", () => {
  it("zwraca dane z odpowiedzi 2xx", async () => {
    const zestawy = await BladApi.dane(api.GET("/zestawy"));

    expect(zestawy).toEqual([zestaw]);
  });

  it("rzuca BladApi z kodem i wszystkimi szczegółami dla 422", async () => {
    server.use(
      http.post("/api/zestawy", () =>
        HttpResponse.json(
          {
            kod: "niespojne_dane",
            komunikat: "Dane są niespójne.",
            szczegoly: ["zajecia[3]: nieznana sala X1", "grupy[7]: brak rocznika"],
          },
          { status: 422 },
        ),
      ),
    );

    const blad = await BladApi.dane(api.POST("/zestawy", { body: {} })).catch((e: unknown) => e);

    expect(blad).toBeInstanceOf(BladApi);
    expect(blad).toMatchObject({
      status: 422,
      kod: "niespojne_dane",
      message: "Dane są niespójne.",
      szczegoly: ["zajecia[3]: nieznana sala X1", "grupy[7]: brak rocznika"],
    });
  });

  it("zamienia odpowiedź bez treści Blad na blad_serwera", async () => {
    server.use(http.get("/api/zestawy", () => new HttpResponse("Bad Gateway", { status: 502 })));

    const blad = await BladApi.dane(api.GET("/zestawy")).catch((e: unknown) => e);

    expect(blad).toMatchObject({ status: 502, kod: "blad_serwera", szczegoly: [] });
  });
});
