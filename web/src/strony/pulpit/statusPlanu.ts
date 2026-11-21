import type { Schematy } from "../../api/klient.ts";

type Uruchomienie = Schematy["Uruchomienie"];
type Walidator = Schematy["Zdrowie"]["walidator"];

// Kolory statusów z briefu: czerwony tylko dla naruszeń twardych i błędów, zielony tylko dla „0 naruszeń”,
// neutralny dla stanów bez rozstrzygnięcia (w tym walidator „atrapa”), „brak” dla zestawu bez planu.
export type RodzajStatusu = "blad" | "ok" | "neutralny" | "brak";

export type StatusPlanu = { rodzaj: RodzajStatusu; tekst: string; znak: string };

const znaki: Record<RodzajStatusu, string> = { blad: "⚠", ok: "✓", neutralny: "ⓘ", brak: "–" };

export class StatusyPlanow {
  static zUruchomienia(
    uruchomienie: Uruchomienie | undefined,
    walidator: Walidator | undefined,
  ): StatusPlanu {
    const [rodzaj, tekst] = StatusyPlanow.rodzajITekst(uruchomienie, walidator);
    return { rodzaj, tekst, znak: znaki[rodzaj] };
  }

  // Ostatnie uruchomienie każdego zestawu; lista z API jest posortowana od najnowszych.
  static ostatnieWgZestawu(uruchomienia: Uruchomienie[]): Map<string, Uruchomienie> {
    const ostatnie = new Map<string, Uruchomienie>();
    for (const uruchomienie of uruchomienia) {
      if (!ostatnie.has(uruchomienie.zestaw_id)) {
        ostatnie.set(uruchomienie.zestaw_id, uruchomienie);
      }
    }
    return ostatnie;
  }

  private static rodzajITekst(
    uruchomienie: Uruchomienie | undefined,
    walidator: Walidator | undefined,
  ): [RodzajStatusu, string] {
    if (!uruchomienie) {
      return ["brak", "Brak planu"];
    }

    switch (uruchomienie.status) {
      case "w_kolejce":
        return ["neutralny", "W kolejce"];
      case "trwa":
        return ["neutralny", "Generowanie…"];
      case "blad":
        return ["blad", "Błąd generowania"];
      case "z_naruszeniami":
        return ["blad", `Plan z naruszeniami · ${uruchomienie.liczba_naruszen ?? "?"}`];
      case "poprawny":
        return walidator === "internal/validator"
          ? ["ok", "Plan poprawny · 0 naruszeń"]
          : ["neutralny", "Status niesprawdzony"];
    }
  }
}
