// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import React from "react";

import { render } from "@testing-library/react";

import FailureIcon from "@components/FailureIcon";

it("renders without crashing", () => {
    render(<FailureIcon />);
});
