import { Link } from "react-router";
import type { Schematy } from "../../../api/klient.ts";
import { Format } from "../../../lib/format.ts";
import type { StatusPlanu } from "../model/typy.ts";
import styles from "./PasekStatusu.module.css";

type PasekStatusuProps = {
  uruchomienie: Schematy["Uruchomienie"];
  status: StatusPlanu;
  liczbaNaruszen: number;
  panelOtwarty: boolean;
  onPrzelaczPanel: () => void;
};

const sekundy = new Intl.NumberFormat("pl-PL", { maximumFractionDigits: 1, minimumFractionDigits: 1 });
const dataICzas = new Intl.DateTimeFormat("pl-PL", {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

class CzasUruchomienia {
  static sekundy(uruchomienie: Schematy["Uruchomienie"]): string | null {
    if (!uruchomienie.rozpoczeto || !uruchomienie.zakonczono) {
      return null;
    }

    const ms = new Date(uruchomienie.zakonczono).getTime() - new Date(uruchomienie.rozpoczeto).getTime();
    return `${sekundy.format(ms / 1000)} s`;
  }
}

// Pasek nad siatką z makiety: status (otwiera panel), koszt, seed, data i czas trwania, raport jakości.
export function PasekStatusu({
  uruchomienie,
  status,
  liczbaNaruszen,
  panelOtwarty,
  onPrzelaczPanel,
}: PasekStatusuProps) {
  const czas = CzasUruchomienia.sekundy(uruchomienie);
  const tekst =
    status === "naruszenia"
      ? `Plan z naruszeniami · ${liczbaNaruszen}`
      : status === "poprawny"
        ? "Plan poprawny · 0 naruszeń"
        : "Status niesprawdzony";

  return (
    <section aria-label="Status planu" className={styles.pasek}>
      <button
        type="button"
        className={styles[status]}
        onClick={onPrzelaczPanel}
        aria-expanded={panelOtwarty}
        aria-controls="panel-statusu"
      >
        <IkonaStatusu status={status} />
        {tekst}
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
          <path d="M9 6l6 6-6 6" />
        </svg>
      </button>
      {uruchomienie.zatrzymane && (
        <span
          className={styles.zatrzymany}
          title="Optymalizację przerwano ręcznie. Pokazano najlepszy dotychczasowy plan."
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <rect x="6" y="6" width="12" height="12" rx="1.5" />
          </svg>
          {czas ? `Zatrzymany po ${czas}` : "Zatrzymany"}
        </span>
      )}
      <div className={styles.koszt}>
        <span className={styles.cichy}>Koszt planu</span>
        <span className={styles.kosztWartosc}>
          {uruchomienie.koszt === null ? "–" : Format.liczba(uruchomienie.koszt ?? 0)}
        </span>
      </div>
      <span className={styles.separator} aria-hidden="true" />
      <span className={styles.cichy}>
        Seed <span className={styles.seed}>{uruchomienie.seed}</span>
      </span>
      <span className={styles.cichy}>
        {dataICzas.format(new Date(uruchomienie.utworzono))}
        {czas && ` · ${uruchomienie.zatrzymane ? `zatrzymano po ${czas}` : czas}`} ·{" "}
        {Format.ilosc(uruchomienie.iteracje, "iteracja", "iteracje", "iteracji")}
      </span>
      <div className={styles.akcje}>
        <Link to={`/uruchomienia/${uruchomienie.id}/raport`} className={styles.raport}>
          Raport jakości
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
            <path d="M5 12h14M13 6l6 6-6 6" />
          </svg>
        </Link>
      </div>
    </section>
  );
}

function IkonaStatusu({ status }: { status: StatusPlanu }) {
  const wspolne = {
    width: 16,
    height: 16,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true,
  };
  if (status === "naruszenia") {
    return (
      <svg {...wspolne} strokeWidth="2.2">
        <path d="M10.3 3.9L1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z" />
        <path d="M12 9v4M12 17h.01" />
      </svg>
    );
  }
  if (status === "poprawny") {
    return (
      <svg {...wspolne} strokeWidth="2.6">
        <path d="M5 12l5 5L20 7" />
      </svg>
    );
  }
  return (
    <svg {...wspolne} strokeWidth="2.2">
      <circle cx="12" cy="12" r="9" />
      <path d="M12 11v5M12 7.5h.01" />
    </svg>
  );
}
