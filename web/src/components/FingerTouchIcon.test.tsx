import React from "react";

import { mount } from "enzyme";

import FingerTouchIcon from "./FingerTouchIcon";

it("renders without crashing", () => {
    mount(<FingerTouchIcon size={32} />);
});
