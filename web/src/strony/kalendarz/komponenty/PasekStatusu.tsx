import { useLocation, useNavigate } from "react-router";
import type { Schematy } from "../../../api/klient.ts";
import type { StatusPlanu } from "../model/typy.ts";
import styles from "./PasekStatusu.module.css";

type PasekStatusuProps = {
  uruchomienie: Schematy["Uruchomienie"];
  status: StatusPlanu;
  liczbaNaruszen: number;
  panelOtwarty: boolean;
  onPrzelaczPanel: () => void;
};

// Pasek nad siatką: „Wstecz” i status planu (otwiera panel). Bez danych technicznych (koszt, seed, iteracje).
export function PasekStatusu({
  uruchomienie,
  status,
  liczbaNaruszen,
  panelOtwarty,
  onPrzelaczPanel,
}: PasekStatusuProps) {
  const tekst =
    status === "naruszenia"
      ? `Plan z naruszeniami · ${liczbaNaruszen}`
      : status === "poprawny"
        ? "Plan poprawny · 0 naruszeń"
        : "Status niesprawdzony";

  return (
    <section aria-label="Status planu" className={styles.pasek}>
      <PrzyciskWstecz />
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
          title="Generowanie przerwano ręcznie. Pokazano najlepszy plan znaleziony do tej chwili."
        >
          Generowanie przerwane wcześniej
        </span>
      )}
    </section>
  );
}

// Zmiana widoku dodaje wpis w historii, więc „Wstecz” wraca do poprzedniego widoku.
// Strona otwarta wprost z linku (pierwszy wpis, klucz „default”) wraca na pulpit.
function PrzyciskWstecz() {
  const navigate = useNavigate();
  const { key } = useLocation();
  return (
    <button
      type="button"
      aria-label="Wstecz do poprzedniego widoku"
      className={styles.wstecz}
      onClick={() => void (key !== "default" ? navigate(-1) : navigate("/"))}
    >
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        <path d="M15 18l-6-6 6-6" />
      </svg>
    </button>
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
