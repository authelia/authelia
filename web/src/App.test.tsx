import React from "react";

import { shallow } from "enzyme";

import App from "@root/App";

it("renders without crashing", () => {
    shallow(<App />);
});
