import { useRef, useState } from "react";
import { useNavigate } from "react-router";
import { BladApi } from "../../api/klient.ts";
import { useUruchomienia, useUruchomSolver, useZdrowie, useZestawy } from "../../api/zapytania.ts";
import { AktywnyZestaw } from "./AktywnyZestaw.tsx";
import { BledyWgrywania } from "./BledyWgrywania.tsx";
import { ListaZestawow } from "./ListaZestawow.tsx";
import { OknoOstrzezen } from "./OknoOstrzezen.tsx";
import { PostepWgrywania } from "./PostepWgrywania.tsx";
import { StatusyPlanow } from "./statusPlanu.ts";
import { StrefaWgrywania } from "./StrefaWgrywania.tsx";
import { useWgrywaniePliku } from "./useWgrywaniePliku.ts";
import styles from "./StronaPulpit.module.css";

// Ekran „1 · Pulpit” z makiety i jego stany: brak danych (1a), modal ostrzeżeń (1b), wgrywanie (1c),
// błąd pliku (1d), nowy zestaw przy istniejących (1e), zestaw wgrany (1).
export function StronaPulpit() {
  const navigate = useNavigate();
  const zestawy = useZestawy();
  const uruchomienia = useUruchomienia();
  const zdrowie = useZdrowie();
  const uruchomSolver = useUruchomSolver();
  const { stan: wgrywanie, wgraj, anuluj } = useWgrywaniePliku();

  const wyborPliku = useRef<HTMLInputElement>(null);
  const [nowyZestaw, setNowyZestaw] = useState(false);
  const [ostrzezeniaOtwarte, setOstrzezeniaOtwarte] = useState(false);
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
  const aktywny = lista[0];
  const ostatnie = StatusyPlanow.ostatnieWgZestawu(uruchomienia.data ?? []);
  const walidator = zdrowie.data?.walidator;

  const wybierzPlik = () => wyborPliku.current?.click();
  const wgrajPlik = (plik: File) => {
    setNowyZestaw(false);
    setBladGenerowania(null);
    void wgraj(plik);
  };
  const generuj = async (zestawId: string) => {
    setOstrzezeniaOtwarte(false);
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

  const pokazStrefe = wgrywanie.etap === "bezczynny" && (!aktywny || nowyZestaw);
  const pokazAktywny = wgrywanie.etap === "bezczynny" && aktywny && !nowyZestaw;
  const uruchomienieAktywnego = aktywny ? ostatnie.get(aktywny.id) : undefined;

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

      {pokazStrefe && (
        <StrefaWgrywania
          wariant={aktywny ? "nowy" : "pusty"}
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
      {pokazAktywny && (
        <>
          <AktywnyZestaw
            zestaw={aktywny}
            uruchomienie={uruchomienieAktywnego}
            status={StatusyPlanow.zUruchomienia(uruchomienieAktywnego, walidator)}
            generowanie={{ trwa: generowanyId === aktywny.id, blad: bladGenerowania }}
            onGeneruj={() => void generuj(aktywny.id)}
            onPokazOstrzezenia={() => setOstrzezeniaOtwarte(true)}
          />
          <OknoOstrzezen
            ostrzezenia={aktywny.ostrzezenia}
            otwarte={ostrzezeniaOtwarte}
            onZmianaOtwarcia={setOstrzezeniaOtwarte}
            onGeneruj={uruchomienieAktywnego ? undefined : () => void generuj(aktywny.id)}
          />
        </>
      )}

      {lista.length > 0 && (
        <ListaZestawow
          zestawy={lista}
          ostatnieUruchomienia={ostatnie}
          walidator={walidator}
          aktywnyId={aktywny?.id}
          generowanyId={generowanyId}
          onWgrajNowy={() => setNowyZestaw(true)}
          onGeneruj={(zestawId) => void generuj(zestawId)}
        />
      )}
    </>
  );
}
