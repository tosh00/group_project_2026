import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { daneZestawu, zestaw } from "../mocks/dane.ts";
import { BladApi } from "./klient.ts";
import { useDaneZestawu } from "./zapytania.ts";

function opakowanie() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("useDaneZestawu", () => {
  it("pobiera dane wejściowe zestawu", async () => {
    const { result } = renderHook(() => useDaneZestawu(zestaw.id), { wrapper: opakowanie() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(daneZestawu);
    expect(result.current.data?.zajecia[0]?.powiazanie).toBeNull();
  });

  it("nie wysyła żądania bez identyfikatora zestawu", () => {
    const { result } = renderHook(() => useDaneZestawu(undefined), { wrapper: opakowanie() });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("zwraca BladApi dla nieznanego zestawu", async () => {
    const { result } = renderHook(() => useDaneZestawu("nie-ma"), { wrapper: opakowanie() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(BladApi);
    expect(result.current.error).toMatchObject({ status: 404, kod: "nie_znaleziono" });
  });
});
