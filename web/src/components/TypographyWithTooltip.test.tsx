import React from "react";

import { render } from "@testing-library/react";

import TypographyWithTooltip, { Props } from "@components/TypographyWithTooltip";

const defaultProps: Props = {
    value: "Example",
    variant: "h5",
};

it("renders without crashing", () => {
    render(<TypographyWithTooltip {...defaultProps} />);
});

it("renders with tooltip without crashing", () => {
    const props: Props = {
        ...defaultProps,
        tooltip: "A tooltip",
    };
    render(<TypographyWithTooltip {...props} />);
});

it("renders the text correctly", () => {
    const props: Props = {
        ...defaultProps,
        value: "Test text",
    };
    const { getByText } = render(<TypographyWithTooltip {...props} />);
    const element = getByText(props.value!);
    expect(element).toBeInTheDocument();
});

it("renders the tooltip correctly", () => {
    const props: Props = {
        ...defaultProps,
        tooltip: "Test tooltip",
    };
    const { getByText } = render(<TypographyWithTooltip {...props} />);
    const element = getByText(props.value!);
    expect(element).toHaveAttribute("aria-label", props.tooltip);
});
