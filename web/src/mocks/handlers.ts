import { http, HttpResponse, sse } from "msw";
import type { Schematy } from "../api/klient.ts";
import * as dane from "./dane.ts";

type Blad = Schematy["Blad"];

// Zdarzenia SSE z `GET /api/uruchomienia/{id}/postep`; `data` to JSON jako tekst, jak w backendzie.
type ZdarzeniaPostepu = { stan: string; postep: string; koniec: string };

const nieZnaleziono = () =>
  HttpResponse.json<Blad>({ kod: "nie_znaleziono", komunikat: "Nie ma takiego zasobu." }, { status: 404 });

// Te same handlery działają w przeglądarce (`pnpm dev:mock`) i w testach (`src/mocks/server.ts`).
// Testy nadpisują je przez `server.use(...)`, np. żeby sprawdzić błąd 422.
export const handlers = [
  http.get("/api/zdrowie", () => HttpResponse.json(dane.zdrowie)),
  http.get("/api/wagi/domyslne", () => HttpResponse.json(dane.wagiDomyslne)),
  http.get("/api/zestawy", () => HttpResponse.json([dane.zestaw])),
  http.get("/api/zestawy/:zestawId", ({ params }) =>
    params.zestawId === dane.zestaw.id ? HttpResponse.json(dane.zestaw) : nieZnaleziono(),
  ),
  http.get("/api/uruchomienia", () => HttpResponse.json([dane.uruchomienie])),
  http.get("/api/uruchomienia/:uruchomienieId", ({ params }) =>
    params.uruchomienieId === dane.uruchomienie.id ? HttpResponse.json(dane.uruchomienie) : nieZnaleziono(),
  ),
  sse<ZdarzeniaPostepu>("/api/uruchomienia/:uruchomienieId/postep", ({ client }) => {
    const iteracji = 1_000_000;
    client.send({
      event: "stan",
      data: JSON.stringify({ ...dane.uruchomienie, status: "trwa", faza: "optymalizacja" }),
    });
    for (let krok = 1; krok <= 10; krok++) {
      const postep: Schematy["Postep"] = {
        faza: "optymalizacja",
        iteracja: (iteracji / 10) * krok,
        iteracji,
        koszt: 900 - krok * 48.75,
        naruszenia: 0,
        czas_ms: krok * 400,
      };
      client.send({ event: "postep", data: JSON.stringify(postep) });
    }
    client.send({ event: "koniec", data: JSON.stringify(dane.uruchomienie) });
    client.close();
  }),
];
