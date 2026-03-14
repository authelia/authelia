import { StrictMode } from "react";

import createCache from "@emotion/cache";
import { CacheProvider } from "@emotion/react";
import { createRoot } from "react-dom/client";

import "@root/index.css";
import App from "@root/App";
import "@i18n/index";

const nonce = document.head.querySelector("[property=csp-nonce][content]")?.getAttribute("content") || undefined;

const muiCache = createCache({
    key: "mui",
    nonce: nonce,
    prepend: true,
});

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <CacheProvider value={muiCache}>
            <App />
        </CacheProvider>
    </StrictMode>,
);
