import React from "react";

import { render } from "@testing-library/react";

import LinearProgressBar from "@components/LinearProgressBar";

it("renders without crashing", () => {
    render(<LinearProgressBar value={40} />);
});

it("renders adjusted height without crashing", () => {
    render(<LinearProgressBar value={40} height={2} />);
});
