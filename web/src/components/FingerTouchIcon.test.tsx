// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import React from "react";

import { render } from "@testing-library/react";

import FingerTouchIcon from "@components/FingerTouchIcon";

it("renders without crashing", () => {
    render(<FingerTouchIcon size={32} />);
});

it("renders animated without crashing", () => {
    render(<FingerTouchIcon size={32} animated />);
});

it("renders animated and strong without crashing", () => {
    render(<FingerTouchIcon size={32} animated strong />);
});
