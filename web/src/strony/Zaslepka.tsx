import type { ReactNode } from "react";
import styles from "./Zaslepka.module.css";

type ZaslepkaProps = {
  tytul: string;
  children?: ReactNode;
};

// Tymczasowa treść strony, dopóki ekran z designu nie jest zrobiony.
export function Zaslepka({ tytul, children }: ZaslepkaProps) {
  return (
    <section className={styles.zaslepka}>
      <h1 className={styles.tytul}>{tytul}</h1>
      {children}
    </section>
  );
}
