import { Ikona } from "../../ui/Ikona.tsx";
import przyciski from "../../ui/przyciski.module.css";
import type { KrokWgrywania } from "./useWgrywaniePliku.ts";
import styles from "./PostepWgrywania.module.css";

type StanKroku = "gotowy" | "trwa" | "czeka";

type PostepWgrywaniaProps = { plik: string; krok: KrokWgrywania; onAnuluj: () => void };

export function PostepWgrywania({ plik, krok, onAnuluj }: PostepWgrywaniaProps) {
  const kroki: { etykieta: string; stan: StanKroku }[] = [
    { etykieta: "Wysłano", stan: "gotowy" },
    { etykieta: "Poprawny JSON", stan: krok === "json" ? "trwa" : "gotowy" },
    { etykieta: "Spójność tabel i odwołań", stan: krok === "serwer" ? "trwa" : "czeka" },
    { etykieta: "Podsumowanie", stan: "czeka" },
  ];

  return (
    <section aria-label="Wgrywanie pliku" aria-busy="true" className={styles.karta}>
      <div className={styles.gora}>
        <div className={styles.ikona}>
          <Ikona nazwa="plik" rozmiar={24} grubosc={1.8} />
        </div>
        <div className={styles.opis}>
          <div className={styles.tytul}>Sprawdzam plik…</div>
          <div className={styles.cichy}>{plik}</div>
        </div>
        <button type="button" className={przyciski.cichy} onClick={onAnuluj}>
          Anuluj
        </button>
      </div>
      <div role="progressbar" aria-label="Postęp wgrywania" className={styles.pasek}>
        <div className={styles.pasekWypelnienie} />
      </div>
      <ol className={styles.kroki}>
        {kroki.map(({ etykieta, stan }) => (
          <li key={etykieta} className={styles[stan]}>
            {stan === "gotowy" && <Ikona nazwa="ok" grubosc={2.2} />}
            {stan === "trwa" && <span className={styles.krecidlo} aria-hidden="true" />}
            {stan === "czeka" && <span className={styles.kolko} aria-hidden="true" />}
            {etykieta}
          </li>
        ))}
      </ol>
    </section>
  );
}
