import React from "react";

import { render } from "@testing-library/react";

import FixedTextField from "@components/FixedTextField";

it("renders without crashing", () => {
    render(<FixedTextField />);
});
