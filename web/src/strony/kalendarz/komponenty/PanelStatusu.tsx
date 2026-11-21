import type { IndeksPlanu } from "../model/indeks.ts";
import type { Naruszenie, NaruszeniaPlanu } from "../model/naruszenia.ts";
import { Tygodnie } from "../model/tygodnie.ts";
import { SKROTY_DNI } from "../model/widok.ts";
import styles from "./PanelStatusu.module.css";

type PanelStatusuProps = {
  naruszenia: NaruszeniaPlanu;
  indeks: IndeksPlanu;
  wybrane: number | null;
  onWybierz: (naruszenie: Naruszenie) => void;
  onZamknij: () => void;
};

// Panel boczny z makiety: lista naruszeń pogrupowana wg reguły albo informacja o statusie planu.
export function PanelStatusu({ naruszenia, indeks, wybrane, onWybierz, onZamknij }: PanelStatusuProps) {
  const { status } = naruszenia;
  const tytul =
    status === "naruszenia"
      ? `Naruszenia reguł twardych · ${naruszenia.lista.length}`
      : status === "poprawny"
        ? "Reguły twarde · bez naruszeń"
        : "Status niesprawdzony";
  const regulyOk = naruszenia.regulyBezNaruszen();

  return (
    <aside id="panel-statusu" aria-label="Panel statusu planu" className={styles.panel}>
      <div className={styles.naglowek}>
        <h2 className={styles.tytul}>{tytul}</h2>
        <button type="button" aria-label="Zamknij panel" className={styles.zamknij} onClick={onZamknij}>
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            aria-hidden="true"
          >
            <path d="M6 6l12 12M18 6L6 18" />
          </svg>
        </button>
      </div>

      {status === "naruszenia" && (
        <>
          <p className={styles.wskazowka}>
            Plan jest niedopuszczalny, dopóki lista nie jest pusta. Kliknij pozycję, aby przejść do zajęć.
          </p>
          {naruszenia.wgReguly().map((grupa) => (
            <div key={grupa.nazwa} className={styles.grupa}>
              <div className={styles.grupaNaglowek}>
                <span>{grupa.etykieta}</span>
                <span>{grupa.naruszenia.length}</span>
              </div>
              {grupa.naruszenia.map((n) => (
                <PozycjaNaruszenia
                  key={n.indeks}
                  naruszenie={n}
                  naruszenia={naruszenia}
                  indeks={indeks}
                  wybrana={wybrane === n.indeks}
                  onWybierz={onWybierz}
                />
              ))}
            </div>
          ))}
        </>
      )}
      {status === "poprawny" && <div className={styles.ok}>Wszystkie reguły twarde są spełnione.</div>}
      {status === "niesprawdzony" && (
        <div className={styles.niesprawdzony}>
          <div className={styles.pogrubiony}>Walidator nie jest jeszcze podłączony</div>
          <div className={styles.cichy}>
            Plan ma wszystkie zajęcia na siatce, ale nikt nie sprawdził reguł twardych. Nie traktuj go jako
            poprawnego.
          </div>
        </div>
      )}

      <div className={styles.reguly}>
        <div className={styles.regulyTytul}>
          {status === "naruszenia"
            ? `Bez naruszeń · ${regulyOk.length} reguł`
            : status === "poprawny"
              ? "Sprawdzone · 10 reguł"
              : "Niesprawdzone · 10 reguł"}
        </div>
        <div className={styles.regulyLista}>
          {regulyOk.map((r) => (
            <span key={r} className={styles.regula}>
              {status === "niesprawdzony" ? "–" : "✓"} {r}
            </span>
          ))}
        </div>
      </div>
    </aside>
  );
}

function PozycjaNaruszenia({
  naruszenie,
  naruszenia,
  indeks,
  wybrana,
  onWybierz,
}: {
  naruszenie: Naruszenie;
  naruszenia: NaruszeniaPlanu;
  indeks: IndeksPlanu;
  wybrana: boolean;
  onWybierz: (naruszenie: Naruszenie) => void;
}) {
  const zajecia = naruszenia.zajeciaNaruszenia(naruszenie, indeks);
  const pierwsze = zajecia[0];
  const dzien = naruszenie.dzien ?? pierwsze?.dzien;
  const sloty =
    naruszenie.sloty && naruszenie.sloty.length
      ? [Math.min(...naruszenie.sloty), Math.max(...naruszenie.sloty)]
      : pierwsze
        ? [pierwsze.slotStart, pierwsze.slotStart + pierwsze.dlugosc - 1]
        : null;
  const sala = naruszenie.sala_id ?? (zajecia.length === 1 ? pierwsze?.salaId : undefined);
  const tygodnie = naruszenie.tygodnie ?? pierwsze?.tygodnie;
  const meta = [
    dzien ? SKROTY_DNI[dzien - 1] : null,
    sloty ? `sloty ${sloty[0]}–${sloty[1]}` : null,
    sala ?? "",
    tygodnie ? Tygodnie.opisZawsze(tygodnie) : null,
  ].filter((m): m is string => m !== null);

  return (
    <button
      type="button"
      aria-pressed={wybrana}
      className={wybrana ? `${styles.pozycja} ${styles.wybrana}` : styles.pozycja}
      onClick={() => onWybierz(naruszenie)}
    >
      <span className={styles.opis}>{naruszenie.opis}</span>
      <span>
        {zajecia.map((z) => `${z.id} · ${z.przedmiot} (${z.forma})`).join(", ") ||
          naruszenie.zajecia_id.join(", ")}
      </span>
      <span className={styles.meta}>
        {meta.map((m, i) => (
          <span key={i} className={m ? styles.metaPozycja : `${styles.metaPozycja} ${styles.metaPusta}`}>
            {m || "brak sali"}
          </span>
        ))}
      </span>
    </button>
  );
}
