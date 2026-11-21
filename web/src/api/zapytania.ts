import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, BladApi, type Schematy } from "./klient.ts";

// Wszystkie klucze zapytań w jednym miejscu, żeby unieważnianie było spójne.
export const klucze = {
  zdrowie: ["zdrowie"] as const,
  wagiDomyslne: ["wagi", "domyslne"] as const,
  zestawy: ["zestawy"] as const,
  zestaw: (zestawId: string) => ["zestawy", zestawId] as const,
  daneZestawu: (zestawId: string) => ["zestawy", zestawId, "dane"] as const,
  uruchomienia: (zestawId?: string) => ["uruchomienia", { zestawId }] as const,
  uruchomienie: (uruchomienieId: string) => ["uruchomienie", uruchomienieId] as const,
};

export function useZdrowie() {
  return useQuery({
    queryKey: klucze.zdrowie,
    queryFn: () => BladApi.dane(api.GET("/zdrowie")),
    refetchInterval: 30_000,
  });
}

export function useWagiDomyslne() {
  return useQuery({
    queryKey: klucze.wagiDomyslne,
    queryFn: () => BladApi.dane(api.GET("/wagi/domyslne")),
    staleTime: Infinity,
  });
}

export function useZestawy() {
  return useQuery({
    queryKey: klucze.zestawy,
    queryFn: () => BladApi.dane(api.GET("/zestawy")),
  });
}

// Dane wejściowe zestawu (sale, prowadzący, grupy, przedmioty, zajęcia) do kart i filtrów kalendarza.
// Zestaw po wgraniu się nie zmienia, więc dane nie wymagają odświeżania.
export function useDaneZestawu(zestawId: string | undefined) {
  return useQuery({
    queryKey: klucze.daneZestawu(zestawId ?? ""),
    queryFn: () =>
      BladApi.dane(api.GET("/zestawy/{zestawId}/dane", { params: { path: { zestawId: zestawId! } } })),
    enabled: zestawId !== undefined,
    staleTime: Infinity,
  });
}

// Wszystkie uruchomienia (najnowsze pierwsze); pulpit bierze z nich ostatni plan każdego zestawu.
export function useUruchomienia() {
  return useQuery({
    queryKey: klucze.uruchomienia(),
    queryFn: () => BladApi.dane(api.GET("/uruchomienia")),
  });
}

type WgranieZestawu = { dane: Schematy["DaneWejsciowe"]; nazwa?: string; sygnal?: AbortSignal };

export function useWgrajZestaw() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ dane, nazwa, sygnal }: WgranieZestawu) =>
      BladApi.dane(api.POST("/zestawy", { body: dane, params: { query: { nazwa } }, signal: sygnal })),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: klucze.zestawy }),
  });
}

export function useUruchomSolver() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ zestawId, ustawienia }: { zestawId: string; ustawienia?: Schematy["NoweUruchomienie"] }) =>
      BladApi.dane(
        api.POST("/zestawy/{zestawId}/uruchomienia", {
          params: { path: { zestawId } },
          body: ustawienia ?? {},
        }),
      ),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["uruchomienia"] }),
  });
}

// Statusy, po których uruchomienie się nie zmieni (ADR 0006: plan jest przy `poprawny` i `z_naruszeniami`).
const statusyZPlanem: ReadonlySet<Schematy["StatusUruchomienia"]> = new Set(["poprawny", "z_naruszeniami"]);

export function maPlan(uruchomienie: Schematy["Uruchomienie"] | undefined): boolean {
  return uruchomienie !== undefined && statusyZPlanem.has(uruchomienie.status);
}

// Stan uruchomienia. W trakcie aktualizuje go SSE (`usePostepUruchomienia`), więc bez odpytywania.
export function useUruchomienie(uruchomienieId: string) {
  return useQuery({
    queryKey: klucze.uruchomienie(uruchomienieId),
    queryFn: () =>
      BladApi.dane(api.GET("/uruchomienia/{uruchomienieId}", { params: { path: { uruchomienieId } } })),
  });
}

export function useZestaw(zestawId: string | undefined) {
  return useQuery({
    queryKey: klucze.zestaw(zestawId ?? ""),
    queryFn: () =>
      BladApi.dane(api.GET("/zestawy/{zestawId}", { params: { path: { zestawId: zestawId! } } })),
    enabled: zestawId !== undefined,
    staleTime: Infinity,
  });
}

// Plan i raport istnieją dopiero po zakończeniu uruchomienia i już się nie zmieniają.
export function usePlan(uruchomienie: Schematy["Uruchomienie"] | undefined) {
  const uruchomienieId = uruchomienie?.id ?? "";
  return useQuery({
    queryKey: [...klucze.uruchomienie(uruchomienieId), "plan"],
    queryFn: () =>
      BladApi.dane(api.GET("/uruchomienia/{uruchomienieId}/plan", { params: { path: { uruchomienieId } } })),
    enabled: maPlan(uruchomienie),
    staleTime: Infinity,
  });
}

export function useRaport(uruchomienie: Schematy["Uruchomienie"] | undefined) {
  const uruchomienieId = uruchomienie?.id ?? "";
  return useQuery({
    queryKey: [...klucze.uruchomienie(uruchomienieId), "raport"],
    queryFn: () =>
      BladApi.dane(
        api.GET("/uruchomienia/{uruchomienieId}/raport", { params: { path: { uruchomienieId } } }),
      ),
    enabled: maPlan(uruchomienie),
    staleTime: Infinity,
  });
}

export function useZatrzymajUruchomienie() {
  return useMutation({
    mutationFn: (uruchomienieId: string) =>
      BladApi.dane(
        api.POST("/uruchomienia/{uruchomienieId}/zatrzymaj", { params: { path: { uruchomienieId } } }),
      ),
  });
}

// Usuwa zestaw z jego uruchomieniami i planami (nieodwracalne; 409 `trwa_generowanie` w trakcie generowania).
export function useUsunZestaw() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (zestawId: string) =>
      BladApi.dane(api.DELETE("/zestawy/{zestawId}", { params: { path: { zestawId } } })),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: klucze.zestawy }),
        queryClient.invalidateQueries({ queryKey: ["uruchomienia"] }),
      ]),
  });
}
