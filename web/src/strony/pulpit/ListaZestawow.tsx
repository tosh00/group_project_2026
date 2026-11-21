import { Link } from "react-router";
import type { Schematy } from "../../api/klient.ts";
import { Format } from "../../lib/format.ts";
import { Ikona } from "../../ui/Ikona.tsx";
import przyciski from "../../ui/przyciski.module.css";
import { ChipStatusu } from "./ChipStatusu.tsx";
import { StatusyPlanow } from "./statusPlanu.ts";
import styles from "./ListaZestawow.module.css";

type ListaZestawowProps = {
  zestawy: Schematy["Zestaw"][];
  ostatnieUruchomienia: Map<string, Schematy["Uruchomienie"]>;
  walidator: Schematy["Zdrowie"]["walidator"] | undefined;
  generowanyId: string | undefined;
  onGeneruj: (zestawId: string) => void;
  onPokazOstrzezenia: (zestaw: Schematy["Zestaw"]) => void;
  onUsun: (zestaw: Schematy["Zestaw"]) => void;
};

export function ListaZestawow({
  zestawy,
  ostatnieUruchomienia,
  walidator,
  generowanyId,
  onGeneruj,
  onPokazOstrzezenia,
  onUsun,
}: ListaZestawowProps) {
  return (
    <section aria-labelledby="zestawy-h" className={styles.sekcja}>
      <div className={styles.naglowek}>
        <h2 id="zestawy-h" className={styles.tytul}>
          Twoje zestawy i plany
        </h2>
      </div>
      <ul className={styles.lista}>
        {zestawy.map((zestaw) => {
          const uruchomienie = ostatnieUruchomienia.get(zestaw.id);
          const status = StatusyPlanow.zUruchomienia(uruchomienie, walidator);
          const zajecia = zestaw.podsumowanie.zajecia;
          return (
            <li key={zestaw.id} className={styles.wiersz}>
              <span className={styles.ikona}>
                <Ikona nazwa="kalendarz" rozmiar={20} grubosc={1.8} />
              </span>
              <span className={styles.opis}>
                <span className={styles.nazwa}>{zestaw.nazwa}</span>
                <span className={styles.cichy}>
                  {Format.dataWzgledna(zestaw.utworzono)} ·{" "}
                  {Format.ilosc(zajecia, "zajęcie", "zajęcia", "zajęć")}
                </span>
              </span>
              {zestaw.ostrzezenia.length > 0 && (
                <button
                  type="button"
                  className={styles.ostrzezenia}
                  aria-haspopup="dialog"
                  onClick={() => onPokazOstrzezenia(zestaw)}
                >
                  <Ikona nazwa="ostrzezenie" rozmiar={14} grubosc={2.2} />
                  {Format.ilosc(zestaw.ostrzezenia.length, "uwaga", "uwagi", "uwag")} do danych
                </button>
              )}
              <span className={styles.status}>
                <ChipStatusu status={status} />
              </span>
              <span className={styles.wygenerowany}>
                {uruchomienie ? `plan z ${Format.dataWzgledna(uruchomienie.utworzono)}` : ""}
              </span>
              {uruchomienie ? (
                <Link to={`/uruchomienia/${uruchomienie.id}`} className={przyciski.miekki}>
                  Otwórz plan
                  <Ikona nazwa="dalej" rozmiar={14} grubosc={2.2} />
                </Link>
              ) : (
                <button
                  type="button"
                  className={przyciski.miekki}
                  onClick={() => onGeneruj(zestaw.id)}
                  disabled={generowanyId === zestaw.id}
                >
                  Generuj
                  <Ikona nazwa="dalej" rozmiar={14} grubosc={2.2} />
                </button>
              )}
              <button
                type="button"
                className={styles.usun}
                aria-label={`Usuń ${zestaw.nazwa}`}
                title="Usuń zestaw i jego plany"
                onClick={() => onUsun(zestaw)}
              >
                <Ikona nazwa="kosz" rozmiar={18} grubosc={1.9} />
              </button>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
