// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { StrictMode } from "react";

import { createRoot } from "react-dom/client";

import "@root/index.css";
import App from "@root/App";
import "@i18n/index";
import { registerRateLimitInterceptor } from "@services/RateLimit";

registerRateLimitInterceptor();

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <App />
    </StrictMode>,
);
