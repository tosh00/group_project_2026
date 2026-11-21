// Formatowanie liczb, dat i odmiany po polsku w jednym miejscu, żeby wszystkie ekrany pisały to samo.
export class Format {
  private static readonly liczby = new Intl.NumberFormat("pl-PL", {
    maximumFractionDigits: 0,
    useGrouping: "always",
  });
  private static readonly godzina = new Intl.DateTimeFormat("pl-PL", { hour: "2-digit", minute: "2-digit" });
  private static readonly data = new Intl.DateTimeFormat("pl-PL", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });

  // 1284.4 → „1 284”
  static liczba(wartosc: number): string {
    return Format.liczby.format(wartosc);
  }

  // Odmiana rzeczownika po liczebniku: (1, „zajęcie”, „zajęcia”, „zajęć”) → „zajęcie”.
  static odmiana(liczba: number, pojedyncza: string, kilka: string, wiele: string): string {
    if (liczba === 1) {
      return pojedyncza;
    }

    const jednosci = liczba % 10;
    const dziesiatki = liczba % 100;
    return jednosci >= 2 && jednosci <= 4 && (dziesiatki < 12 || dziesiatki > 14) ? kilka : wiele;
  }

  // „174 zajęcia”, „1 błąd”, „23 błędy”.
  static ilosc(liczba: number, pojedyncza: string, kilka: string, wiele: string): string {
    return `${Format.liczba(liczba)} ${Format.odmiana(liczba, pojedyncza, kilka, wiele)}`;
  }

  // Dziś: „dziś, 11:05”, wcześniej: „14.11.2026”.
  static dataWzgledna(iso: string, teraz: Date = new Date()): string {
    const chwila = new Date(iso);
    if (chwila.toDateString() === teraz.toDateString()) {
      return `dziś, ${Format.godzina.format(chwila)}`;
    }

    return Format.data.format(chwila);
  }
}
