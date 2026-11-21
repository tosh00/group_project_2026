import { Link, Outlet, useMatches } from "react-router";
import logo from "../assets/logo-wfiis.png";
import { ZnacznikSerwera } from "./ZnacznikSerwera.tsx";
import styles from "./Uklad.module.css";

export function Uklad() {
  const szeroki = useMatches().some((m) => (m.handle as { szeroki?: boolean } | undefined)?.szeroki);
  return (
    <div className={szeroki ? `${styles.uklad} ${styles.szeroki}` : styles.uklad}>
      <header className={styles.naglowek}>
        <div className={styles.naglowekWnetrze}>
          <Link to="/" aria-label="Pulpit" className={styles.logo}>
            <img src={logo} alt="Wydział Fizyki i Informatyki Stosowanej, Uniwersytet Łódzki" height={44} />
          </Link>
          <ZnacznikSerwera />
        </div>
      </header>
      <main className={styles.tresc}>
        <Outlet />
      </main>
    </div>
  );
}
