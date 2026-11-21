import { useMemo, useState, type CSSProperties } from "react";
import * as Popover from "@radix-ui/react-popover";
import { Command } from "cmdk";
import { Kolory } from "../model/kolory.ts";
import type { Filtry } from "../model/typy.ts";
import type { OpisWidoku } from "../model/widok.ts";
import { Wyszukiwarka as ZrodloWynikow, type Cel } from "../model/wyszukiwanie.ts";
import styles from "./Wyszukiwarka.module.css";

type WyszukiwarkaProps = {
  zrodlo: ZrodloWynikow;
  filtry: Filtry;
  opis: OpisWidoku;
  naruszenWidoku: number;
  onWybierz: (cel: Cel) => void;
};

// Wybór „czyj plan” z makiety: przycisk z bieżącym widokiem i lista z wyszukiwaniem.
export function Wyszukiwarka({ zrodlo, filtry, opis, naruszenWidoku, onWybierz }: WyszukiwarkaProps) {
  const [otwarta, setOtwarta] = useState(false);
  const [fraza, setFraza] = useState("");
  const sekcje = useMemo(() => zrodlo.szukaj(fraza, opis.rocznik), [zrodlo, fraza, opis.rocznik]);

  const wybierz = (cel: Cel) => {
    setOtwarta(false);
    setFraza("");
    onWybierz(cel);
  };

  return (
    <Popover.Root open={otwarta} onOpenChange={setOtwarta}>
      <Popover.Trigger
        className={otwarta ? `${styles.wybor} ${styles.wyborOtwarty}` : styles.wybor}
        aria-haspopup="listbox"
      >
        <IkonaLupy rozmiar={15} />
        <span className={styles.wyborNazwa}>{opis.nazwa}</span>
        <span className={styles.wyborPodpis}>{opis.podpis}</span>
        {naruszenWidoku > 0 ? (
          <span className={styles.licznikNaruszen}>
            <IkonaUwagi />({naruszenWidoku})
          </span>
        ) : (
          <span className={styles.licznikZero}>(0)</span>
        )}
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <path d="M6 9l6 6 6-6" />
        </svg>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content align="start" sideOffset={6} className={styles.okno}>
          <Command shouldFilter={false} label="Wyszukiwarka planów">
            <div className={styles.pole}>
              <IkonaLupy rozmiar={16} />
              <Command.Input
                value={fraza}
                onValueChange={setFraza}
                placeholder="Szukaj: P07, A28, Analiza…"
                aria-label="Szukaj rocznika, grupy, prowadzącego, sali lub przedmiotu"
                className={styles.wejscie}
                autoFocus
              />
              {fraza && (
                <button
                  type="button"
                  aria-label="Wyczyść wyszukiwanie"
                  className={styles.wyczysc}
                  onClick={() => setFraza("")}
                >
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2.2"
                    strokeLinecap="round"
                    aria-hidden="true"
                  >
                    <path d="M6 6l12 12M18 6L6 18" />
                  </svg>
                </button>
              )}
            </div>
            <Command.List className={styles.lista} aria-label="Wyniki wyszukiwania">
              {sekcje.length === 0 && (
                <div className={styles.brak}>
                  <div className={styles.brakTytul}>Brak wyników dla „{fraza}”</div>
                  Szukaj po kodzie (P07, A281, INF1-R1-G2) albo nazwie przedmiotu.
                </div>
              )}
              {sekcje.map((sekcja) => (
                <Command.Group
                  key={sekcja.tytul}
                  heading={
                    <span className={styles.sekcja}>
                      <span>{sekcja.tytul}</span>
                      <span className={styles.notka}>{sekcja.notka}</span>
                    </span>
                  }
                >
                  {sekcja.wyniki.map((w) => {
                    const biezacy =
                      w.cel.wymiar === filtry.wymiar &&
                      w.cel.id === filtry.id &&
                      !w.klucz.startsWith("przedmiot:");
                    const [przed, trafienie, po] = ZrodloWynikow.podswietl(w.etykieta, fraza);
                    return (
                      <Command.Item
                        key={w.klucz}
                        value={w.klucz}
                        onSelect={() => wybierz(w.cel)}
                        className={`${styles.wynik} ${w.wciecie ? styles.wciety : ""} ${biezacy ? styles.biezacy : ""}`}
                      >
                        <span
                          className={w.kolor ? styles.probka : styles.probkaPusta}
                          style={
                            w.kolor
                              ? ({
                                  "--probka-tlo": Kolory.rgba(w.kolor, 0.2),
                                  "--probka-ramka": Kolory.rgba(w.kolor, 0.5),
                                  "--probka-pasek": w.kolor,
                                } as CSSProperties)
                              : undefined
                          }
                          aria-hidden="true"
                        />
                        <span className={styles.wynikTekst}>
                          <span className={styles.wynikEtykieta}>
                            {przed}
                            {trafienie && <mark className={styles.trafienie}>{trafienie}</mark>}
                            {po}
                          </span>
                          <span className={styles.wynikOpis}>{w.opis}</span>
                        </span>
                        {w.naruszen > 0 && (
                          <span className={styles.wynikNaruszenia}>
                            <IkonaUwagi />
                            {w.naruszen}
                          </span>
                        )}
                        {biezacy && <span className={styles.teraz}>teraz</span>}
                      </Command.Item>
                    );
                  })}
                </Command.Group>
              ))}
            </Command.List>
          </Command>
          <div className={styles.stopka}>
            <span>↑ ↓ wybór</span>
            <span>Enter otwiera plan</span>
            <span>Esc zamyka</span>
          </div>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}

function IkonaLupy({ rozmiar }: { rozmiar: number }) {
  return (
    <svg
      width={rozmiar}
      height={rozmiar}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="11" cy="11" r="7" />
      <path d="M20 20l-3.5-3.5" />
    </svg>
  );
}

function IkonaUwagi() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M10.3 3.9L1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z" />
      <path d="M12 9v4" />
    </svg>
  );
}
