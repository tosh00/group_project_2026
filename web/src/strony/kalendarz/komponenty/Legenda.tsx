import type { CSSProperties } from "react";
import { Format } from "../../../lib/format.ts";
import type { IndeksPlanu } from "../model/indeks.ts";
import { Kolory, KOLORY_FORM, NAZWY_FORM } from "../model/kolory.ts";
import { FORMY, type Filtry, type Wymiar } from "../model/typy.ts";
import type { WidokPlanu } from "../model/widok.ts";
import styles from "./Legenda.module.css";

type LegendaProps = {
  indeks: IndeksPlanu;
  widok: WidokPlanu;
  filtry: Filtry;
  rocznikWidoku: string;
  onPrzejdz: (wymiar: Wymiar, id: string) => void;
};

type Pozycja = {
  klucz: string;
  etykieta: string;
  opis: string;
  kolor: string;
  znak: string;
  licznik: string;
  aktywna: boolean;
  wcisnieta: boolean;
  cel?: [Wymiar, string];
};

export function Legenda({ indeks, widok, filtry, rocznikWidoku, onPrzejdz }: LegendaProps) {
  const ile = (warunek: (z: WidokPlanu["widoczne"][number]) => boolean) =>
    widok.widoczne.filter(warunek).length;
  let tytul: string;
  let pozycje: Pozycja[];

  if (filtry.kolor === "forma") {
    tytul = "Kolory: formy zajęć";
    pozycje = FORMY.map((f) => ({
      klucz: f,
      etykieta: NAZWY_FORM[f],
      opis: `litera ${f}`,
      kolor: KOLORY_FORM[f],
      znak: f,
      licznik: String(ile((z) => z.forma === f)),
      aktywna: true,
      wcisnieta: false,
    }));
  } else if (widok.widokStudentow) {
    tytul = `Kolory: grupy rocznika ${rocznikWidoku}`;
    const widoczneGrupy = filtry.wymiar === "grupa" ? indeks.grupyPowiazane(filtry.id) : null;
    pozycje = (indeks.grupyRocznika.get(rocznikWidoku) ?? []).map((g) => {
      const calyRocznik = g.typ === "rocznik";
      const pokazana = !widoczneGrupy || widoczneGrupy.has(g.id);
      return {
        klucz: g.id,
        etykieta: calyRocznik ? `Cały rocznik ${rocznikWidoku}` : g.nazwa,
        opis: `${calyRocznik ? (g.podzial ?? "wykłady i zajęcia wspólne") : g.id} · ${g.liczba_studentow} os.`,
        kolor: indeks.kolorGrupy.get(g.id) ?? KOLORY_FORM.W,
        znak: calyRocznik ? "R" : indeks.skrotGrupy(g.id).slice(-3),
        licznik: pokazana ? `${ile((z) => z.grupaId === g.id)} zaj.` : "pokaż",
        aktywna: pokazana,
        wcisnieta: filtry.wymiar === "grupa" && filtry.id === g.id,
        cel: calyRocznik ? ["rocznik", rocznikWidoku] : ["grupa", g.id],
      };
    });
  } else {
    tytul = "Kolory: roczniki";
    const roczniki = new Set(widok.zajeciaWidoku.map((z) => z.rocznikId));
    pozycje = indeks.roczniki
      .filter((r) => roczniki.has(r.id))
      .map((r) => ({
        klucz: r.id,
        etykieta: `Rocznik ${r.id}`,
        opis: `${r.nazwa} · ${r.liczba_studentow} os.`,
        kolor: indeks.kolorRocznika.get(r.id) ?? KOLORY_FORM.W,
        znak: r.id.slice(-1),
        licznik: `${ile((z) => z.rocznikId === r.id)} zaj.`,
        aktywna: true,
        wcisnieta: false,
        cel: ["rocznik", r.id],
      }));
  }

  return (
    <aside aria-labelledby="legenda-h" className={styles.legenda}>
      <h2 id="legenda-h" className={styles.tytul}>
        Legenda
      </h2>
      <div>
        <div className={styles.sekcja}>{tytul}</div>
        <ul className={styles.lista}>
          {pozycje.map((p) => (
            <li key={p.klucz}>
              <button
                type="button"
                disabled={!p.cel}
                aria-pressed={p.wcisnieta}
                className={`${styles.pozycja} ${p.wcisnieta ? styles.wcisnieta : ""} ${p.aktywna ? "" : styles.nieaktywna}`}
                onClick={() => p.cel && onPrzejdz(...p.cel)}
              >
                <span
                  className={styles.probka}
                  style={
                    {
                      "--probka-tlo": Kolory.rgba(p.kolor, 0.13),
                      "--probka-ramka": Kolory.rgba(p.kolor, 0.45),
                      "--probka-pasek": p.kolor,
                    } as CSSProperties
                  }
                >
                  {p.znak}
                </span>
                <span className={styles.opisPozycji}>
                  <span className={styles.etykieta}>{p.etykieta}</span>
                  <span className={styles.podpis}>{p.opis}</span>
                </span>
                <span className={styles.licznik}>{p.licznik}</span>
              </button>
            </li>
          ))}
        </ul>
      </div>

      {filtry.kolor !== "forma" && (
        <div>
          <div className={styles.sekcja}>Forma zajęć (litera na karcie)</div>
          <div className={styles.formy}>
            {FORMY.map((f) => (
              <span key={f} className={styles.forma}>
                <span className={styles.odznaka}>{f}</span>
                {NAZWY_FORM[f]}
              </span>
            ))}
          </div>
        </div>
      )}

      <div>
        <div className={styles.sekcja}>Oznaczenia</div>
        <ul className={styles.oznaczenia}>
          <li>
            <span className={`${styles.wzor} ${styles.wzorNiechciany}`} aria-hidden="true" />
            Sloty niechciane: przed 9:00 i po 17:00
          </li>
          <li>
            <span className={`${styles.wzor} ${styles.wzorNaruszenie}`} aria-hidden="true" />
            Naruszenie reguły twardej (ramka i „!”)
          </li>
          <li>
            <span className={`${styles.wzor} ${styles.wzorTygodnie}`} aria-hidden="true" />
            Zajęcia nie przez cały semestr
          </li>
          {filtry.wymiar === "prowadzacy" && (
            <>
              <li>
                <span className={`${styles.wzor} ${styles.wzorWykluczone}`} aria-hidden="true" />
                Godziny wykluczone przez prowadzącego
              </li>
              <li>
                <span className={`${styles.wzor} ${styles.wzorOkienko}`} aria-hidden="true" />
                Okienko między zajęciami
              </li>
              <li>
                <span className={`${styles.wzor} ${styles.wzorWolny}`} aria-hidden="true" />
                Dzień wolny
              </li>
            </>
          )}
        </ul>
      </div>
      <p className={styles.uwaga}>
        {Format.ilosc(widok.widoczne.length, "zajęcia", "zajęcia", "zajęć")} na siatce
      </p>
    </aside>
  );
}
