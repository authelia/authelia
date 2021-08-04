import React from "react";

import { render } from "@testing-library/react";

import App from "@root/App";

it("renders without crashing", () => {
    render(<App />);
});
