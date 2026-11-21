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
  aktywnyId: string | undefined;
  generowanyId: string | undefined;
  onWgrajNowy: () => void;
  onGeneruj: (zestawId: string) => void;
};

export function ListaZestawow({
  zestawy,
  ostatnieUruchomienia,
  walidator,
  aktywnyId,
  generowanyId,
  onWgrajNowy,
  onGeneruj,
}: ListaZestawowProps) {
  return (
    <section aria-labelledby="zestawy-h" className={styles.sekcja}>
      <div className={styles.naglowek}>
        <h2 id="zestawy-h" className={styles.tytul}>
          Twoje zestawy i plany
        </h2>
        <button type="button" className={przyciski.obrys} onClick={onWgrajNowy}>
          <Ikona nazwa="wgraj" grubosc={2.2} />
          Wgraj nowy zestaw
        </button>
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
                <span className={styles.nazwa}>
                  {zestaw.nazwa}
                  {zestaw.id === aktywnyId && <span className={styles.aktywny}>aktywny</span>}
                </span>
                <span className={styles.cichy}>
                  {Format.dataWzgledna(zestaw.utworzono)} ·{" "}
                  {Format.ilosc(zajecia, "zajęcie", "zajęcia", "zajęć")}
                </span>
              </span>
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
            </li>
          );
        })}
      </ul>
    </section>
  );
}
