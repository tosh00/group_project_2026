import { describe, expect, it } from "vitest";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "../../mocks/server.ts";
import { uruchomienie, zestaw } from "../../mocks/dane.ts";
import { renderTrase } from "../../test/render.tsx";

const bezPlanow = () => server.use(http.get("/api/uruchomienia", () => HttpResponse.json([])));

describe("StronaPulpit", () => {
  it("bez zestawów pokazuje strefę wgrywania i trzy kroki (1a)", async () => {
    server.use(http.get("/api/zestawy", () => HttpResponse.json([])));

    renderTrase("/");

    expect(await screen.findByRole("heading", { name: "Wgraj dane, żeby ułożyć plan" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Wybierz plik z dysku" })).toBeInTheDocument();
    expect(screen.getByText("Wygeneruj plan")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Twoje zestawy i plany" })).not.toBeInTheDocument();
  });

  it("przy istniejących zestawach u góry jest pole na kolejny plik, a niżej lista bez „aktywny”", async () => {
    bezPlanow();

    renderTrase("/");

    expect(await screen.findByRole("heading", { name: "Wgraj nowy zestaw danych" })).toBeInTheDocument();
    const lista = screen.getByRole("region", { name: "Twoje zestawy i plany" });
    expect(within(lista).getByText(zestaw.nazwa)).toBeInTheDocument();
    expect(within(lista).getByText("Brak planu")).toBeInTheDocument();
    expect(screen.queryByText("aktywny")).not.toBeInTheDocument();
    expect(screen.queryByText("Zawartość tego zestawu")).not.toBeInTheDocument();
  });

  it("otwiera okno ostrzeżeń z wiersza zestawu (1b)", async () => {
    bezPlanow();
    const uzytkownik = userEvent.setup();
    renderTrase("/");

    await uzytkownik.click(await screen.findByRole("button", { name: /3 uwagi do danych/ }));

    const okno = await screen.findByRole("dialog", { name: "Ostrzeżenia w danych · 3" });
    expect(within(okno).getByText("sale[B151].pojemnosc")).toBeInTheDocument();
    expect(within(okno).getByRole("button", { name: "Generuj plan mimo to" })).toBeInTheDocument();
  });

  it("dla planu z atrapą walidatora pokazuje „Status niesprawdzony”, nie zielony status", async () => {
    renderTrase("/");

    expect(await screen.findAllByText("Status niesprawdzony")).not.toHaveLength(0);
    expect(screen.queryByText(/Plan poprawny/)).not.toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /Otwórz plan/ })[0]).toHaveAttribute(
      "href",
      `/uruchomienia/${uruchomienie.id}`,
    );
  });

  it("pokazuje wszystkie błędy pliku z serwera (1d)", async () => {
    server.use(
      http.post("/api/zestawy", () =>
        HttpResponse.json(
          {
            kod: "niespojne_dane",
            komunikat: "Dane są niezgodne z formatem dane_planu.json.",
            szczegoly: [
              'zajecia Z004: dozwolone_sale: nieznany identyfikator "X999"',
              'grupy INF1-R3-ALL: rocznik: nieznany identyfikator "NIE"',
            ],
          },
          { status: 422 },
        ),
      ),
    );
    const uzytkownik = userEvent.setup();
    const { container } = renderTrase("/");
    await screen.findByRole("heading", { name: "Wgraj nowy zestaw danych" });

    const plik = new File(['{"meta":{}}'], "zly.json", { type: "application/json" });
    await uzytkownik.upload(container.querySelector<HTMLInputElement>('input[type="file"]')!, plik);

    expect(await screen.findByText("Nie wgrano pliku · 2 błędy blokujące")).toBeInTheDocument();
    expect(screen.getByText("zajecia[Z004].dozwolone_sale")).toBeInTheDocument();
    expect(screen.getByText('Nieznany identyfikator "NIE"')).toBeInTheDocument();
  });

  it("zgłasza niepoprawny JSON bez wysyłania pliku", async () => {
    const uzytkownik = userEvent.setup();
    const { container } = renderTrase("/");
    await screen.findByRole("heading", { name: "Wgraj nowy zestaw danych" });

    const plik = new File(["{nie json"], "zepsuty.json", { type: "application/json" });
    await uzytkownik.upload(container.querySelector<HTMLInputElement>('input[type="file"]')!, plik);

    expect(await screen.findByText("Nie wgrano pliku · 1 błąd blokujący")).toBeInTheDocument();
    expect(screen.getByText(/Niepoprawny JSON/)).toBeInTheDocument();
  });

  it("„Generuj” uruchamia solver i przechodzi do kalendarza", async () => {
    bezPlanow();
    server.use(
      http.post("/api/zestawy/:zestawId/uruchomienia", () =>
        HttpResponse.json({ ...uruchomienie, status: "w_kolejce" }, { status: 202 }),
      ),
    );
    const uzytkownik = userEvent.setup();
    renderTrase("/");

    await uzytkownik.click(await screen.findByRole("button", { name: "Generuj" }));

    expect(await screen.findByText("Plan rocznika INF1-R1")).toBeInTheDocument();
  });

  it("pokazuje komunikat, gdy kolejka solvera jest pełna", async () => {
    bezPlanow();
    server.use(
      http.post("/api/zestawy/:zestawId/uruchomienia", () =>
        HttpResponse.json(
          { kod: "kolejka_pelna", komunikat: "Kolejka uruchomień jest pełna, spróbuj później." },
          { status: 503 },
        ),
      ),
    );
    const uzytkownik = userEvent.setup();
    renderTrase("/");

    await uzytkownik.click(await screen.findByRole("button", { name: "Generuj" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Kolejka uruchomień jest pełna");
  });
});
