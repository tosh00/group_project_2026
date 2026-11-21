import type { ReactNode } from "react";
import { NAZWY_FORM } from "../model/kolory.ts";
import { FORMY, type Filtry, type Wymiar } from "../model/typy.ts";
import styles from "./PanelFiltrow.module.css";

type PanelFiltrowProps = {
  filtry: Filtry;
  widokStudentow: boolean;
  zakresTygodnia: string; // np. „23.11–27.11”
  wyszukiwarka: ReactNode;
  onWymiar: (wymiar: Wymiar) => void;
  onZmien: (zmiana: Partial<Filtry>) => void;
};

const WYMIARY: { wymiar: Wymiar; etykieta: string }[] = [
  { wymiar: "rocznik", etykieta: "Rocznik" },
  { wymiar: "grupa", etykieta: "Grupa" },
  { wymiar: "prowadzacy", etykieta: "Prowadzący" },
  { wymiar: "sala", etykieta: "Sala" },
];

const TYGODNIE = Array.from({ length: 14 }, (_, i) => i + 1);

// Pasek filtrów z makiety: wymiar widoku, wyszukiwarka, formy, tydzień, „tylko naruszenia”, kolor kart.
export function PanelFiltrow({
  filtry,
  widokStudentow,
  zakresTygodnia,
  wyszukiwarka,
  onWymiar,
  onZmien,
}: PanelFiltrowProps) {
  const przelaczForme = (forma: (typeof FORMY)[number]) => {
    const ukryte = new Set(filtry.ukryteFormy);
    if (ukryte.has(forma)) {
      ukryte.delete(forma);
    } else {
      ukryte.add(forma);
    }
    onZmien({ ukryteFormy: ukryte });
  };

  return (
    <section aria-label="Filtry" className={styles.panel}>
      <div className={styles.wiersz}>
        <div role="group" aria-label="Wymiar widoku" className={styles.segmenty}>
          {WYMIARY.map(({ wymiar, etykieta }) => (
            <button
              key={wymiar}
              type="button"
              aria-pressed={filtry.wymiar === wymiar}
              className={filtry.wymiar === wymiar ? styles.segmentWlaczony : styles.segment}
              onClick={() => onWymiar(wymiar)}
            >
              {etykieta}
            </button>
          ))}
        </div>
        {wyszukiwarka}
        <span className={styles.separator} aria-hidden="true" />
        <div role="group" aria-label="Forma zajęć" className={styles.grupa}>
          <span className={styles.cichy}>Forma</span>
          {FORMY.map((forma) => {
            const wlaczona = !filtry.ukryteFormy.has(forma);
            return (
              <button
                key={forma}
                type="button"
                aria-pressed={wlaczona}
                title={NAZWY_FORM[forma]}
                className={wlaczona ? styles.forma : styles.formaWylaczona}
                onClick={() => przelaczForme(forma)}
              >
                {forma}
              </button>
            );
          })}
        </div>
      </div>
      <div className={styles.wiersz}>
        <div role="group" aria-label="Tydzień semestru" className={styles.tygodnie}>
          <span className={styles.etykietaTygodni}>Tydzień</span>
          <button
            type="button"
            aria-pressed={filtry.tydzien === 0}
            className={filtry.tydzien === 0 ? styles.tydzienWybrany : styles.tydzien}
            onClick={() => onZmien({ tydzien: 0 })}
          >
            Obecny
          </button>
          {TYGODNIE.map((t) => (
            <button
              key={t}
              type="button"
              aria-pressed={filtry.tydzien === t}
              className={
                filtry.tydzien === t
                  ? `${styles.tydzienWybrany} ${styles.numer}`
                  : `${styles.tydzien} ${styles.numer}`
              }
              onClick={() => onZmien({ tydzien: t })}
            >
              {t}
            </button>
          ))}
        </div>
        {zakresTygodnia && <span className={styles.zakres}>{zakresTygodnia}</span>}
        <div className={styles.prawo}>
          <button
            type="button"
            role="switch"
            aria-checked={filtry.tylkoNaruszenia}
            className={styles.przelacznik}
            onClick={() => onZmien({ tylkoNaruszenia: !filtry.tylkoNaruszenia })}
          >
            <span className={filtry.tylkoNaruszenia ? styles.torWlaczony : styles.tor}>
              <span className={styles.galka} />
            </span>
            Tylko naruszenia
          </button>
          <div role="group" aria-label="Kolor kart" className={styles.grupa}>
            <span className={styles.cichy}>Kolor wg</span>
            <div className={styles.segmenty}>
              <button
                type="button"
                aria-pressed={filtry.kolor === "grupa"}
                className={filtry.kolor === "grupa" ? styles.segmentWlaczony : styles.segment}
                onClick={() => onZmien({ kolor: "grupa" })}
              >
                {widokStudentow ? "grup" : "roczników"}
              </button>
              <button
                type="button"
                aria-pressed={filtry.kolor === "forma"}
                className={filtry.kolor === "forma" ? styles.segmentWlaczony : styles.segment}
                onClick={() => onZmien({ kolor: "forma" })}
              >
                form
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
