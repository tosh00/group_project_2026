# Frontend planera

React 19 + TypeScript + Vite, TanStack Query, React Router, MSW, Vitest (ADR 0011).
Design: canvas „Planer zajęć WFiIS” (Claude Design), brief: `docs/design/claude-design-brief.md`.

## Komendy

| Komenda                         | Co robi                                                                                               |
| ------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `pnpm dev`                      | serwer deweloperski na `:5173`, `/api` przez proxy do `http://localhost:8080` (zmiana: `API_URL=...`) |
| `pnpm dev:mock`                 | to samo, ale API z `src/mocks` (MSW), bez backendu i bazy                                             |
| `pnpm test`                     | Vitest + Testing Library (MSW przechwytuje żądania)                                                   |
| `pnpm lint` / `pnpm format`     | ESLint + Prettier                                                                                     |
| `pnpm typecheck` / `pnpm build` | TypeScript, build produkcyjny do `dist/`                                                              |
| `pnpm gen:api`                  | typy z `../api/openapi.yaml` do `src/api/schema.gen.ts`                                               |

## Struktura

```
src/
  api/          klient (openapi-fetch), hooki zapytań, SSE z postępem, schema.gen.ts (generowany)
  app/          router, układ strony, elementy wspólne dla wszystkich stron
  strony/       jedna strona na trasę (pulpit, zestaw, kalendarz, raport)
  mocks/        dane i handlery MSW (przeglądarka: browser.ts, testy: server.ts)
  styles/       tokens.css (kolory, odstępy, fonty z designu), global.css
  test/         setup Vitest i renderTrase / renderZDostawcami
```

## Konwencje

- **API tylko przez `src/api`.** Każde zapytanie to hook w `zapytania.ts` z kluczem z obiektu `klucze`.
  Odpowiedź przechodzi przez `BladApi.dane(...)`, które rzuca `BladApi` (`kod`, `komunikat`, `szczegoly`).
  Typów API nie piszemy ręcznie: bierzemy je z `Schematy["..."]` (`klient.ts`).
- **Postęp uruchomienia:** `usePostepUruchomienia(id)`. Zdarzenia `stan` i `koniec` aktualizują cache
  uruchomienia, więc komponenty czytają stan przez zwykłe zapytanie.
- **Style:** CSS Modules obok komponentu (`Nazwa.module.css`), kolory i odstępy tylko z `tokens.css`.
  Na razie jest tylko jasny motyw; ciemny dojdzie jako blok `[data-theme="dark"]` w `tokens.css`.
  Czerwony tylko dla naruszeń twardych, zielony tylko dla „0 naruszeń”, walidator „atrapa” to stan neutralny.
- **Nazwy domenowe po polsku bez polskich znaków** (`zestaw`, `uruchomienie`, `prowadzacy`), jak w backendzie.
- **Filtry kalendarza w URL** (parametry zapytania), żeby dało się podlinkować widok.
- **Mocki:** nowy endpoint to handler w `mocks/handlers.ts` z danymi typowanymi z `Schematy`. SSE jest w
  `mocks/handlerySse.ts`, tylko w przeglądarce, bo jsdom nie ma `EventSource`.

## Docker

`web/Dockerfile` buduje aplikację i serwuje ją przez nginx (`nginx.conf.template`): `/api` idzie do
`API_UPSTREAM` (domyślnie `http://server:8080`) bez buforowania, żeby SSE działało. Nieznane ścieżki
dostają `index.html`. Całość: `docker compose up --build` w katalogu głównym repo.
