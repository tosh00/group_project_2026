import { useState, type DragEvent } from "react";
import { Ikona } from "../../ui/Ikona.tsx";
import przyciski from "../../ui/przyciski.module.css";
import styles from "./StrefaWgrywania.module.css";

type StrefaWgrywaniaProps = {
  // „pusty”: makieta 1a (brak zestawów, z krokami), „nowy”: makieta 1e (są wcześniejsze zestawy poniżej).
  wariant: "pusty" | "nowy";
  onWybierzPlik: () => void;
  onPlik: (plik: File) => void;
};

const kroki = [
  { numer: 1, tytul: "Wgraj plik", opis: "Sprawdzimy format i spójność danych." },
  { numer: 2, tytul: "Wygeneruj plan", opis: "Algorytm ułoży zajęcia w siatce tygodnia." },
  { numer: 3, tytul: "Sprawdź wynik", opis: "Kalendarz pokaże plan i ewentualne kolizje." },
];

export function StrefaWgrywania({ wariant, onWybierzPlik, onPlik }: StrefaWgrywaniaProps) {
  const [nadStrefa, setNadStrefa] = useState(false);
  const pusty = wariant === "pusty";

  const upusc = (zdarzenie: DragEvent) => {
    zdarzenie.preventDefault();
    setNadStrefa(false);
    const plik = zdarzenie.dataTransfer.files[0];
    if (plik) {
      onPlik(plik);
    }
  };

  return (
    <section aria-labelledby="start-h" className={pusty ? styles.sekcjaPusta : styles.sekcja}>
      <div className={styles.naglowek}>
        <h1 id="start-h" className={styles.tytul}>
          {pusty ? "Wgraj dane, żeby ułożyć plan" : "Wgraj nowy zestaw danych"}
        </h1>
        <p className={styles.podtytul}>
          {pusty
            ? "Nie masz jeszcze żadnego zestawu danych."
            : "Albo wróć do jednego z wcześniejszych planów poniżej."}
        </p>
      </div>
      <div
        className={nadStrefa ? `${styles.strefa} ${styles.strefaAktywna}` : styles.strefa}
        onDragOver={(zdarzenie) => {
          zdarzenie.preventDefault();
          setNadStrefa(true);
        }}
        onDragLeave={() => setNadStrefa(false)}
        onDrop={upusc}
      >
        <div className={styles.ikona}>
          <Ikona nazwa="wgraj" rozmiar={30} grubosc={1.8} />
        </div>
        <div className={styles.przeciagnij}>Przeciągnij tutaj plik .json</div>
        <div className={styles.cichy}>albo</div>
        <button type="button" className={`${przyciski.glowny} ${przyciski.duzy}`} onClick={onWybierzPlik}>
          Wybierz plik z dysku
        </button>
        <div className={styles.uwaga}>
          Tylko format .json z tabelami: zajęcia, sale, prowadzący, roczniki, grupy i studenci.
        </div>
      </div>
      {pusty && (
        <ol className={styles.kroki}>
          {kroki.map((krok) => (
            <li key={krok.numer} className={styles.krok}>
              <span className={styles.numer}>{krok.numer}</span>
              <span>
                <b className={styles.krokTytul}>{krok.tytul}</b>
                <span className={styles.krokOpis}>{krok.opis}</span>
              </span>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
