// Układ kart w kolumnie dnia: zajęcia nachodzące na siebie dostają osobne tory (algorytm z makiety).
// Grupa kart, które pośrednio na siebie nachodzą, dzieli liczbę torów, więc wszystkie mają tę samą szerokość.

export type Pozycja = { start: number; dlugosc: number };
export type PozycjaWTorze<T> = { element: T; tor: number; torow: number };

export class UkladTorow {
  static rozloz<T>(elementy: readonly T[], pozycja: (element: T) => Pozycja): PozycjaWTorze<T>[] {
    const posortowane = [...elementy].sort((a, b) => {
      const pa = pozycja(a);
      const pb = pozycja(b);
      return pa.start - pb.start || pb.dlugosc - pa.dlugosc;
    });

    const wynik: PozycjaWTorze<T>[] = [];
    let skupisko: PozycjaWTorze<T>[] = [];
    let konceTorow: number[] = [];
    let koniecSkupiska = 0;

    const zamknijSkupisko = () => {
      for (const p of skupisko) {
        p.torow = konceTorow.length;
      }
      wynik.push(...skupisko);
      skupisko = [];
      konceTorow = [];
      koniecSkupiska = 0;
    };

    for (const element of posortowane) {
      const { start, dlugosc } = pozycja(element);
      if (skupisko.length > 0 && start >= koniecSkupiska) {
        zamknijSkupisko();
      }

      let tor = konceTorow.findIndex((koniec) => koniec <= start);
      if (tor === -1) {
        tor = konceTorow.length;
        konceTorow.push(0);
      }
      konceTorow[tor] = start + dlugosc;
      koniecSkupiska = Math.max(koniecSkupiska, start + dlugosc);
      skupisko.push({ element, tor, torow: 0 });
    }
    zamknijSkupisko();

    return wynik;
  }
}
