import React from "react";

import { mount } from "enzyme";

import FailureIcon from "@components/FailureIcon";

it("renders without crashing", () => {
    mount(<FailureIcon />);
});
