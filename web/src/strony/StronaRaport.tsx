import { useParams } from "react-router";
import { Zaslepka } from "./Zaslepka.tsx";

export function StronaRaport() {
  const { uruchomienieId } = useParams();
  return (
    <Zaslepka tytul="Raport jakości">
      <p>Uruchomienie {uruchomienieId}: status, naruszenia i rozbicie kosztu na kryteria.</p>
    </Zaslepka>
  );
}
