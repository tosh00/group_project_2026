import { sse } from "msw";
import type { Schematy } from "../api/klient.ts";
import * as dane from "./dane.ts";

// Zdarzenia SSE z `GET /api/uruchomienia/{id}/postep`; `data` to JSON jako tekst, jak w backendzie.
type ZdarzeniaPostepu = { stan: string; postep: string; koniec: string };

// Osobny moduł, bo `sse()` wymaga EventSource już przy imporcie, a jsdom w testach go nie ma.
export const handlerySse = [
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
