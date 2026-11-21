import { afterEach, describe, expect, it, vi } from "vitest";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "../../mocks/server.ts";
import { uruchomienie } from "../../mocks/dane.ts";
import { renderTrase } from "../../test/render.tsx";
import { dane, plan, raportZNaruszeniami } from "./model/dane.testowe.ts";

const adres = `/uruchomienia/${uruchomienie.id}`;

const planTestowy = (raport = raportZNaruszeniami) =>
  server.use(
    http.get("/api/zestawy/:zestawId/dane", () => HttpResponse.json(dane)),
    http.get("/api/uruchomienia/:id/plan", () => HttpResponse.json(plan)),
    http.get("/api/uruchomienia/:id/raport", () => HttpResponse.json(raport)),
  );

afterEach(() => vi.unstubAllGlobals());

describe("StronaKalendarz", () => {
  it("pokazuje plan rocznika z kartami, statusem i niesprawdzonym walidatorem", async () => {
    renderTrase(adres);

    expect(await screen.findByText("Plan rocznika INF1-R1")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Status niesprawdzony/ })).toBeInTheDocument();
    expect(screen.getByText("Plan nie został jeszcze sprawdzony")).toBeInTheDocument();
    expect(screen.queryByText(/Koszt planu|Seed/)).not.toBeInTheDocument();
    expect(screen.getByLabelText(/Analiza matematyczna I, wykład, sala A281/)).toBeInTheDocument();
  });

  it("pokazuje naruszenia: status, czerwoną kartę i listę w panelu", async () => {
    planTestowy();
    renderTrase(`${adres}?widok=rocznik&id=INF1-R1&tydzien=2`);

    expect(await screen.findByRole("button", { name: /Plan z naruszeniami · 1/ })).toBeInTheDocument();
    const panel = screen.getByRole("complementary", { name: "Panel statusu planu" });
    expect(within(panel).getByText("Sala B151 ma 16 miejsc, potrzeba 24")).toBeInTheDocument();
    expect(within(panel).getByText("pon")).toBeInTheDocument();
    expect(screen.getByLabelText(/naruszenie: Sala B151 ma 16 miejsc/)).toBeInTheDocument();
  });

  it("kliknięcie sali na karcie otwiera plan sali, a „Wstecz” wraca do rocznika", async () => {
    planTestowy();
    const uzytkownik = userEvent.setup();
    renderTrase(adres);

    const karta = await screen.findByLabelText(/Analiza matematyczna I, wykład, sala A281/);
    await uzytkownik.click(within(karta).getByRole("button", { name: "A281" }));

    expect(await screen.findByText("Plan sali A281 · Duża Aula")).toBeInTheDocument();
    expect(screen.getByRole("meter", { name: "Zajętość sali" })).toHaveAttribute("aria-valuenow", "6");

    await uzytkownik.click(screen.getByRole("button", { name: "Wstecz do poprzedniego widoku" }));
    expect(await screen.findByText("Plan rocznika INF1-R1")).toBeInTheDocument();
  });

  it("wyszukiwarka otwiera plan prowadzącego", async () => {
    planTestowy();
    const uzytkownik = userEvent.setup();
    renderTrase(adres);

    await uzytkownik.click(await screen.findByRole("button", { name: /INF1-R1/, expanded: false }));
    await uzytkownik.type(screen.getByPlaceholderText("Szukaj: P07, A28, Analiza…"), "P07");
    await uzytkownik.click(await screen.findByRole("option", { name: /^dr P07/ }));

    expect(await screen.findByText("Plan prowadzącego dr P07")).toBeInTheDocument();
    expect(screen.getByText(/Dni wolne:/)).toHaveTextContent("Dni wolne: śr, czw, pt");
  });

  it("pusty wynik filtra pozwala wyczyścić filtry zawężające", async () => {
    planTestowy();
    const uzytkownik = userEvent.setup();
    renderTrase(`${adres}?widok=rocznik&id=INF1-R1&tydzien=3&bez=W`);

    expect(await screen.findByText("Brak zajęć dla INF1-R1 w tygodniu 3")).toBeInTheDocument();
    await uzytkownik.click(screen.getByRole("button", { name: "Wyczyść filtry zawężające" }));

    expect(screen.queryByText(/Brak zajęć dla/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Obecny", pressed: true })).toBeInTheDocument();
  });

  it("w trakcie generowania pokazuje postęp i pozwala zatrzymać", async () => {
    const zrodla: { url: string; close: () => void }[] = [];
    vi.stubGlobal(
      "EventSource",
      class {
        url: string;
        constructor(url: string) {
          this.url = url;
          zrodla.push(this);
        }
        addEventListener() {}
        close() {}
      },
    );
    let zatrzymano = false;
    server.use(
      http.get("/api/uruchomienia/:id", () =>
        HttpResponse.json({
          ...uruchomienie,
          status: "trwa",
          faza: "konstrukcja",
          koszt: null,
          liczba_naruszen: null,
        }),
      ),
      http.post("/api/uruchomienia/:id/zatrzymaj", () => {
        zatrzymano = true;
        return HttpResponse.json(uruchomienie, { status: 202 });
      }),
    );
    const uzytkownik = userEvent.setup();
    renderTrase(adres);

    expect(await screen.findByRole("heading", { name: "Układanie zajęć…" })).toBeInTheDocument();
    expect(zrodla.map((z) => z.url)).toEqual([`/api/uruchomienia/${uruchomienie.id}/postep`]);

    await uzytkownik.click(screen.getByRole("button", { name: "Zatrzymaj" }));
    await vi.waitFor(() => expect(zatrzymano).toBe(true));
  });

  it("pokazuje błąd uruchomienia z opisem", async () => {
    server.use(
      http.get("/api/uruchomienia/:id", () =>
        HttpResponse.json({ ...uruchomienie, status: "blad", blad: "przerwane restartem serwera" }),
      ),
    );
    renderTrase(adres);

    expect(
      await screen.findByRole("heading", { name: "Nie udało się wygenerować planu" }),
    ).toBeInTheDocument();
    expect(screen.getByText("przerwane restartem serwera")).toBeInTheDocument();
  });
});
