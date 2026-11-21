import { useRef, useState } from "react";
import { BladApi, type Schematy } from "../../api/klient.ts";
import { useWgrajZestaw } from "../../api/zapytania.ts";
import { BledyPliku, type BladPliku } from "./bledyPliku.ts";

// Kroki z makiety „1c · Wgrywanie”: plik czytamy i parsujemy w przeglądarce, spójność sprawdza serwer.
export type KrokWgrywania = "json" | "serwer";

export type StanWgrywania =
  | { etap: "bezczynny" }
  | { etap: "wgrywanie"; plik: string; krok: KrokWgrywania }
  | { etap: "blad"; plik: string; bledy: BladPliku[] };

export function useWgrywaniePliku() {
  const wgrajZestaw = useWgrajZestaw();
  const [stan, setStan] = useState<StanWgrywania>({ etap: "bezczynny" });
  const przerwanie = useRef<AbortController | null>(null);

  // Zwraca wgrany zestaw (podsumowanie policzone przez serwer z pliku) albo null przy błędzie i anulowaniu.
  const wgraj = async (plik: File): Promise<Schematy["Zestaw"] | null> => {
    przerwanie.current?.abort();
    const kontroler = new AbortController();
    przerwanie.current = kontroler;

    setStan({ etap: "wgrywanie", plik: plik.name, krok: "json" });
    let dane: Schematy["DaneWejsciowe"];
    try {
      dane = JSON.parse(await plik.text()) as Schematy["DaneWejsciowe"];
    } catch (blad) {
      const opis = blad instanceof SyntaxError ? blad.message : "Nie udało się odczytać pliku.";
      setStan({
        etap: "blad",
        plik: plik.name,
        bledy: [{ miejsce: "plik", opis: `Niepoprawny JSON: ${opis}` }],
      });
      return null;
    }
    if (kontroler.signal.aborted) {
      return null;
    }

    setStan({ etap: "wgrywanie", plik: plik.name, krok: "serwer" });
    try {
      const zestaw = await wgrajZestaw.mutateAsync({ dane, nazwa: plik.name, sygnal: kontroler.signal });
      setStan({ etap: "bezczynny" });
      return zestaw;
    } catch (blad) {
      if (kontroler.signal.aborted) {
        return null;
      }

      const bledy =
        blad instanceof BladApi
          ? BledyPliku.zBleduApi(blad)
          : [{ miejsce: "serwer", opis: "Brak połączenia z serwerem." }];
      setStan({ etap: "blad", plik: plik.name, bledy });
      return null;
    }
  };

  const anuluj = () => {
    przerwanie.current?.abort();
    setStan({ etap: "bezczynny" });
  };

  return { stan, wgraj, anuluj };
}
