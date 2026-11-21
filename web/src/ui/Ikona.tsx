// Ikony z makiety (obrys 24×24, kolor z `currentColor`). Dekoracyjne, więc ukryte przed czytnikami ekranu.
const sciezki = {
  wgraj: ["M12 16V4M6 10l6-6 6 6M4 20h16"],
  plik: ["M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z", "M14 3v5h5"],
  ok: ["M5 12l5 5L20 7"],
  blad: ["M15 9l-6 6M9 9l6 6", "M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18z"],
  ostrzezenie: [
    "M10.3 3.9L1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z",
    "M12 9v4M12 17h.01",
  ],
  zamknij: ["M6 6l12 12M18 6L6 18"],
  dalej: ["M9 6l6 6-6 6"],
  kalendarz: [
    "M5 4h14a2 2 0 0 1 2 2v13a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z",
    "M3 9h18M8 2v4M16 2v4",
  ],
  zajecia: ["M4 5a2 2 0 0 1 2-2h12v16H6a2 2 0 0 0-2 2zM4 21V5M8 7h6"],
  sale: ["M4 21V4h11v17M15 6h5v15M2 21h20M11 12h.01"],
  prowadzacy: ["M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM4 21a8 8 0 0 1 16 0"],
  roczniki: ["M12 3l9 5-9 5-9-5zM3 13l9 5 9-5"],
  grupy: [
    "M9 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7zM2 20a7 7 0 0 1 14 0M16 4.5a3.5 3.5 0 0 1 0 6.5M18 13.5a7 7 0 0 1 4 6.5",
  ],
  studenci: ["M2 9l10-5 10 5-10 5zM6 11v5c3 2 9 2 12 0v-5M22 9v6"],
} as const;

export type NazwaIkony = keyof typeof sciezki | "start";

type IkonaProps = {
  nazwa: NazwaIkony;
  rozmiar?: number;
  grubosc?: number;
};

export function Ikona({ nazwa, rozmiar = 16, grubosc = 2 }: IkonaProps) {
  if (nazwa === "start") {
    return (
      <svg width={rozmiar} height={rozmiar} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M8 5v14l11-7z" />
      </svg>
    );
  }

  return (
    <svg
      width={rozmiar}
      height={rozmiar}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={grubosc}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {sciezki[nazwa].map((d) => (
        <path key={d} d={d} />
      ))}
    </svg>
  );
}
