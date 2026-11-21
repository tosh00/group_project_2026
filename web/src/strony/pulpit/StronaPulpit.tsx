import { useRef, useState } from "react";
import { useNavigate } from "react-router";
import { BladApi, type Schematy } from "../../api/klient.ts";
import { useUruchomienia, useUruchomSolver, useZdrowie, useZestawy } from "../../api/zapytania.ts";
import { BledyWgrywania } from "./BledyWgrywania.tsx";
import { ListaZestawow } from "./ListaZestawow.tsx";
import { OknoOstrzezen } from "./OknoOstrzezen.tsx";
import { PostepWgrywania } from "./PostepWgrywania.tsx";
import { StatusyPlanow } from "./statusPlanu.ts";
import { StrefaWgrywania } from "./StrefaWgrywania.tsx";
import { useWgrywaniePliku } from "./useWgrywaniePliku.ts";
import styles from "./StronaPulpit.module.css";

// Pulpit: u góry zawsze pole do wgrania kolejnego pliku (makieta 1a/1e), pod nim lista zestawów i planów.
// Stany wgrywania (1c) i błędu pliku (1d) zastępują pole, dopóki trwają.
export function StronaPulpit() {
  const navigate = useNavigate();
  const zestawy = useZestawy();
  const uruchomienia = useUruchomienia();
  const zdrowie = useZdrowie();
  const uruchomSolver = useUruchomSolver();
  const { stan: wgrywanie, wgraj, anuluj } = useWgrywaniePliku();

  const wyborPliku = useRef<HTMLInputElement>(null);
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
  const ostatnie = StatusyPlanow.ostatnieWgZestawu(uruchomienia.data ?? []);
  const walidator = zdrowie.data?.walidator;

  const wybierzPlik = () => wyborPliku.current?.click();
  const wgrajPlik = (plik: File) => {
    setBladGenerowania(null);
    void wgraj(plik);
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
            wgrajPlik(plik);
          }
        }}
      />

      {wgrywanie.etap === "bezczynny" && (
        <StrefaWgrywania
          wariant={lista.length ? "nowy" : "pusty"}
          onWybierzPlik={wybierzPlik}
          onPlik={wgrajPlik}
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
