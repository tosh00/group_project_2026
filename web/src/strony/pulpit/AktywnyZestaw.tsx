import { Link } from "react-router";
import type { Schematy } from "../../api/klient.ts";
import { Format } from "../../lib/format.ts";
import { Ikona, type NazwaIkony } from "../../ui/Ikona.tsx";
import przyciski from "../../ui/przyciski.module.css";
import type { StatusPlanu } from "./statusPlanu.ts";
import styles from "./AktywnyZestaw.module.css";

type AktywnyZestawProps = {
  zestaw: Schematy["Zestaw"];
  uruchomienie: Schematy["Uruchomienie"] | undefined;
  status: StatusPlanu;
  generowanie: { trwa: boolean; blad: string | null };
  onGeneruj: () => void;
  onPokazOstrzezenia: () => void;
};

export function AktywnyZestaw({
  zestaw,
  uruchomienie,
  status,
  generowanie,
  onGeneruj,
  onPokazOstrzezenia,
}: AktywnyZestawProps) {
  const p = zestaw.podsumowanie;
  const statystyki: { wartosc: string; etykieta: string; ikona: NazwaIkony }[] = [
    {
      wartosc: Format.liczba(p.zajecia),
      etykieta: Format.odmiana(p.zajecia, "zajęcie", "zajęcia", "zajęć"),
      ikona: "zajecia",
    },
    {
      wartosc: `${p.sale_do_planowania} z ${p.sale}`,
      etykieta: "sal do planowania",
      ikona: "sale",
    },
    {
      wartosc: Format.liczba(p.prowadzacy),
      etykieta: Format.odmiana(p.prowadzacy, "prowadzący", "prowadzących", "prowadzących"),
      ikona: "prowadzacy",
    },
    {
      wartosc: Format.liczba(p.roczniki),
      etykieta: Format.odmiana(p.roczniki, "rocznik", "roczniki", "roczników"),
      ikona: "roczniki",
    },
    {
      wartosc: Format.liczba(p.grupy),
      etykieta: Format.odmiana(p.grupy, "grupa", "grupy", "grup"),
      ikona: "grupy",
    },
    {
      wartosc: Format.liczba(p.studenci),
      etykieta: Format.odmiana(p.studenci, "student", "studentów", "studentów"),
      ikona: "studenci",
    },
  ];
  const liczbaOstrzezen = zestaw.ostrzezenia.length;

  return (
    <section aria-labelledby="aktywny-h" className={styles.karta}>
      <div className={styles.gora}>
        <div className={styles.nazwa}>
          <div className={styles.nadtytul}>Ostatnio wgrany zestaw</div>
          <h1 id="aktywny-h" className={styles.tytul}>
            {zestaw.nazwa}
          </h1>
          <div className={styles.cichy}>wgrano {Format.dataWzgledna(zestaw.utworzono)}</div>
        </div>
        <div className={styles.znaczniki}>
          <span className={styles.bezBledow}>
            <Ikona nazwa="ok" rozmiar={14} grubosc={2.4} />0 błędów w danych
          </span>
          {liczbaOstrzezen > 0 && (
            <button
              type="button"
              className={styles.ostrzezenia}
              aria-haspopup="dialog"
              onClick={onPokazOstrzezenia}
            >
              <Ikona nazwa="ostrzezenie" rozmiar={14} grubosc={2.2} />
              {Format.ilosc(liczbaOstrzezen, "ostrzeżenie", "ostrzeżenia", "ostrzeżeń")} · pokaż
            </button>
          )}
        </div>
      </div>
      <div className={styles.dol}>
        <div className={styles.zawartosc}>
          <div className={styles.etykietaSekcji}>Zawartość tego zestawu</div>
          <div className={styles.statystyki}>
            {statystyki.map((s) => (
              <div key={s.ikona} className={styles.statystyka}>
                <span className={styles.statystykaIkona}>
                  <Ikona nazwa={s.ikona} rozmiar={18} grubosc={1.9} />
                </span>
                <span className={styles.statystykaTekst}>
                  <span className={styles.wartosc}>{s.wartosc}</span>
                  <span className={styles.etykieta}>{s.etykieta}</span>
                </span>
              </div>
            ))}
          </div>
        </div>
        <div className={styles.plan}>
          <div className={styles.planEtykieta}>Plan zajęć</div>
          {uruchomienie ? (
            <>
              <div className={styles.planTytul}>{status.tekst}</div>
              <div className={styles.planOpis}>
                Wygenerowany {Format.dataWzgledna(uruchomienie.utworzono)}.
              </div>
              <Link
                to={`/uruchomienia/${uruchomienie.id}`}
                className={`${przyciski.glowny} ${styles.planPrzycisk}`}
              >
                Otwórz plan
                <Ikona nazwa="dalej" rozmiar={16} grubosc={2.2} />
              </Link>
            </>
          ) : (
            <>
              <div className={styles.planTytul}>Jeszcze nie wygenerowany</div>
              <div className={styles.planOpis}>
                Generowanie trwa od ułamka sekundy do około minuty. Ostrzeżenia nie blokują startu.
              </div>
              <button
                type="button"
                className={`${przyciski.glowny} ${styles.planPrzycisk}`}
                onClick={onGeneruj}
                disabled={generowanie.trwa}
              >
                <Ikona nazwa="start" rozmiar={16} />
                {generowanie.trwa ? "Uruchamiam…" : "Generuj plan"}
              </button>
            </>
          )}
          {generowanie.blad && (
            <p role="alert" className={styles.bladGenerowania}>
              {generowanie.blad}
            </p>
          )}
        </div>
      </div>
    </section>
  );
}
