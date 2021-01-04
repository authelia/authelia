import React from "react";

import { mount } from "enzyme";

import FixedTextField from "./FixedTextField";

it("renders without crashing", () => {
    mount(<FixedTextField />);
});
