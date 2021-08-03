import React from "react";

import { render } from "@testing-library/react";

import NotificationBar from "@components/NotificationBar";

it("renders without crashing", () => {
    render(<NotificationBar onClose={() => {}} />);
});
