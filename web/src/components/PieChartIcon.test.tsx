import React from "react";

import { mount } from "enzyme";

import PieChartIcon from "@components/PieChartIcon";

it("renders without crashing", () => {
    mount(<PieChartIcon progress={40} />);
});
