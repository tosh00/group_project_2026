import { Link } from "react-router";
import type { Schematy } from "../../../api/klient.ts";
import type { usePostepUruchomienia } from "../../../api/postep.ts";
import przyciski from "../../../ui/przyciski.module.css";
import styles from "./PostepUruchomienia.module.css";

type PostepUruchomieniaProps = {
  uruchomienie: Schematy["Uruchomienie"];
  postep: ReturnType<typeof usePostepUruchomienia>;
  zatrzymywanie: boolean;
  onZatrzymaj: () => void;
};

const FAZY: { faza: Schematy["Faza"]; etykieta: string }[] = [
  { faza: "konstrukcja", etykieta: "Układanie zajęć" },
  { faza: "optymalizacja", etykieta: "Ulepszanie planu" },
  { faza: "walidacja", etykieta: "Sprawdzanie" },
];

const sekundy = new Intl.NumberFormat("pl-PL", { maximumFractionDigits: 1, minimumFractionDigits: 1 });

// Generowanie w toku (brief: kolejka → konstrukcja → optymalizacja → walidacja). Makieta nie ma tego ekranu,
// więc korzysta z elementów pulpitu: karta, pasek postępu, kroki, przycisk „Zatrzymaj”.
export function PostepUruchomienia({
  uruchomienie,
  postep,
  zatrzymywanie,
  onZatrzymaj,
}: PostepUruchomieniaProps) {
  const { ostatni } = postep;
  const wKolejce = uruchomienie.status === "w_kolejce";
  const faza = ostatni?.faza ?? uruchomienie.faza ?? "konstrukcja";
  const indeksFazy = FAZY.findIndex((f) => f.faza === faza);
  const ulamek = ostatni && ostatni.iteracji > 0 ? ostatni.iteracja / ostatni.iteracji : null;

  return (
    <section aria-label="Generowanie planu" aria-busy="true" className={styles.karta}>
      <div className={styles.gora}>
        <div className={styles.opis}>
          <div className={styles.nadtytul}>Generowanie planu</div>
          <h1 className={styles.tytul}>
            {wKolejce ? "W kolejce…" : `${FAZY[indeksFazy]?.etykieta ?? "Trwa"}…`}
          </h1>
          <div className={styles.cichy}>
            {wKolejce
              ? "Serwer liczy jeden plan naraz. Generowanie zacznie się, gdy skończy się poprzednie."
              : "Generowanie trwa od ułamka sekundy do około minuty."}
          </div>
        </div>
        <button type="button" className={przyciski.cichy} onClick={onZatrzymaj} disabled={zatrzymywanie}>
          {zatrzymywanie ? "Zatrzymuję…" : "Zatrzymaj"}
        </button>
      </div>

      <div
        role="progressbar"
        aria-label="Postęp generowania"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={ulamek === null ? undefined : Math.round(ulamek * 100)}
        className={styles.pasek}
      >
        {ulamek === null ? (
          <div className={styles.pasekNieokreslony} />
        ) : (
          <div className={styles.pasekWypelnienie} style={{ width: `${Math.round(ulamek * 100)}%` }} />
        )}
      </div>

      <ol className={styles.kroki}>
        <li className={wKolejce ? styles.trwa : styles.gotowy}>W kolejce</li>
        {FAZY.map((f, i) => (
          <li
            key={f.faza}
            className={
              wKolejce || i > indeksFazy ? styles.czeka : i === indeksFazy ? styles.trwa : styles.gotowy
            }
          >
            {f.etykieta}
          </li>
        ))}
      </ol>

      {ostatni && <p className={styles.cichy}>Trwa już {sekundy.format(ostatni.czas_ms / 1000)} s.</p>}
    </section>
  );
}

export function BladUruchomienia({ uruchomienie }: { uruchomienie: Schematy["Uruchomienie"] }) {
  return (
    <section aria-label="Błąd uruchomienia" className={`${styles.karta} ${styles.kartaBlad}`}>
      <div className={styles.nadtytulBlad}>Błąd generowania</div>
      <h1 className={styles.tytul}>Nie udało się wygenerować planu</h1>
      <p>{uruchomienie.blad ?? "Uruchomienie zakończyło się błędem."}</p>
      <div>
        <Link to="/" className={przyciski.glowny}>
          Wróć do pulpitu
        </Link>
      </div>
    </section>
  );
}
