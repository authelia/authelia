import { render, screen } from "@testing-library/react";

import {
    Field,
    FieldContent,
    FieldDescription,
    FieldError,
    FieldGroup,
    FieldLabel,
    FieldLegend,
    FieldSeparator,
    FieldSet,
    FieldTitle,
} from "@components/UI/Field";

describe("FieldSet", () => {
    it("renders a fieldset", () => {
        const { container } = render(<FieldSet>content</FieldSet>);
        const el = container.querySelector('[data-slot="field-set"]');
        expect(el?.tagName).toBe("FIELDSET");
        expect(el).toHaveTextContent("content");
    });

    it("merges a custom class name", () => {
        const { container } = render(<FieldSet className="custom" />);
        expect(container.querySelector('[data-slot="field-set"]')).toHaveClass("custom");
    });
});

describe("FieldLegend", () => {
    it("defaults to the legend variant", () => {
        render(<FieldLegend>Legend</FieldLegend>);
        expect(screen.getByText("Legend")).toHaveAttribute("data-variant", "legend");
    });

    it("supports the label variant", () => {
        render(<FieldLegend variant="label">Label</FieldLegend>);
        expect(screen.getByText("Label")).toHaveAttribute("data-variant", "label");
    });
});

describe("FieldGroup", () => {
    it("renders its children", () => {
        const { container } = render(<FieldGroup>grouped</FieldGroup>);
        expect(container.querySelector('[data-slot="field-group"]')).toHaveTextContent("grouped");
    });
});

describe("Field", () => {
    it("defaults to a vertical orientation", () => {
        render(<Field>field</Field>);
        expect(screen.getByRole("group")).toHaveAttribute("data-orientation", "vertical");
    });

    it.each(["horizontal", "responsive", "vertical"] as const)("supports the %s orientation", (orientation) => {
        render(<Field orientation={orientation}>field</Field>);
        expect(screen.getByRole("group")).toHaveAttribute("data-orientation", orientation);
    });

    it("merges a custom class name", () => {
        render(<Field className="custom">field</Field>);
        expect(screen.getByRole("group")).toHaveClass("custom");
    });
});

describe("FieldContent", () => {
    it("renders its children", () => {
        const { container } = render(<FieldContent>inner</FieldContent>);
        expect(container.querySelector('[data-slot="field-content"]')).toHaveTextContent("inner");
    });
});

describe("FieldLabel", () => {
    it("renders a label", () => {
        const { container } = render(<FieldLabel htmlFor="input">Name</FieldLabel>);
        expect(container.querySelector('[data-slot="field-label"]')).toHaveTextContent("Name");
    });
});

describe("FieldTitle", () => {
    it("renders its children", () => {
        render(<FieldTitle>Title</FieldTitle>);
        expect(screen.getByText("Title")).toBeInTheDocument();
    });
});

describe("FieldDescription", () => {
    it("renders a paragraph", () => {
        render(<FieldDescription>Helpful text</FieldDescription>);
        expect(screen.getByText("Helpful text").tagName).toBe("P");
    });
});

describe("FieldSeparator", () => {
    it("renders without content", () => {
        const { container } = render(<FieldSeparator />);
        const el = container.querySelector('[data-slot="field-separator"]');
        expect(el).toHaveAttribute("data-content", "false");
        expect(container.querySelector('[data-slot="field-separator-content"]')).toBeNull();
    });

    it("renders with content", () => {
        const { container } = render(<FieldSeparator>or</FieldSeparator>);
        expect(container.querySelector('[data-slot="field-separator"]')).toHaveAttribute("data-content", "true");
        expect(screen.getByText("or")).toBeInTheDocument();
    });
});

describe("FieldError", () => {
    it("renders nothing without children or errors", () => {
        const { container } = render(<FieldError />);
        expect(container.firstChild).toBeNull();
    });

    it("renders nothing for an empty error list", () => {
        const { container } = render(<FieldError errors={[]} />);
        expect(container.firstChild).toBeNull();
    });

    it("prefers explicit children", () => {
        render(<FieldError errors={[{ message: "ignored" }]}>Custom message</FieldError>);
        expect(screen.getByRole("alert")).toHaveTextContent("Custom message");
        expect(screen.queryByText("ignored")).not.toBeInTheDocument();
    });

    it("renders a single error as plain text", () => {
        render(<FieldError errors={[{ message: "Required" }]} />);
        expect(screen.getByRole("alert")).toHaveTextContent("Required");
        expect(document.querySelector("ul")).toBeNull();
    });

    it("deduplicates identical errors", () => {
        render(<FieldError errors={[{ message: "Required" }, { message: "Required" }]} />);
        expect(screen.getByRole("alert")).toHaveTextContent("Required");
        expect(document.querySelector("ul")).toBeNull();
    });

    it("renders several errors as a list", () => {
        render(<FieldError errors={[{ message: "Too short" }, { message: "Needs a number" }]} />);
        expect(screen.getByText("Too short").tagName).toBe("LI");
        expect(screen.getByText("Needs a number").tagName).toBe("LI");
    });

    it("skips entries without a message", () => {
        render(<FieldError errors={[{ message: "Too short" }, undefined, {}]} />);
        expect(screen.getByRole("alert")).toBeInTheDocument();
        expect(document.querySelectorAll("li")).toHaveLength(1);
    });

    it("merges a custom class name", () => {
        render(<FieldError className="custom">Boom</FieldError>);
        expect(screen.getByRole("alert")).toHaveClass("custom");
    });
});
