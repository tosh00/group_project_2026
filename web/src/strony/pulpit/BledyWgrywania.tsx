import { useState } from "react";
import { Format } from "../../lib/format.ts";
import { Ikona } from "../../ui/Ikona.tsx";
import przyciski from "../../ui/przyciski.module.css";
import type { BladPliku } from "./bledyPliku.ts";
import styles from "./BledyWgrywania.module.css";

type BledyWgrywaniaProps = { plik: string; bledy: BladPliku[]; onWybierzPlik: () => void };

export function BledyWgrywania({ plik, bledy, onWybierzPlik }: BledyWgrywaniaProps) {
  const [skopiowano, setSkopiowano] = useState(false);

  const kopiuj = async () => {
    await navigator.clipboard.writeText(bledy.map((b) => `${b.miejsce}\t${b.opis}`).join("\n"));
    setSkopiowano(true);
  };

  return (
    <section aria-label="Błędy pliku" className={styles.karta}>
      <div className={styles.gora}>
        <span className={styles.ikona}>
          <Ikona nazwa="blad" rozmiar={28} />
        </span>
        <div className={styles.opis}>
          <div className={styles.tytul}>
            Nie wgrano pliku ·{" "}
            {Format.ilosc(bledy.length, "błąd blokujący", "błędy blokujące", "błędów blokujących")}
          </div>
          <div>{plik} · popraw plik i wgraj go ponownie</div>
        </div>
        <button type="button" className={`${przyciski.cichy} ${styles.naTle}`} onClick={kopiuj}>
          {skopiowano ? "Skopiowano" : "Kopiuj listę"}
        </button>
        <button type="button" className={przyciski.glowny} onClick={onWybierzPlik}>
          Wybierz inny plik
        </button>
      </div>
      <ul className={styles.lista}>
        {bledy.map((blad, i) => (
          <li key={`${blad.miejsce}-${i}`} className={styles.wiersz}>
            <code className={styles.miejsce}>{blad.miejsce}</code>
            <span className={styles.tresc}>{blad.opis}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}
