import React from "react";

import { mount } from "enzyme";

import FixedTextField from "@components/FixedTextField";

it("renders without crashing", () => {
    mount(<FixedTextField />);
});
