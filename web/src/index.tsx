import React, { StrictMode } from "react";

import { createRoot } from "react-dom/client";

import "@root/index.css";
import App from "@root/App";
import "@i18n/index";

const nonce = document.head.querySelector("[property=csp-nonce][content]")?.getAttribute("content") || undefined;

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <App nonce={nonce} />
    </StrictMode>,
);
