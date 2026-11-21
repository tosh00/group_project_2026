import { Link, Outlet } from "react-router";
import styles from "./Uklad.module.css";

export function Uklad() {
  return (
    <div className={styles.uklad}>
      <header className={styles.naglowek}>
        <Link to="/" className={styles.nazwa}>
          Planer zajęć WFiIS
        </Link>
      </header>
      <main className={styles.tresc}>
        <Outlet />
      </main>
    </div>
  );
}
