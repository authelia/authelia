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

it("renders the correct heading element for variant", () => {
    const props: Props = {
        ...defaultProps,
        value: "Test text",
        variant: "h3",
    };
    render(<TypographyWithTooltip {...props} />);
    const element = screen.getByText(props.value!);
    expect(element.tagName).toBe("H3");
});
