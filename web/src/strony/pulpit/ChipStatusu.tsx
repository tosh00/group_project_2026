import przyciski from "../../ui/przyciski.module.css";
import type { StatusPlanu } from "./statusPlanu.ts";

export function ChipStatusu({ status }: { status: StatusPlanu }) {
  return (
    <span className={przyciski[`chip_${status.rodzaj}`]}>
      <span aria-hidden="true">{status.znak}</span>
      {status.tekst}
    </span>
  );
}
