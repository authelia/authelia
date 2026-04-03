import { render, screen } from "@testing-library/react";

import TypographyWithTooltip, { Props } from "@components/TypographyWithTooltip";

const defaultProps: Props = {
    value: "Example",
    variant: "h5",
};

it("renders without crashing", () => {
    render(<TypographyWithTooltip {...defaultProps} />);
});

it("renders with tooltip", () => {
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
    render(<TypographyWithTooltip {...props} />);
    expect(screen.getByText(props.value!)).toBeInTheDocument();
});

it("renders the tooltip correctly", () => {
    const props: Props = {
        ...defaultProps,
        tooltip: "Test tooltip",
    };
    render(<TypographyWithTooltip {...props} />);
    expect(screen.getByText(props.value!)).toHaveAttribute("aria-label", props.tooltip);
});
