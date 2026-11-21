import "@testing-library/jest-dom/vitest";
import { afterAll, afterEach, beforeAll } from "vitest";
import { cleanup } from "@testing-library/react";
import { server } from "../mocks/server.ts";

// jsdom nie ma API, których używają Radix (Popover) i cmdk (wyszukiwarka kalendarza).
class ResizeObserverTestowy {
  observe() {}
  unobserve() {}
  disconnect() {}
}
globalThis.ResizeObserver ??= ResizeObserverTestowy;
Element.prototype.scrollIntoView ??= function () {};

// Względne adresy (`/api/...`) w jsdom rozwiązują się względem tego originu.
beforeAll(() => server.listen({ onUnhandledFrame: "error" }));
afterEach(() => {
  cleanup();
  server.resetHandlers();
});
afterAll(() => server.close());
