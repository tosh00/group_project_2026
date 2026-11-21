import * as Dialog from "@radix-ui/react-dialog";
import type { Schematy } from "../../api/klient.ts";
import { BladApi } from "../../api/klient.ts";
import { useUsunZestaw } from "../../api/zapytania.ts";
import { Format } from "../../lib/format.ts";
import { Ikona } from "../../ui/Ikona.tsx";
import przyciski from "../../ui/przyciski.module.css";
import styles from "./OknoUsuwania.module.css";

type OknoUsuwaniaProps = {
  zestaw: Schematy["Zestaw"];
  liczbaPlanow: number;
  onZamknij: () => void;
};

// Potwierdzenie usunięcia zestawu z planami. Usuwanie jest nieodwracalne, więc wymaga drugiego kliknięcia.
export function OknoUsuwania({ zestaw, liczbaPlanow, onZamknij }: OknoUsuwaniaProps) {
  const usun = useUsunZestaw();
  const blad =
    usun.error instanceof BladApi ? usun.error.message : usun.error ? "Brak połączenia z serwerem." : null;

  return (
    <Dialog.Root open onOpenChange={(otwarte) => !otwarte && !usun.isPending && onZamknij()}>
      <Dialog.Portal>
        <Dialog.Overlay className={styles.zaslona} />
        <Dialog.Content className={styles.okno}>
          <div className={styles.gora}>
            <span className={styles.ikona}>
              <Ikona nazwa="kosz" rozmiar={20} />
            </span>
            <div>
              <Dialog.Title className={styles.tytul}>Usunąć „{zestaw.nazwa}”?</Dialog.Title>
              <Dialog.Description className={styles.opis}>
                {liczbaPlanow > 0
                  ? `Razem z danymi zostanie usunięte ${Format.ilosc(liczbaPlanow, "wygenerowany plan", "wygenerowane plany", "wygenerowanych planów")}. `
                  : "Zostaną usunięte wgrane dane. "}
                Tej operacji nie da się cofnąć.
              </Dialog.Description>
            </div>
          </div>
          {blad && (
            <p role="alert" className={styles.blad}>
              {blad}
            </p>
          )}
          <div className={styles.stopka}>
            <Dialog.Close className={przyciski.cichy} disabled={usun.isPending}>
              Anuluj
            </Dialog.Close>
            <button
              type="button"
              className={styles.usun}
              disabled={usun.isPending}
              onClick={() => usun.mutate(zestaw.id, { onSuccess: onZamknij })}
            >
              {usun.isPending ? "Usuwam…" : "Usuń"}
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
