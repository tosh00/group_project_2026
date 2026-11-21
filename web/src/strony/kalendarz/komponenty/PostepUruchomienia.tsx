import { Link } from "react-router";
import type { Schematy } from "../../../api/klient.ts";
import type { usePostepUruchomienia } from "../../../api/postep.ts";
import { Format } from "../../../lib/format.ts";
import przyciski from "../../../ui/przyciski.module.css";
import styles from "./PostepUruchomienia.module.css";

type PostepUruchomieniaProps = {
  uruchomienie: Schematy["Uruchomienie"];
  postep: ReturnType<typeof usePostepUruchomienia>;
  zatrzymywanie: boolean;
  onZatrzymaj: () => void;
};

const FAZY: { faza: Schematy["Faza"]; etykieta: string }[] = [
  { faza: "konstrukcja", etykieta: "Konstrukcja" },
  { faza: "optymalizacja", etykieta: "Optymalizacja" },
  { faza: "walidacja", etykieta: "Walidacja" },
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
  const { historia, ostatni } = postep;
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
        aria-label="Postęp optymalizacji"
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

      {ostatni && (
        <dl className={styles.liczby}>
          <div>
            <dt>Najlepszy koszt</dt>
            <dd>{Format.liczba(ostatni.koszt)}</dd>
          </div>
          <div>
            <dt>Naruszenia (wg solvera)</dt>
            <dd className={ostatni.naruszenia > 0 ? styles.blad : undefined}>{ostatni.naruszenia}</dd>
          </div>
          <div>
            <dt>Iteracje</dt>
            <dd>
              {Format.liczba(ostatni.iteracja)} z {Format.liczba(ostatni.iteracji)}
            </dd>
          </div>
          <div>
            <dt>Czas</dt>
            <dd>{sekundy.format(ostatni.czas_ms / 1000)} s</dd>
          </div>
        </dl>
      )}
      {historia.length > 1 && <WykresKosztu historia={historia} />}
    </section>
  );
}

function WykresKosztu({ historia }: { historia: Schematy["Postep"][] }) {
  const koszty = historia.map((p) => p.koszt);
  const min = Math.min(...koszty);
  const max = Math.max(...koszty);
  const zakres = max - min || 1;
  const punkty = koszty
    .map((k, i) => `${(i / (koszty.length - 1)) * 100},${100 - ((k - min) / zakres) * 100}`)
    .join(" ");
  return (
    <figure className={styles.wykres}>
      <figcaption className={styles.cichy}>Spadek kosztu w trakcie optymalizacji</figcaption>
      <svg
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
        role="img"
        aria-label={`Koszt od ${Format.liczba(max)} do ${Format.liczba(min)}`}
      >
        <polyline
          points={punkty}
          fill="none"
          stroke="var(--kolor-marka)"
          strokeWidth="2"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </figure>
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
