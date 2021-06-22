import React from "react";

import { mount } from "enzyme";

import InformationIcon from "@components/InformationIcon";

it("renders without crashing", () => {
    mount(<InformationIcon />);
});
