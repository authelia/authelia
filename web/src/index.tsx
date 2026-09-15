// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { StrictMode } from "react";

import axios from "axios";
import { createRoot } from "react-dom/client";

import "@root/index.css";
import App from "@root/App";
import "@i18n/index";

// The CSRF token for the current session is delivered in this cookie and must be echoed in this header. Axios only
// attaches it to same-origin requests.
axios.defaults.xsrfCookieName = "authelia_csrf";
axios.defaults.xsrfHeaderName = "X-CSRF-Token";

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <App />
    </StrictMode>,
);
