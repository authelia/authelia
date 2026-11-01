// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render } from "@testing-library/react";

import InformationIcon from "@components/InformationIcon";

it("renders without crashing", () => {
    render(<InformationIcon />);
});
