import { useParams } from "react-router";
import { Zaslepka } from "./Zaslepka.tsx";

export function StronaZestaw() {
  const { zestawId } = useParams();
  return (
    <Zaslepka tytul="Zestaw danych">
      <p>Zestaw {zestawId}: podsumowanie, wagi, generowanie i lista uruchomień.</p>
    </Zaslepka>
  );
}
