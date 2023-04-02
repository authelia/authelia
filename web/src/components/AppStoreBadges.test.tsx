// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import React from "react";

import { render } from "@testing-library/react";

import AppStoreBadges from "@components/AppStoreBadges";

it("renders without crashing", () => {
    render(<AppStoreBadges iconSize={32} appleStoreLink="http://apple" googlePlayLink="http://google" />);
});
