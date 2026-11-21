import { useQuery } from "@tanstack/react-query";
import { api, BladApi } from "./klient.ts";

// Wszystkie klucze zapytań w jednym miejscu, żeby unieważnianie było spójne.
export const klucze = {
  zdrowie: ["zdrowie"] as const,
  wagiDomyslne: ["wagi", "domyslne"] as const,
  zestawy: ["zestawy"] as const,
  zestaw: (zestawId: string) => ["zestawy", zestawId] as const,
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
