import React from "react";

import { mount } from "enzyme";

import PieChartIcon from "./PieChartIcon";

it("renders without crashing", () => {
    mount(<PieChartIcon progress={40} />);
});
