# 0007. Router HTTP: chi

- Status: proponowana
- Data: 2026-11-18
- Decydują: @KajetanJanusz, @krzysiu-w

## Kontekst

Backend (`cmd/server`, `internal/api`) potrzebuje routera HTTP: parametry ścieżki (`/uruchomienia/{id}`),
grupy tras pod `/api`, middleware (logowanie, odzyskiwanie po panice, limit rozmiaru żądania)
i strumień SSE z postępem solvera. Otwarta kwestia #2.

## Decyzja

Używamy `github.com/go-chi/chi/v5`. Handlery mają zwykłą sygnaturę `http.HandlerFunc`.

## Rozważane opcje

- **chi:** zgodny z `net/http` (handlery i middleware ze standardowej biblioteki działają bez adapterów),
  mały, bez zależności, gotowe middleware (`Recoverer`, `RequestID`, `Timeout`). Minus: jedna zależność więcej.
- **Samo `net/http` (Go 1.22+):** `ServeMux` ma już metody i parametry ścieżki. Minus: brak grup tras
  i middleware per grupa, trzeba to pisać samemu.
- **Gin / Echo:** własny typ kontekstu zamiast `http.Handler`, więcej magii, trudniej mieszać ze standardem.
  Nie potrzebujemy ich dodatkowych funkcji.

## Konsekwencje

- Handlery są niezależne od routera (`http.HandlerFunc`), więc ewentualna zmiana na `net/http` jest tania.
- SSE piszemy ręcznie na `http.Flusher`, bez biblioteki.
- Z `docs/open-questions.md` znika kwestia #2.
