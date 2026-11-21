import { Link, useLocation, useNavigate } from "react-router";
import styles from "./Okruszki.module.css";

type OkruszkiProps = {
  zestaw: { id: string; nazwa: string } | undefined;
  uruchomienie: string;
  biezacy?: string;
};

// Okruszki z makiety: „Pulpit / zestaw / uruchomienie / widok” i przycisk „Wstecz do poprzedniego widoku”.
export function Okruszki({ zestaw, uruchomienie, biezacy }: OkruszkiProps) {
  const navigate = useNavigate();
  const { key } = useLocation();
  // Zmiana wymiaru dodaje wpis w historii, więc „Wstecz” wraca do poprzedniego widoku.
  // Strona otwarta wprost z linku (pierwszy wpis, klucz „default”) wraca na pulpit.
  const wstecz = () => {
    if (key !== "default") {
      void navigate(-1);
    } else {
      void navigate("/");
    }
  };

  return (
    <div className={styles.okruszki}>
      <button
        type="button"
        aria-label="Wstecz do poprzedniego widoku"
        className={styles.wstecz}
        onClick={wstecz}
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
      <nav aria-label="Okruszki" className={styles.nawigacja}>
        <Link to="/">Pulpit</Link>
        <span aria-hidden="true">/</span>
        {zestaw && (
          <>
            <Link to="/">{zestaw.nazwa}</Link>
            <span aria-hidden="true">/</span>
          </>
        )}
        {biezacy ? (
          <>
            <span>{uruchomienie}</span>
            <span aria-hidden="true">/</span>
            <span aria-current="page" className={styles.biezacy}>
              {biezacy}
            </span>
          </>
        ) : (
          <span aria-current="page" className={styles.biezacy}>
            {uruchomienie}
          </span>
        )}
      </nav>
    </div>
  );
}
