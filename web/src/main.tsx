import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createBrowserRouter, RouterProvider } from "react-router";
import { trasy } from "./app/router.tsx";
import "./styles/tokens.css";
import "./styles/global.css";

// `pnpm dev:mock` (VITE_MOCK=1): odpowiedzi z src/mocks zamiast backendu. Kod MSW nie trafia do zwykłego builda.
async function uruchomMocki() {
  if (import.meta.env.VITE_MOCK !== "1") {
    return;
  }

  const { worker } = await import("./mocks/browser.ts");
  await worker.start({ onUnhandledFrame: "bypass" });
}

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 10_000, retry: 1 } },
});
const router = createBrowserRouter(trasy);

void uruchomMocki().then(() => {
  createRoot(document.getElementById("root")!).render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>,
  );
});
