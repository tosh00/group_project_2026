# 0011. Stack frontendu: React + Vite, typy z OpenAPI, własna siatka kalendarza

- Status: proponowana
- Data: 2026-11-21
- Decydują: @shokera, @KajetanJanusz

## Kontekst

Frontend (`web/`) jeszcze nie istnieje. `docs/plan.md` wymienia stack wstępnie (React 18, Vite, TanStack Query,
React Router, Vitest), ale bez decyzji o tym, jak front zna typy API, jak rysujemy kalendarz, czym stylujemy
i jak pracujemy bez backendu. Design MVP jest gotowy (canvas „Planer zajęć WFiIS”, 2026-11-21):
pulpit z zestawami, postęp na żywo, kalendarz 5 dni × 12 slotów z filtrami, panel naruszeń, raport.
Kontrakt backendu to `api/openapi.yaml` (0.1), postęp idzie przez SSE.

## Decyzja

- **Baza:** React 19 + TypeScript (`strict`) + Vite, menedżer pakietów pnpm, Node 22+.
- **Dane z serwera:** TanStack Query. Typy i klient generujemy z `api/openapi.yaml`
  (`openapi-typescript` → `web/src/api/schema.gen.ts`, klient `openapi-fetch`); typów API nie piszemy ręcznie.
- **Postęp:** natywny `EventSource` w jednym hooku, który dopisuje stan do cache TanStack Query.
- **Routing:** React Router; stan filtrów kalendarza w parametrach URL (link do konkretnego widoku).
- **Kalendarz:** własny komponent na CSS Grid, bez biblioteki kalendarza.
- **Style:** CSS Modules + zmienne CSS (tokeny z designu) w `web/src/styles/tokens.css`. Na start tylko jasny
  motyw; kolory wyłącznie przez tokeny, więc ciemny motyw to później jeden blok `[data-theme="dark"]`.
- **Komponenty z dostępnością:** Radix UI (dialog, popover, tooltip, suwak), `cmdk` do wyszukiwarki,
  dodawane razem z ekranami, które ich potrzebują.
- **Wykresy:** Recharts (spadek kosztu na żywo, rozbicie kosztu w raporcie).
- **Mock API:** MSW, te same handlery w przeglądarce (`VITE_MOCK=1`) i w testach.
- **Testy i jakość:** Vitest + Testing Library, ESLint + Prettier, CI w `.github/workflows/web.yml`.
- **Etap 2 (ADR 0010):** dnd-kit do edycji ręcznej, dopiero gdy wróci edycja.

## Rozważane opcje

- **FullCalendar / react-big-calendar:** gotowa siatka, ale model oparty na datach i godzinach, a my mamy sloty 1–12,
  maski tygodni 1–14 i 5–6 równoległych kart w komórce. Dostosowanie kosztowałoby więcej niż własna siatka.
- **Tailwind:** szybkie prototypowanie, ale makiety mają gotowe tokeny, a CSS Modules nie dokładają kroku budowania
  ani nowej składni dla osób, które jej nie znają.
- **Ręczne typy + axios:** zero generatora, ale typy rozjeżdżają się z kontraktem bez ostrzeżenia.
- **Next.js:** SSR i routing plikowy są niepotrzebne dla aplikacji jednego użytkownika za API w Go.
- **Biblioteka komponentów (MUI, Mantine):** szybki start, ale własny wygląd trzeba zwalczać, żeby trafić w design.

## Konsekwencje

- Zmiana `api/openapi.yaml` wymaga `pnpm gen:api` w `web/`; CI sprawdza, czy `schema.gen.ts` jest aktualny,
  więc niezgodność kontraktu wychodzi w PR zmieniającym kontrakt.
- Dev server Vite przekazuje `/api` do backendu (`localhost:8080`), więc backend nie potrzebuje CORS.
- Frontend da się rozwijać i testować bez Go i bazy (MSW).
- `docs/plan.md` (sekcja Stack) odsyła do tego ADR-u.
