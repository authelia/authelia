import React from "react";

import { mount } from "enzyme";

import TimerIcon from "./TimerIcon";

it("renders without crashing", () => {
    mount(<TimerIcon width={32} height={32} />);
});
