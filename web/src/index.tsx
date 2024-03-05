import React from "react";

import { createRoot } from "react-dom/client";

import "@root/index.css";
import App from "@root/App";
import * as serviceWorker from "@root/serviceWorker";
import "@i18n/index";

const nonce = document.head.querySelector("[property=csp-nonce][content]")?.getAttribute("content") || undefined;
const root = createRoot(document.getElementById("root")!);
root.render(<App nonce={nonce} />);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
