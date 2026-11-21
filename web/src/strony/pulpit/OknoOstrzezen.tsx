import * as Dialog from "@radix-ui/react-dialog";
import type { Schematy } from "../../api/klient.ts";
import { Ikona } from "../../ui/Ikona.tsx";
import przyciski from "../../ui/przyciski.module.css";
import styles from "./OknoOstrzezen.module.css";

type OknoOstrzezenProps = {
  ostrzezenia: Schematy["Ostrzezenie"][];
  otwarte: boolean;
  onZmianaOtwarcia: (otwarte: boolean) => void;
  // Brak, gdy zestaw ma już plan: wtedy okno tylko informuje.
  onGeneruj?: () => void;
};

// Makieta „1b · Pulpit · modal ostrzeżeń”.
export function OknoOstrzezen({ ostrzezenia, otwarte, onZmianaOtwarcia, onGeneruj }: OknoOstrzezenProps) {
  return (
    <Dialog.Root open={otwarte} onOpenChange={onZmianaOtwarcia}>
      <Dialog.Portal>
        <Dialog.Overlay className={styles.zaslona} />
        <Dialog.Content className={styles.okno}>
          <div className={styles.gora}>
            <span className={styles.ikona}>
              <Ikona nazwa="ostrzezenie" rozmiar={20} />
            </span>
            <div className={styles.opis}>
              <Dialog.Title className={styles.tytul}>
                Ostrzeżenia w danych · {ostrzezenia.length}
              </Dialog.Title>
              <Dialog.Description className={styles.cichy}>
                Nie blokują generowania. Jeśli chcesz je usunąć, popraw plik i wgraj go ponownie.
              </Dialog.Description>
            </div>
            <Dialog.Close aria-label="Zamknij" className={styles.zamknij}>
              <Ikona nazwa="zamknij" rozmiar={18} />
            </Dialog.Close>
          </div>
          <ul className={styles.lista}>
            {ostrzezenia.map((o, i) => (
              <li key={`${o.kod}-${o.sala_id ?? i}`} className={styles.pozycja}>
                <div className={styles.pozycjaOpis}>{o.opis}</div>
                <code className={styles.miejsce}>{o.sala_id ? `sale[${o.sala_id}].pojemnosc` : o.kod}</code>
              </li>
            ))}
          </ul>
          <div className={styles.stopka}>
            <Dialog.Close className={przyciski.cichy}>Zamknij</Dialog.Close>
            {onGeneruj && (
              <button type="button" className={przyciski.glowny} onClick={onGeneruj}>
                Generuj plan mimo to
              </button>
            )}
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
