import { Link, Outlet } from "react-router";
import { ZnacznikSerwera } from "./ZnacznikSerwera.tsx";
import styles from "./Uklad.module.css";

export function Uklad() {
  return (
    <div className={styles.uklad}>
      <header className={styles.naglowek}>
        <Link to="/" className={styles.nazwa}>
          Planer zajęć WFiIS
        </Link>
        <ZnacznikSerwera />
      </header>
      <main className={styles.tresc}>
        <Outlet />
      </main>
    </div>
  );
}
