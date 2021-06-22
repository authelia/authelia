import React from "react";

import { mount } from "enzyme";

import NotificationBar from "@components/NotificationBar";

it("renders without crashing", () => {
    mount(<NotificationBar onClose={() => {}} />);
});
