import { useSearchParams } from "react-router";
import { FORMY, type Filtry, type Forma, type Wymiar } from "./model/typy.ts";

const WYMIARY: readonly Wymiar[] = ["rocznik", "grupa", "prowadzacy", "sala"];

// Stan filtrów w parametrach URL (brief: link do konkretnego widoku):
//   ?widok=prowadzacy&id=P07&tydzien=6&bez=LEK,SEM&naruszenia=1&kolor=forma
// Zmiana wymiaru (czyj plan) dodaje wpis do historii, więc „Wstecz” wraca do poprzedniego widoku.
// Filtry zawężające podmieniają bieżący wpis, żeby historia nie puchła od kliknięć.
export class ParametryFiltrow {
  static odczytaj(parametry: URLSearchParams, domyslnyRocznik: string): Filtry {
    const wymiar = parametry.get("widok") as Wymiar | null;
    const poprawnyWymiar = wymiar && WYMIARY.includes(wymiar) ? wymiar : "rocznik";
    const tydzien = Number(parametry.get("tydzien") ?? 0);
    const ukryte = (parametry.get("bez") ?? "")
      .split(",")
      .filter((f): f is Forma => FORMY.includes(f as Forma));
    return {
      wymiar: poprawnyWymiar,
      id: parametry.get("id") ?? (poprawnyWymiar === "rocznik" ? domyslnyRocznik : ""),
      tydzien: Number.isInteger(tydzien) && tydzien >= 0 && tydzien <= 14 ? tydzien : 0,
      ukryteFormy: new Set(ukryte),
      tylkoNaruszenia: parametry.get("naruszenia") === "1",
      kolor: parametry.get("kolor") === "forma" ? "forma" : "grupa",
    };
  }

  static zapisz(filtry: Filtry): URLSearchParams {
    const p = new URLSearchParams({ widok: filtry.wymiar, id: filtry.id });
    if (filtry.tydzien) {
      p.set("tydzien", String(filtry.tydzien));
    }
    if (filtry.ukryteFormy.size) {
      p.set("bez", FORMY.filter((f) => filtry.ukryteFormy.has(f)).join(","));
    }
    if (filtry.tylkoNaruszenia) {
      p.set("naruszenia", "1");
    }
    if (filtry.kolor === "forma") {
      p.set("kolor", "forma");
    }
    return p;
  }
}

export function useFiltryKalendarza(domyslnyRocznik: string) {
  const [parametry, setParametry] = useSearchParams();
  const filtry = ParametryFiltrow.odczytaj(parametry, domyslnyRocznik);

  const zmien = (zmiana: Partial<Filtry>) =>
    setParametry(ParametryFiltrow.zapisz({ ...filtry, ...zmiana }), { replace: true });

  // Przejście do planu innej osoby, grupy lub sali: nowy wpis w historii.
  const przejdz = (wymiar: Wymiar, id: string, zmiana: Partial<Filtry> = {}) =>
    setParametry(ParametryFiltrow.zapisz({ ...filtry, ...zmiana, wymiar, id }));

  const wyczyscZawezajace = () => zmien({ tydzien: 0, ukryteFormy: new Set(), tylkoNaruszenia: false });

  return { filtry, zmien, przejdz, wyczyscZawezajace };
}
