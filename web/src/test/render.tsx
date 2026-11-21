import type { ReactElement } from "react";
import { render } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryRouter, RouterProvider } from "react-router";
import { trasy } from "../app/router.tsx";

// Świeży QueryClient na test (bez retry, żeby błędy API były widoczne od razu).
function nowyQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

export function renderZDostawcami(element: ReactElement) {
  return render(<QueryClientProvider client={nowyQueryClient()}>{element}</QueryClientProvider>);
}

export function renderTrase(sciezka: string) {
  const router = createMemoryRouter(trasy, { initialEntries: [sciezka] });
  return renderZDostawcami(<RouterProvider router={router} />);
}
