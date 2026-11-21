import type { RouteObject } from "react-router";
import { Uklad } from "./Uklad.tsx";
import { StronaPulpit } from "../strony/pulpit/StronaPulpit.tsx";
import { StronaZestaw } from "../strony/StronaZestaw.tsx";
import { StronaRaport } from "../strony/StronaRaport.tsx";
import { StronaNieZnaleziono } from "../strony/StronaNieZnaleziono.tsx";

// Trasy odpowiadają ekranom z designu: 1 · Pulpit, 2 · Kalendarz, raport jakości.
// Filtry kalendarza (rocznik, prowadzący, sala, tydzień) będą w parametrach zapytania, nie w ścieżce.
export const trasy: RouteObject[] = [
  {
    element: <Uklad />,
    children: [
      { index: true, element: <StronaPulpit /> },
      { path: "zestawy/:zestawId", element: <StronaZestaw /> },
      // Kalendarz zajmuje całą szerokość okna (makieta „2 · Kalendarz”), pozostałe ekrany 1120 px.
      // Ładowany dopiero przy wejściu (Radix Popover, cmdk), żeby pulpit startował szybciej.
      {
        path: "uruchomienia/:uruchomienieId",
        handle: { szeroki: true },
        lazy: () =>
          import("../strony/kalendarz/StronaKalendarz.tsx").then((m) => ({ Component: m.StronaKalendarz })),
      },
      { path: "uruchomienia/:uruchomienieId/raport", element: <StronaRaport /> },
      { path: "*", element: <StronaNieZnaleziono /> },
    ],
  },
];
