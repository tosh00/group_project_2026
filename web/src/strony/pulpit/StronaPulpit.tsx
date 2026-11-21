import { useRef, useState } from "react";
import { useNavigate } from "react-router";
import { BladApi, type Schematy } from "../../api/klient.ts";
import { useUruchomienia, useUruchomSolver, useZdrowie, useZestawy } from "../../api/zapytania.ts";
import { AktywnyZestaw } from "./AktywnyZestaw.tsx";
import { BledyWgrywania } from "./BledyWgrywania.tsx";
import { ListaZestawow } from "./ListaZestawow.tsx";
import { OknoOstrzezen } from "./OknoOstrzezen.tsx";
import { OknoUsuwania } from "./OknoUsuwania.tsx";
import { PostepWgrywania } from "./PostepWgrywania.tsx";
import { StatusyPlanow } from "./statusPlanu.ts";
import { StrefaWgrywania } from "./StrefaWgrywania.tsx";
import { useWgrywaniePliku } from "./useWgrywaniePliku.ts";
import styles from "./StronaPulpit.module.css";

// Pulpit: u góry pole do wgrania pliku (makieta 1a/1e), pod nim lista zestawów i planów.
// W miejscu pola: postęp wgrywania (1c), błędy pliku (1d), a po udanym wgraniu karta nowego zestawu
// z kafelkami policzonymi przez serwer z pliku (1), dopóki użytkownik nie wybierze „Wgraj inny plik”.
export function StronaPulpit() {
  const navigate = useNavigate();
  const zestawy = useZestawy();
  const uruchomienia = useUruchomienia();
  const zdrowie = useZdrowie();
  const uruchomSolver = useUruchomSolver();
  const { stan: wgrywanie, wgraj, anuluj } = useWgrywaniePliku();

  const wyborPliku = useRef<HTMLInputElement>(null);
  const [wgranyId, setWgranyId] = useState<string | null>(null);
  const [usuwany, setUsuwany] = useState<Schematy["Zestaw"] | null>(null);
  const [ostrzezeniaZestawu, setOstrzezeniaZestawu] = useState<Schematy["Zestaw"] | null>(null);
  const [generowanyId, setGenerowanyId] = useState<string>();
  const [bladGenerowania, setBladGenerowania] = useState<string | null>(null);

  if (zestawy.isPending) {
    return <p className={styles.ladowanie}>Ładowanie zestawów…</p>;
  }
  if (zestawy.isError) {
    return (
      <p role="alert" className={styles.blad}>
        Nie udało się pobrać zestawów: {zestawy.error.message}
      </p>
    );
  }

  const lista = zestawy.data;
  // Dane z listy (odświeżanej po wgraniu), żeby karta miała ten sam stan co wiersz zestawu.
  const wgrany = lista.find((z) => z.id === wgranyId);
  const ostatnie = StatusyPlanow.ostatnieWgZestawu(uruchomienia.data ?? []);
  const walidator = zdrowie.data?.walidator;

  const wybierzPlik = () => wyborPliku.current?.click();
  const wgrajPlik = async (plik: File) => {
    setBladGenerowania(null);
    setWgranyId(null);
    const zestaw = await wgraj(plik);
    setWgranyId(zestaw?.id ?? null);
  };
  const generuj = async (zestawId: string) => {
    setOstrzezeniaZestawu(null);
    setGenerowanyId(zestawId);
    setBladGenerowania(null);
    try {
      const uruchomienie = await uruchomSolver.mutateAsync({ zestawId });
      void navigate(`/uruchomienia/${uruchomienie.id}`);
    } catch (blad) {
      setBladGenerowania(blad instanceof BladApi ? blad.message : "Brak połączenia z serwerem.");
    } finally {
      setGenerowanyId(undefined);
    }
  };

  return (
    <>
      <input
        ref={wyborPliku}
        type="file"
        accept=".json,application/json"
        hidden
        onChange={(zdarzenie) => {
          const plik = zdarzenie.target.files?.[0];
          zdarzenie.target.value = "";
          if (plik) {
            void wgrajPlik(plik);
          }
        }}
      />

      {wgrywanie.etap === "bezczynny" && wgrany && (
        <AktywnyZestaw
          zestaw={wgrany}
          uruchomienie={ostatnie.get(wgrany.id)}
          status={StatusyPlanow.zUruchomienia(ostatnie.get(wgrany.id), walidator)}
          generowanie={{ trwa: generowanyId === wgrany.id, blad: null }}
          onGeneruj={() => void generuj(wgrany.id)}
          onPokazOstrzezenia={() => setOstrzezeniaZestawu(wgrany)}
          onWgrajKolejny={() => setWgranyId(null)}
        />
      )}
      {wgrywanie.etap === "bezczynny" && !wgrany && (
        <StrefaWgrywania
          wariant={lista.length ? "nowy" : "pusty"}
          onWybierzPlik={wybierzPlik}
          onPlik={(plik) => void wgrajPlik(plik)}
        />
      )}
      {wgrywanie.etap === "wgrywanie" && (
        <PostepWgrywania plik={wgrywanie.plik} krok={wgrywanie.krok} onAnuluj={anuluj} />
      )}
      {wgrywanie.etap === "blad" && (
        <BledyWgrywania plik={wgrywanie.plik} bledy={wgrywanie.bledy} onWybierzPlik={wybierzPlik} />
      )}
      {bladGenerowania && (
        <p role="alert" className={styles.blad}>
          {bladGenerowania}
        </p>
      )}

      {lista.length > 0 && (
        <ListaZestawow
          zestawy={lista}
          ostatnieUruchomienia={ostatnie}
          walidator={walidator}
          generowanyId={generowanyId}
          onGeneruj={(zestawId) => void generuj(zestawId)}
          onPokazOstrzezenia={setOstrzezeniaZestawu}
          onUsun={setUsuwany}
        />
      )}
      {usuwany && (
        <OknoUsuwania
          zestaw={usuwany}
          liczbaPlanow={(uruchomienia.data ?? []).filter((u) => u.zestaw_id === usuwany.id).length}
          onZamknij={() => setUsuwany(null)}
        />
      )}
      {ostrzezeniaZestawu && (
        <OknoOstrzezen
          ostrzezenia={ostrzezeniaZestawu.ostrzezenia}
          otwarte
          onZmianaOtwarcia={(otwarte) => !otwarte && setOstrzezeniaZestawu(null)}
          onGeneruj={
            ostatnie.has(ostrzezeniaZestawu.id) ? undefined : () => void generuj(ostrzezeniaZestawu.id)
          }
        />
      )}
    </>
  );
}
