import { useZdrowie } from "../api/zapytania.ts";
import styles from "./ZnacznikSerwera.module.css";

// W makiecie nagłówek ma tylko logo. Znacznik pojawia się wyłącznie wtedy, gdy backend nie odpowiada,
// bo bez serwera żaden ekran nie zadziała. Stan walidatora pokazują statusy planów, nie nagłówek.
export function ZnacznikSerwera() {
  const { data, isError } = useZdrowie();

  if (!isError && data?.baza !== "blad") {
    return null;
  }

  return (
    <span className={styles.znacznik} role="status">
      Serwer niedostępny
    </span>
  );
}
