import React from "react";

import { mount } from "enzyme";

import LinearProgressBar from "@components/LinearProgressBar";

it("renders without crashing", () => {
    mount(<LinearProgressBar value={40} />);
});
