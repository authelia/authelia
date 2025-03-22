import React from "react";

import createCache from "@emotion/cache";
import { CacheProvider } from "@emotion/react";
import { createRoot } from "react-dom/client";
import { TssCacheProvider } from "tss-react";

import "@root/index.css";
import App from "@root/App";

import "@i18n/index";

const nonce = document.head.querySelector("[property=csp-nonce][content]")?.getAttribute("content") || undefined;

const muiCache = createCache({
    key: "mui",
    prepend: true,
    nonce: nonce,
});

const tssCache = createCache({
    key: "tss",
    nonce: nonce,
});

createRoot(document.getElementById("root")!).render(
    <CacheProvider value={muiCache}>
        <TssCacheProvider value={tssCache}>
            <App />
        </TssCacheProvider>
    </CacheProvider>,
);
