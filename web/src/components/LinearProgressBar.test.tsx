import React from "react";

import { mount } from "enzyme";

import LinearProgressBar from "./LinearProgressBar";

it("renders without crashing", () => {
    mount(<LinearProgressBar value={40} />);
});
