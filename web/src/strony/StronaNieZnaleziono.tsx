import { Link } from "react-router";
import { Zaslepka } from "./Zaslepka.tsx";

export function StronaNieZnaleziono() {
  return (
    <Zaslepka tytul="Nie ma takiej strony">
      <p>
        <Link to="/">Wróć do zestawów danych</Link>
      </p>
    </Zaslepka>
  );
}
