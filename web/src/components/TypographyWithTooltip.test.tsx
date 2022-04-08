import React from "react";

import { render } from "@testing-library/react";

import TypographyWithTooltip from "@components/TypographyWithTootip";

it("renders without crashing", () => {
    render(<TypographyWithTooltip value={"Example"} variant={"h5"} />);
});

it("renders with tooltip without crashing", () => {
    render(<TypographyWithTooltip value={"Example"} tooltip={"A tooltip"} variant={"h5"} />);
});
