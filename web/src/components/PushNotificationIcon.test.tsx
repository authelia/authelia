import React from "react";

import { mount } from "enzyme";

import PushNotificationIcon from "./PushNotificationIcon";

it("renders without crashing", () => {
    mount(<PushNotificationIcon width={32} height={32} />);
});
