import { useParams } from "react-router";
import { BladApi } from "../../api/klient.ts";
import { usePostepUruchomienia } from "../../api/postep.ts";
import {
  maPlan,
  useDaneZestawu,
  usePlan,
  useRaport,
  useUruchomienie,
  useZatrzymajUruchomienie,
} from "../../api/zapytania.ts";
import { Kalendarz } from "./Kalendarz.tsx";
import { BladUruchomienia, PostepUruchomienia } from "./komponenty/PostepUruchomienia.tsx";
import { useFiltryKalendarza } from "./useFiltryKalendarza.ts";
import styles from "./StronaKalendarz.module.css";

export function StronaKalendarz() {
  const { uruchomienieId = "" } = useParams();
  const uruchomienie = useUruchomienie(uruchomienieId);
  const u = uruchomienie.data;
  const trwa = u?.status === "w_kolejce" || u?.status === "trwa";
  const postep = usePostepUruchomienia(trwa ? uruchomienieId : undefined);
  const zatrzymaj = useZatrzymajUruchomienie();
  const dane = useDaneZestawu(maPlan(u) ? u?.zestaw_id : undefined);
  const plan = usePlan(u);
  const raport = useRaport(u);
  const domyslnyRocznik = dane.data
    ? ([...dane.data.roczniki].sort((a, b) => a.id.localeCompare(b.id))[0]?.id ?? "")
    : "";
  const { filtry, zmien, przejdz, wyczyscZawezajace } = useFiltryKalendarza(domyslnyRocznik);

  if (uruchomienie.isPending) {
    return <p className={styles.cichy}>Ładowanie planu…</p>;
  }
  if (uruchomienie.isError) {
    return (
      <p role="alert" className={styles.blad}>
        {uruchomienie.error instanceof BladApi && uruchomienie.error.kod === "nie_znaleziono"
          ? "Nie ma takiego uruchomienia."
          : `Nie udało się pobrać uruchomienia: ${uruchomienie.error.message}`}
      </p>
    );
  }

  if (trwa) {
    return (
      <div className={styles.strona}>
        <PostepUruchomienia
          uruchomienie={u!}
          postep={postep}
          zatrzymywanie={zatrzymaj.isPending}
          onZatrzymaj={() => zatrzymaj.mutate(uruchomienieId)}
        />
      </div>
    );
  }
  if (u!.status === "blad") {
    return (
      <div className={styles.strona}>
        <BladUruchomienia uruchomienie={u!} />
      </div>
    );
  }

  const bladDanych = dane.error ?? plan.error ?? raport.error;
  if (bladDanych) {
    return (
      <p role="alert" className={styles.blad}>
        Nie udało się pobrać planu: {bladDanych.message}
      </p>
    );
  }
  if (!dane.data || !plan.data || !raport.data) {
    return <p className={styles.cichy}>Ładowanie planu…</p>;
  }

  return (
    <Kalendarz
      uruchomienie={u!}
      dane={dane.data}
      plan={plan.data}
      raport={raport.data}
      filtry={filtry}
      onZmien={zmien}
      onPrzejdz={przejdz}
      onWyczyscFiltry={wyczyscZawezajace}
    />
  );
}
