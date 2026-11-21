import { http, HttpResponse } from "msw";
import type { Schematy } from "../api/klient.ts";
import * as dane from "./dane.ts";

type Blad = Schematy["Blad"];

const nieZnaleziono = () =>
  HttpResponse.json<Blad>({ kod: "nie_znaleziono", komunikat: "Nie ma takiego zasobu." }, { status: 404 });

// Handlery REST działają w przeglądarce (`pnpm dev:mock`) i w testach (`src/mocks/server.ts`).
// Testy nadpisują je przez `server.use(...)`, np. żeby sprawdzić błąd 422.
export const handlers = [
  http.get("/api/zdrowie", () => HttpResponse.json(dane.zdrowie)),
  http.get("/api/wagi/domyslne", () => HttpResponse.json(dane.wagiDomyslne)),
  http.get("/api/zestawy", () => HttpResponse.json([dane.zestaw])),
  http.get("/api/zestawy/:zestawId", ({ params }) =>
    params.zestawId === dane.zestaw.id ? HttpResponse.json(dane.zestaw) : nieZnaleziono(),
  ),
  http.delete("/api/zestawy/:zestawId", ({ params }) =>
    params.zestawId === dane.zestaw.id ? new HttpResponse(null, { status: 204 }) : nieZnaleziono(),
  ),
  http.get("/api/zestawy/:zestawId/dane", ({ params }) =>
    params.zestawId === dane.zestaw.id ? HttpResponse.json(dane.daneZestawu) : nieZnaleziono(),
  ),
  http.get("/api/uruchomienia", () => HttpResponse.json([dane.uruchomienie])),
  http.get("/api/uruchomienia/:uruchomienieId/plan", ({ params }) =>
    params.uruchomienieId === dane.uruchomienie.id ? HttpResponse.json(dane.plan) : nieZnaleziono(),
  ),
  http.get("/api/uruchomienia/:uruchomienieId/raport", ({ params }) =>
    params.uruchomienieId === dane.uruchomienie.id ? HttpResponse.json(dane.raport) : nieZnaleziono(),
  ),
  http.get("/api/uruchomienia/:uruchomienieId", ({ params }) =>
    params.uruchomienieId === dane.uruchomienie.id ? HttpResponse.json(dane.uruchomienie) : nieZnaleziono(),
  ),
];
