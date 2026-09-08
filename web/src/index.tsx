import { StrictMode } from "react";

import { createRoot } from "react-dom/client";

import "@root/index.css";
import App from "@root/App";
import "@i18n/index";

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <App />
    </StrictMode>,
);
