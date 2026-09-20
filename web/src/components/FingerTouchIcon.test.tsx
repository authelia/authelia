// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render } from "@testing-library/react";

import FingerTouchIcon from "@components/FingerTouchIcon";

it("renders without crashing", () => {
    render(<FingerTouchIcon size={32} />);
});

it("renders with animation", () => {
    render(<FingerTouchIcon size={32} animated />);
});

it("renders with strong animation", () => {
    render(<FingerTouchIcon size={32} animated strong />);
});
