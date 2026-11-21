import type { Schematy } from "../api/klient.ts";

// Przykładowe odpowiedzi API do pracy bez backendu i do testów. Typy z api/openapi.yaml,
// więc zmiana kontraktu bez poprawienia mocków nie przejdzie `pnpm typecheck`.

export const zdrowie: Schematy["Zdrowie"] = { status: "ok", baza: "ok", walidator: "atrapa" };

export const wagiDomyslne: Required<Schematy["Wagi"]> = {
  okienka_studentow: 10,
  przeciazony_dzien: 15,
  okienka_prowadzacych: 5,
  po_17: 4,
  przed_9: 3,
  za_duza_sala: 6,
  preferencja_dni: 8,
  nie_po_17: 20,
  dzien_wolny: 10,
  przerwa_max_2h: 8,
  pora_dnia: 3,
  sale_a: 3,
  wyklady_rocznika: 2,
};

export const zestaw: Schematy["Zestaw"] = {
  id: "7d2f0c1e-4b5a-4c8e-9f10-2a3b4c5d6e7f",
  nazwa: "Informatyka, semestr zimowy 2026/27",
  utworzono: "2026-11-21T09:30:00Z",
  podsumowanie: {
    zajecia: 174,
    sale: 55,
    sale_do_planowania: 30,
    prowadzacy: 50,
    roczniki: 6,
    grupy: 51,
    studenci: 372,
  },
  ostrzezenia: [
    { kod: "pojemnosc_szacowana", opis: "Sala B151: pojemność szacowana (16 miejsc).", sala_id: "B151" },
    { kod: "pojemnosc_szacowana", opis: "Sala B153: pojemność szacowana (16 miejsc).", sala_id: "B153" },
    { kod: "pojemnosc_szacowana", opis: "Sala B472: pojemność szacowana (16 miejsc).", sala_id: "B472" },
  ],
};

export const uruchomienie: Schematy["Uruchomienie"] = {
  id: "0b9a8c7d-6e5f-4a3b-8c2d-1e0f9a8b7c6d",
  zestaw_id: zestaw.id,
  status: "poprawny",
  wagi: wagiDomyslne,
  seed: 42,
  iteracje: 1_000_000,
  zatrzymane: false,
  liczba_naruszen: 0,
  koszt: 412.5,
  blad: null,
  utworzono: "2026-11-21T09:31:00Z",
  rozpoczeto: "2026-11-21T09:31:00Z",
  zakonczono: "2026-11-21T09:31:04Z",
};
