import type { CSSProperties } from "react";
import type { IndeksPlanu } from "../model/indeks.ts";
import { Kolory, NAZWY_FORM } from "../model/kolory.ts";
import type { Naruszenie } from "../model/naruszenia.ts";
import { NaruszeniaPlanu } from "../model/naruszenia.ts";
import { Tygodnie } from "../model/tygodnie.ts";
import { NAZWY_DNI, SKROTY_DNI, type KartaNaSiatce } from "../model/widok.ts";
import type { KolorKart, Wymiar } from "../model/typy.ts";
import styles from "./KartaZajec.module.css";

type KartaZajecProps = {
  karta: KartaNaSiatce;
  indeks: IndeksPlanu;
  naruszenia: Naruszenie[];
  kolorKart: KolorKart;
  obwiedziona: boolean;
  dymekOtwarty: boolean;
  onWskazanie: (zajeciaId: string | null) => void;
  onPrzejdz: (wymiar: Wymiar, id: string) => void;
};

// Karta zajęć z makiety. Tryb zależy od liczby torów w kolumnie: ≤2 pełna, 3 średnia, ≥4 zwarta.
export function KartaZajec({
  karta,
  indeks,
  naruszenia,
  kolorKart,
  obwiedziona,
  dymekOtwarty,
  onWskazanie,
  onPrzejdz,
}: KartaZajecProps) {
  const { zajecia: z, tor, torow, kolor } = karta;
  const tryb = torow <= 2 ? "pelna" : torow === 3 ? "srednia" : "zwarta";
  const koniec = z.slotStart + z.dlugosc - 1;
  const tygodnie = Tygodnie.opis(z.tygodnie);
  const prowadzacy = z.prowadzacy.map((id) => indeks.nazwaProwadzacego(id)).join(", ");
  const naruszona = naruszenia.length > 0;
  const pierwszeNaruszenie = naruszenia[0];
  const dymekPod = koniec <= 8;
  const dymekZPrawej = z.dzien >= 4;

  const styl = {
    "--karta-gora": `calc((${z.slotStart} - 1) * var(--wysokosc-slotu) + 2px)`,
    "--karta-wysokosc": `calc(${z.dlugosc} * var(--wysokosc-slotu) - 4px)`,
    "--karta-lewo": `${(tor / torow) * 100}%`,
    "--karta-szerokosc": `${100 / torow}%`,
    "--karta-tlo": Kolory.rgba(kolor, 0.13),
    "--karta-ramka": Kolory.rgba(kolor, 0.4),
    "--karta-pasek": kolor,
    "--karta-etykieta-ramka": Kolory.rgba(kolor, 0.75),
    "--odznaka-tlo": kolorKart === "forma" ? kolor : "var(--kolor-odznaka)",
  } as CSSProperties;

  const klasy = [
    styles.karta,
    naruszona && styles.naruszona,
    obwiedziona && styles.obwiedziona,
    karta.wygaszona && styles.wygaszona,
    dymekOtwarty && styles.wskazana,
  ]
    .filter(Boolean)
    .join(" ");

  const odznaka = <span className={styles.odznaka}>{z.forma}</span>;
  const tygodnieEtykieta = tygodnie.pelny && <span className={styles.tygodnie}>{tygodnie.pelny}</span>;
  const grupaCel: [Wymiar, string] = indeks.czyCalyRocznik(z.grupaId)
    ? ["rocznik", z.rocznikId]
    : ["grupa", z.grupaId];

  return (
    <div
      className={klasy}
      style={styl}
      aria-label={`${z.przedmiot}, ${NAZWY_FORM[z.forma]}, sala ${z.salaId}, ${prowadzacy}, ${karta.etykietaGrupy}, ${SKROTY_DNI[z.dzien - 1]} sloty ${z.slotStart}–${koniec}${pierwszeNaruszenie ? `, naruszenie: ${pierwszeNaruszenie.opis}` : ""}`}
      onMouseEnter={() => onWskazanie(z.id)}
      onMouseLeave={() => onWskazanie(null)}
      onFocus={() => onWskazanie(z.id)}
      onBlur={() => onWskazanie(null)}
    >
      <div className={styles[tryb]}>
        {tryb === "pelna" && (
          <>
            <div className={styles.naglowek}>
              {odznaka}
              <span className={styles.przedmiot}>{z.przedmiot}</span>
            </div>
            <div className={styles.linia}>
              <button
                type="button"
                className={styles.link}
                onClick={() => onPrzejdz("sala", z.salaId)}
                title={`Plan sali ${z.salaId}`}
              >
                {z.salaId}
              </button>
              {" · "}
              <button
                type="button"
                className={`${styles.link} ${styles.grupa}`}
                onClick={() => onPrzejdz(...grupaCel)}
                title={`Plan: ${karta.etykietaGrupy}`}
              >
                {karta.etykietaGrupy}
              </button>
            </div>
            <div className={`${styles.linia} ${styles.cicha}`}>
              {z.prowadzacy.map((id, i) => (
                <span key={id}>
                  {i > 0 && ", "}
                  <button
                    type="button"
                    className={styles.link}
                    onClick={() => onPrzejdz("prowadzacy", id)}
                    title={`Plan prowadzącego ${indeks.nazwaProwadzacego(id)}`}
                  >
                    {indeks.nazwaProwadzacego(id)}
                  </button>
                </span>
              ))}
            </div>
            {tygodnieEtykieta}
          </>
        )}
        {tryb === "srednia" && (
          <>
            <div className={styles.naglowekSredni}>
              {odznaka}
              <span className={styles.skrot}>{karta.skrotGrupy}</span>
            </div>
            <span className={styles.przedmiotSredni}>{z.przedmiot}</span>
            <span className={styles.linia}>{z.salaId}</span>
            {tygodnieEtykieta}
          </>
        )}
        {tryb === "zwarta" && (
          <>
            {odznaka}
            <span className={styles.skrotZwarty}>{karta.skrotGrupy}</span>
            <span className={styles.salaZwarta}>{z.salaId}</span>
            {tygodnie.krotki && <span className={styles.tygodnieZwarte}>{tygodnie.krotki}</span>}
          </>
        )}
      </div>

      {naruszona && (
        <span role="img" aria-label="Naruszenie reguły twardej" className={styles.wykrzyknik}>
          <svg
            width="10"
            height="10"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="3.2"
            strokeLinecap="round"
            aria-hidden="true"
          >
            <path d="M12 6v8M12 18.5h.01" />
          </svg>
        </span>
      )}

      {dymekOtwarty && (
        <div
          role="tooltip"
          className={`${styles.dymek} ${dymekPod ? styles.dymekPod : styles.dymekNad} ${dymekZPrawej ? styles.dymekPrawo : styles.dymekLewo}`}
        >
          <div className={styles.dymekNaglowek}>
            {odznaka}
            <span className={styles.dymekTytul}>{z.przedmiot}</span>
          </div>
          <div className={styles.dymekCichy}>
            {NAZWY_FORM[z.forma].charAt(0).toUpperCase() + NAZWY_FORM[z.forma].slice(1)} ·{" "}
            {indeks.czyCalyRocznik(z.grupaId) ? `cały rocznik ${z.rocznikId}` : `grupa ${z.grupaId}`}
          </div>
          <div>
            {NAZWY_DNI[z.dzien - 1]}, {indeks.godziny(z.slotStart, koniec)}
          </div>
          <div>{Tygodnie.opisDlaLudzi(z.tygodnie)}</div>
          <div>
            Sala {z.salaId} · {prowadzacy}
          </div>
          {naruszenia.map((n) => (
            <div key={n.indeks} className={styles.dymekNaruszenie}>
              <svg
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M10.3 3.9L1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z" />
                <path d="M12 9v4M12 17h.01" />
              </svg>
              <span>
                <b>{NaruszeniaPlanu.etykieta(n.nazwa)}:</b> {n.opis}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
