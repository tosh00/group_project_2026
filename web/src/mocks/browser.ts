import { setupWorker } from "msw/browser";
import { handlers } from "./handlers.ts";
import { handlerySse } from "./handlerySse.ts";

export const worker = setupWorker(...handlers, ...handlerySse);
