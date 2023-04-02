// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import React from "react";

import { render } from "@testing-library/react";

import App from "@root/App";
import "@i18n/index.ts";

it("renders without crashing", () => {
    render(<App />);
});
