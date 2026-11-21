import { useMemo, useState } from "react";
import type { Schematy } from "../../api/klient.ts";
import { Legenda } from "./komponenty/Legenda.tsx";
import { Okruszki } from "./komponenty/Okruszki.tsx";
import { PanelFiltrow } from "./komponenty/PanelFiltrow.tsx";
import { PanelStatusu } from "./komponenty/PanelStatusu.tsx";
import { PasekStatusu } from "./komponenty/PasekStatusu.tsx";
import { SiatkaTygodnia } from "./komponenty/SiatkaTygodnia.tsx";
import { Wyszukiwarka } from "./komponenty/Wyszukiwarka.tsx";
import { IndeksPlanu } from "./model/indeks.ts";
import { NaruszeniaPlanu, type Naruszenie } from "./model/naruszenia.ts";
import type { Filtry, Wymiar } from "./model/typy.ts";
import { WidokPlanu } from "./model/widok.ts";
import { Wyszukiwarka as ZrodloWynikow } from "./model/wyszukiwanie.ts";
import styles from "./Kalendarz.module.css";

type KalendarzProps = {
  uruchomienie: Schematy["Uruchomienie"];
  dane: Schematy["DaneWejsciowe"];
  plan: Schematy["Plan"];
  raport: Schematy["Raport"] | undefined;
  filtry: Filtry;
  onZmien: (zmiana: Partial<Filtry>) => void;
  onPrzejdz: (wymiar: Wymiar, id: string, zmiana?: Partial<Filtry>) => void;
  onWyczyscFiltry: () => void;
  zestaw: { id: string; nazwa: string } | undefined;
  etykietaUruchomienia: string;
};

// Ekran „2 · Kalendarz” z makiety: pasek statusu, filtry, siatka tygodnia, legenda i panel statusu.
export function Kalendarz({
  uruchomienie,
  dane,
  plan,
  raport,
  filtry,
  onZmien,
  onPrzejdz,
  onWyczyscFiltry,
  zestaw,
  etykietaUruchomienia,
}: KalendarzProps) {
  const indeks = useMemo(() => new IndeksPlanu(dane, plan), [dane, plan]);
  const naruszenia = useMemo(() => new NaruszeniaPlanu(raport), [raport]);
  const wyszukiwarka = useMemo(() => new ZrodloWynikow(indeks, naruszenia), [indeks, naruszenia]);
  const widok = new WidokPlanu(indeks, naruszenia, filtry);
  const opis = widok.opis();

  const [panelOtwarty, setPanelOtwarty] = useState(true);
  const [wskazane, setWskazane] = useState<string | null>(null);
  const [wybraneNaruszenie, setWybraneNaruszenie] = useState<Naruszenie | null>(null);

  const rocznikWidoku =
    filtry.wymiar === "rocznik"
      ? filtry.id
      : filtry.wymiar === "grupa"
        ? (indeks.grupy.get(filtry.id)?.rocznik ?? "")
        : (widok.zajeciaWidoku[0]?.rocznikId ?? indeks.roczniki[0]?.id ?? "");

  const obwiedzione = new Set<string>([
    ...(wskazane ? naruszenia.zajeciaPowiazane(wskazane) : []),
    ...(wybraneNaruszenie?.zajecia_id ?? []),
  ]);

  // Przełączenie wymiaru bez wskazania konkretnego planu: najbliższy sensowny wybór z bieżącego widoku.
  const domyslnyCel = (wymiar: Wymiar): string => {
    const pierwsze = widok.zajeciaWidoku[0];
    switch (wymiar) {
      case "rocznik":
        return rocznikWidoku || (indeks.roczniki[0]?.id ?? "");
      case "grupa":
        return (
          (indeks.grupyRocznika.get(rocznikWidoku) ?? []).find((g) => g.typ !== "rocznik")?.id ??
          [...indeks.grupy.keys()][0] ??
          ""
        );
      case "prowadzacy":
        return pierwsze?.prowadzacy[0] ?? [...indeks.prowadzacy.keys()][0] ?? "";
      case "sala":
        return pierwsze?.salaId ?? [...indeks.sale.values()].find((s) => s.dostepna_do_planowania)?.id ?? "";
    }
  };

  const przejdz = (wymiar: Wymiar, id: string) => {
    setWskazane(null);
    onPrzejdz(wymiar, id);
  };

  // Kliknięcie naruszenia: plan rocznika tych zajęć bez filtrów zawężających i obwódka wokół kart.
  const wybierzNaruszenie = (n: Naruszenie) => {
    if (wybraneNaruszenie?.indeks === n.indeks) {
      setWybraneNaruszenie(null);
      return;
    }

    setWybraneNaruszenie(n);
    const pierwsze = naruszenia.zajeciaNaruszenia(n, indeks)[0];
    if (pierwsze) {
      onPrzejdz("rocznik", pierwsze.rocznikId, { tydzien: 0, ukryteFormy: new Set() });
    }
  };

  return (
    <div className={styles.kalendarz}>
      <Okruszki zestaw={zestaw} uruchomienie={etykietaUruchomienia} biezacy={opis.okruszek} />
      <PasekStatusu
        uruchomienie={uruchomienie}
        status={naruszenia.status}
        liczbaNaruszen={naruszenia.lista.length}
        panelOtwarty={panelOtwarty}
        onPrzelaczPanel={() => setPanelOtwarty((o) => !o)}
      />
      <PanelFiltrow
        filtry={filtry}
        widokStudentow={widok.widokStudentow}
        onWymiar={(wymiar) => wymiar !== filtry.wymiar && przejdz(wymiar, domyslnyCel(wymiar))}
        onZmien={onZmien}
        wyszukiwarka={
          <Wyszukiwarka
            zrodlo={wyszukiwarka}
            filtry={filtry}
            opis={opis}
            naruszenWidoku={naruszenia.liczbaDla(widok.zajeciaWidoku)}
            onWybierz={(cel) => przejdz(cel.wymiar, cel.id)}
          />
        }
      />
      <div className={styles.uklad}>
        <SiatkaTygodnia
          widok={widok}
          opis={opis}
          indeks={indeks}
          naruszenia={naruszenia}
          filtry={filtry}
          wskazane={wskazane}
          obwiedzione={obwiedzione}
          onWskazanie={setWskazane}
          onPrzejdz={przejdz}
          onWyczyscFiltry={onWyczyscFiltry}
        />
        <div className={styles.bok}>
          <Legenda
            indeks={indeks}
            widok={widok}
            filtry={filtry}
            rocznikWidoku={rocznikWidoku}
            onPrzejdz={przejdz}
          />
          {panelOtwarty && (
            <PanelStatusu
              naruszenia={naruszenia}
              indeks={indeks}
              wybrane={wybraneNaruszenie?.indeks ?? null}
              onWybierz={wybierzNaruszenie}
              onZamknij={() => setPanelOtwarty(false)}
            />
          )}
        </div>
      </div>
    </div>
  );
}
