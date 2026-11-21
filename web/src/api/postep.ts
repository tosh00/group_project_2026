import { useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { Schematy } from "./klient.ts";
import { klucze } from "./zapytania.ts";

type Uruchomienie = Schematy["Uruchomienie"];
type Postep = Schematy["Postep"];

// Ile ostatnich punktów postępu trzymamy dla wykresu kosztu.
const LIMIT_HISTORII = 500;

// Postęp uruchomienia przez SSE (`GET /api/uruchomienia/{id}/postep`, opis zdarzeń w api/openapi.yaml).
// Zdarzenia `stan` i `koniec` aktualizują cache zapytania uruchomienia, więc reszta UI czyta zwykłe zapytanie.
export function usePostepUruchomienia(uruchomienieId: string | undefined) {
  const queryClient = useQueryClient();
  // Historia jest przypisana do uruchomienia, więc po zmianie `uruchomienieId` stara nie wycieka do nowego widoku.
  const [stan, setStan] = useState<{ uruchomienieId?: string; historia: Postep[] }>({ historia: [] });
  const historia = stan.uruchomienieId === uruchomienieId ? stan.historia : [];

  useEffect(() => {
    if (!uruchomienieId) {
      return;
    }

    const zrodlo = new EventSource(`/api/uruchomienia/${uruchomienieId}/postep`);
    const zapiszUruchomienie = (zdarzenie: MessageEvent<string>) => {
      const uruchomienie = JSON.parse(zdarzenie.data) as Uruchomienie;
      queryClient.setQueryData(klucze.uruchomienie(uruchomienieId), uruchomienie);
    };

    zrodlo.addEventListener("stan", zapiszUruchomienie);
    zrodlo.addEventListener("postep", (zdarzenie: MessageEvent<string>) => {
      const postep = JSON.parse(zdarzenie.data) as Postep;
      setStan((poprzedni) => ({
        uruchomienieId,
        historia: [...(poprzedni.uruchomienieId === uruchomienieId ? poprzedni.historia : []), postep].slice(
          -LIMIT_HISTORII,
        ),
      }));
    });
    zrodlo.addEventListener("koniec", (zdarzenie: MessageEvent<string>) => {
      zapiszUruchomienie(zdarzenie);
      zrodlo.close();
      void queryClient.invalidateQueries({ queryKey: ["uruchomienia"] });
    });

    return () => zrodlo.close();
  }, [uruchomienieId, queryClient]);

  return { historia, ostatni: historia.at(-1) };
}
