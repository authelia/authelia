// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render } from "@testing-library/react";

import SuccessIcon from "@components/SuccessIcon";

it("renders without crashing", () => {
    render(<SuccessIcon />);
});
