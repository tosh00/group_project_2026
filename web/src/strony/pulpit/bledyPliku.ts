import type { BladApi } from "../../api/klient.ts";

export type BladPliku = { miejsce: string; opis: string };

// Backend (model.Sprawdz) zwraca błędy jako „zajecia Z004: dozwolone_sale: nieznany identyfikator "X999"”.
// Makieta pokazuje je w dwóch kolumnach: miejsce w danych (`zajecia[Z004].dozwolone_sale`) i opis.
const wzorzecPola = /^(\w+) ([^:]+): (\w+): (.+)$/;
const wzorzecTabeli = /^([^:]+): (.+)$/;

export class BledyPliku {
  static zBleduApi(blad: BladApi): BladPliku[] {
    if (blad.szczegoly.length === 0) {
      return [{ miejsce: blad.kod, opis: blad.message }];
    }

    return blad.szczegoly.map(BledyPliku.zTekstu);
  }

  static zTekstu(tekst: string): BladPliku {
    const pole = wzorzecPola.exec(tekst);
    if (pole) {
      const [, tabela, id, nazwaPola, opis] = pole;
      return { miejsce: `${tabela}[${id}].${nazwaPola}`, opis: BledyPliku.zWielkiej(opis!) };
    }

    const tabela = wzorzecTabeli.exec(tekst);
    if (tabela) {
      return { miejsce: tabela[1]!, opis: BledyPliku.zWielkiej(tabela[2]!) };
    }

    return { miejsce: "plik", opis: BledyPliku.zWielkiej(tekst) };
  }

  private static zWielkiej(tekst: string): string {
    return tekst.charAt(0).toUpperCase() + tekst.slice(1);
  }
}
