import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router";
import { trasy } from "./app/router.tsx";
import "./styles/tokens.css";
import "./styles/global.css";

const router = createBrowserRouter(trasy);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
);
