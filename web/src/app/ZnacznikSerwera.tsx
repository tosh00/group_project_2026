import { useZdrowie } from "../api/zapytania.ts";
import styles from "./ZnacznikSerwera.module.css";

type Wariant = "ok" | "neutralny" | "blad";

// Stan backendu w nagłówku. Walidator „atrapa” to stan neutralny, nie zielony: status planów nie jest sprawdzony.
export function ZnacznikSerwera() {
  const { data, isPending, isError } = useZdrowie();

  const [wariant, tekst]: [Wariant, string] = isPending
    ? ["neutralny", "Łączenie z serwerem…"]
    : isError || data.baza !== "ok"
      ? ["blad", "Serwer niedostępny"]
      : data.walidator === "atrapa"
        ? ["neutralny", "Walidator niepodłączony"]
        : ["ok", "Serwer działa"];

  return (
    <span className={`${styles.znacznik} ${styles[wariant]}`} role="status">
      {tekst}
    </span>
  );
}
