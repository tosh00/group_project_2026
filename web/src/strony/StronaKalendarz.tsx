import { useParams } from "react-router";
import { Zaslepka } from "./Zaslepka.tsx";

export function StronaKalendarz() {
  const { uruchomienieId } = useParams();
  return (
    <Zaslepka tytul="Kalendarz planu">
      <p>Uruchomienie {uruchomienieId}: siatka 5 dni × 12 slotów z filtrami (design: „2 · Kalendarz”).</p>
    </Zaslepka>
  );
}
