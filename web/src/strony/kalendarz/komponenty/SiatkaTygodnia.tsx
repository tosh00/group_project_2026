import type { CSSProperties } from "react";
import { Format } from "../../../lib/format.ts";
import przyciski from "../../../ui/przyciski.module.css";
import type { IndeksPlanu } from "../model/indeks.ts";
import type { NaruszeniaPlanu } from "../model/naruszenia.ts";
import { KalendarzSemestru } from "../model/semestr.ts";
import { LICZBA_SLOTOW, type Filtry, type Wymiar } from "../model/typy.ts";
import { NAZWY_DNI, SKROTY_DNI, type OpisWidoku, type WidokPlanu } from "../model/widok.ts";
import { KartaZajec } from "./KartaZajec.tsx";
import styles from "./SiatkaTygodnia.module.css";

type SiatkaTygodniaProps = {
  widok: WidokPlanu;
  opis: OpisWidoku;
  indeks: IndeksPlanu;
  naruszenia: NaruszeniaPlanu;
  filtry: Filtry;
  wskazane: string | null;
  obwiedzione: ReadonlySet<string>;
  onWskazanie: (zajeciaId: string | null) => void;
  onPrzejdz: (wymiar: Wymiar, id: string) => void;
  onWyczyscFiltry: () => void;
};

const sloty = Array.from({ length: LICZBA_SLOTOW }, (_, i) => i + 1);

export function SiatkaTygodnia({
  widok,
  opis,
  indeks,
  naruszenia,
  filtry,
  wskazane,
  obwiedzione,
  onWskazanie,
  onPrzejdz,
  onWyczyscFiltry,
}: SiatkaTygodniaProps) {
  const dni = widok.dni();
  const pusty = widok.widoczne.length === 0;

  return (
    <section aria-label="Siatka tygodnia" className={styles.siatka}>
      <div className={styles.naglowek}>
        <div className={styles.tytuly}>
          <div className={styles.tytul}>{opis.tytul}</div>
          <div className={styles.cichy}>{opis.podtytul}</div>
        </div>
        {filtry.wymiar === "sala" && <ZajetoscSali widok={widok} />}
        {filtry.wymiar === "prowadzacy" && <StatystykiProwadzacego widok={widok} />}
      </div>

      <div className={styles.przewijanie}>
        <div className={styles.wnetrze}>
          <div className={styles.wierszDni}>
            <div className={styles.etykietaSlot}>Slot</div>
            {dni.map((d) => (
              <div key={d.dzien} className={d.wolny ? `${styles.dzien} ${styles.dzienWolny}` : styles.dzien}>
                <span className={styles.nazwaDnia}>
                  {NAZWY_DNI[d.dzien - 1]}
                  {d.kalendarz?.data && d.kalendarz.numer !== null && (
                    <span className={styles.data}>
                      {d.kalendarz.odrabianie ? d.kalendarz.uwaga : KalendarzSemestru.data(d.kalendarz.data)}
                    </span>
                  )}
                </span>
                <span className={styles.liczbaZajec}>
                  {d.kalendarz?.numer === null
                    ? "brak zajęć"
                    : d.wolny
                      ? "dzień wolny"
                      : Format.ilosc(d.karty.length, "zajęcia", "zajęcia", "zajęć")}
                </span>
              </div>
            ))}
          </div>

          <div className={styles.tresc}>
            <div>
              {sloty.map((slot) => (
                <div
                  key={slot}
                  className={slot === 1 || slot >= 10 ? `${styles.czas} ${styles.niechciany}` : styles.czas}
                >
                  <span className={styles.numerSlotu}>{slot}</span>
                  <span className={styles.godzina}>{indeks.godziny(slot, slot)}</span>
                  {slot === 1 && <span className={styles.etykietaPory}>przed 9:00</span>}
                  {slot === 10 && <span className={styles.etykietaPory}>po 17:00</span>}
                </div>
              ))}
            </div>
            {dni.map((d) => (
              <div key={d.dzien} className={styles.kolumna}>
                {sloty.map((slot) => (
                  <div
                    key={slot}
                    className={
                      slot === 1 || slot >= 10 ? `${styles.wiersz} ${styles.niechciany}` : styles.wiersz
                    }
                  />
                ))}
                {d.bloki.map((b) => (
                  <div
                    key={`${b.rodzaj}-${b.slotOd}`}
                    className={styles[b.rodzaj]}
                    style={
                      {
                        "--blok-gora": `calc((${b.slotOd} - 1) * var(--wysokosc-slotu) + 2px)`,
                        "--blok-wysokosc": `calc(${b.slotDo - b.slotOd + 1} * var(--wysokosc-slotu) - 4px)`,
                      } as CSSProperties
                    }
                    aria-label={`${b.tytul || "Dzień wolny"}, ${SKROTY_DNI[d.dzien - 1]} sloty ${b.slotOd}–${b.slotDo}`}
                  >
                    {b.tytul && <span className={styles.blokTytul}>{b.tytul}</span>}
                    {b.opis && <span className={styles.blokOpis}>{b.opis}</span>}
                  </div>
                ))}
                {d.karty.map((karta) => (
                  <KartaZajec
                    key={karta.zajecia.id}
                    karta={karta}
                    indeks={indeks}
                    naruszenia={naruszenia.dlaZajec(karta.zajecia.id)}
                    kolorKart={filtry.kolor}
                    obwiedziona={obwiedzione.has(karta.zajecia.id)}
                    dymekOtwarty={wskazane === karta.zajecia.id}
                    onWskazanie={onWskazanie}
                    onPrzejdz={onPrzejdz}
                  />
                ))}
              </div>
            ))}

            {pusty && (
              <div className={styles.pustyNakladka}>
                <div className={styles.pusty}>
                  <div className={styles.pustyTytul}>
                    {filtry.tydzien
                      ? `Brak zajęć dla ${opis.nazwa} w tygodniu ${filtry.tydzien}`
                      : `Brak zajęć dla ${opis.nazwa} w tym tygodniu`}
                  </div>
                  <div className={styles.cichy}>
                    Zmień tydzień lub formę zajęć albo wyczyść filtry zawężające.
                  </div>
                  <button type="button" className={przyciski.glowny} onClick={onWyczyscFiltry}>
                    Wyczyść filtry zawężające
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}

function ZajetoscSali({ widok }: { widok: WidokPlanu }) {
  const { slotow, procent } = widok.zajetoscSali();
  return (
    <div className={styles.zajetosc}>
      <span className={styles.cichy}>Zajętość</span>
      <div
        role="meter"
        aria-label="Zajętość sali"
        aria-valuemin={0}
        aria-valuemax={60}
        aria-valuenow={slotow}
        className={styles.miernik}
      >
        <div className={styles.miernikWypelnienie} style={{ width: `${procent}%` }} />
      </div>
      <span className={styles.pogrubiony}>
        {slotow} z 60 slotów · {procent}%
      </span>
    </div>
  );
}

function StatystykiProwadzacego({ widok }: { widok: WidokPlanu }) {
  const { slotow, dniWolne, okienek } = widok.statystykiProwadzacego();
  return (
    <div className={styles.statystyki}>
      <span className={styles.pigulka}>
        <b>{slotow}</b> {Format.odmiana(slotow, "slot", "sloty", "slotów")} w tygodniu
      </span>
      <span className={styles.pigulka}>
        Dni wolne: <b>{dniWolne.length ? dniWolne.join(", ") : "brak"}</b>
      </span>
      <span className={styles.pigulka}>
        Okienka: <b>{okienek}</b>
      </span>
    </div>
  );
}
