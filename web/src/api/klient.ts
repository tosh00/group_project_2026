import createClient from "openapi-fetch";
import type { components, paths } from "./schema.gen.ts";

export type Schematy = components["schemas"];

// Jedyny klient HTTP aplikacji. Typy ścieżek i odpowiedzi pochodzą z api/openapi.yaml (`pnpm gen:api`).
// Pełny adres (z originem strony), bo openapi-fetch tworzy `Request`, a ten w Node/testach nie przyjmuje ścieżki względnej.
// `fetch` pobierany przy każdym żądaniu, nie przy imporcie: MSW podmienia go dopiero po starcie.
export const api = createClient<paths>({
  baseUrl: `${window.location.origin}/api`,
  fetch: (zadanie) => globalThis.fetch(zadanie),
});

type OdpowiedzKlienta<T> = { data?: T; error?: unknown; response: Response };

// Błąd z API w formacie `Blad` (`{kod, komunikat, szczegoly}`), rzucany z zapytań TanStack Query.
export class BladApi extends Error {
  readonly status: number;
  readonly kod: string;
  readonly szczegoly: string[];

  constructor({ status, kod, komunikat, szczegoly }: { status: number } & Schematy["Blad"]) {
    super(komunikat);
    this.name = "BladApi";
    this.status = status;
    this.kod = kod;
    this.szczegoly = szczegoly ?? [];
  }

  // Zamienia odpowiedź openapi-fetch na dane albo rzuca BladApi; jedno miejsce mapowania błędów.
  static async dane<T>(zadanie: Promise<OdpowiedzKlienta<T>>): Promise<T> {
    const { data, error, response } = await zadanie;
    if (response.ok) {
      return data as T;
    }

    throw BladApi.zTresci({ status: response.status, tresc: error });
  }

  static zTresci({ status, tresc }: { status: number; tresc: unknown }): BladApi {
    if (BladApi.czyBlad(tresc)) {
      return new BladApi({ status, ...tresc });
    }

    return new BladApi({ status, kod: "blad_serwera", komunikat: `Serwer odpowiedział kodem ${status}.` });
  }

  static czyBlad(tresc: unknown): tresc is Schematy["Blad"] {
    return (
      typeof tresc === "object" &&
      tresc !== null &&
      typeof (tresc as Schematy["Blad"]).kod === "string" &&
      typeof (tresc as Schematy["Blad"]).komunikat === "string"
    );
  }
}
