import { fireEvent, render, screen } from "@testing-library/react";

import {
    InputGroup,
    InputGroupAddon,
    InputGroupButton,
    InputGroupInput,
    InputGroupText,
} from "@components/UI/InputGroup";

describe("InputGroup", () => {
    it("renders the group and its control", () => {
        const { container } = render(
            <InputGroup className="custom">
                <InputGroupInput id="search" placeholder="Search" />
            </InputGroup>,
        );

        expect(container.querySelector('[data-slot="input-group"]')).toHaveClass("custom");
        expect(container.querySelector('[data-slot="input-group-control"]')).toBeInTheDocument();
    });
});

describe("InputGroupAddon", () => {
    it("defaults to an inline start alignment", () => {
        render(
            <InputGroup>
                <InputGroupAddon>
                    <InputGroupText>@</InputGroupText>
                </InputGroupAddon>
                <InputGroupInput id="handle" />
            </InputGroup>,
        );

        expect(document.querySelector('[data-slot="input-group-addon"]')).toHaveAttribute("data-align", "inline-start");
    });

    it.each(["block-end", "block-start", "inline-end", "inline-start"] as const)(
        "supports the %s alignment",
        (align) => {
            render(
                <InputGroup>
                    <InputGroupAddon align={align}>
                        <InputGroupText>@</InputGroupText>
                    </InputGroupAddon>
                    <InputGroupInput id="handle" />
                </InputGroup>,
            );

            expect(document.querySelector('[data-slot="input-group-addon"]')).toHaveAttribute("data-align", align);
        },
    );

    it("focuses the input when the addon is clicked", () => {
        render(
            <InputGroup>
                <InputGroupAddon>
                    <InputGroupText>@</InputGroupText>
                </InputGroupAddon>
                <InputGroupInput id="handle" />
            </InputGroup>,
        );

        fireEvent.click(document.querySelector('[data-slot="input-group-addon"]') as HTMLElement);

        expect(document.getElementById("handle")).toHaveFocus();
    });

    it("does not steal focus when a button inside the addon is clicked", () => {
        const onClick = vi.fn();

        render(
            <InputGroup>
                <InputGroupAddon>
                    <InputGroupButton onClick={onClick}>Go</InputGroupButton>
                </InputGroupAddon>
                <InputGroupInput id="handle" />
            </InputGroup>,
        );

        fireEvent.click(screen.getByText("Go"));

        expect(onClick).toHaveBeenCalled();
        expect(document.getElementById("handle")).not.toHaveFocus();
    });
});

describe("InputGroupButton", () => {
    it("defaults to an extra small ghost button", () => {
        render(<InputGroupButton>Go</InputGroupButton>);

        const button = screen.getByRole("button", { name: "Go" });
        expect(button).toHaveAttribute("type", "button");
        expect(button).toHaveAttribute("data-size", "xs");
    });

    it.each(["icon-sm", "icon-xs", "sm", "xs"] as const)("supports the %s size", (size) => {
        render(<InputGroupButton size={size}>Go</InputGroupButton>);

        expect(screen.getByRole("button", { name: "Go" })).toHaveAttribute("data-size", size);
    });

    it("accepts an explicit type", () => {
        render(<InputGroupButton type="submit">Submit</InputGroupButton>);

        expect(screen.getByRole("button", { name: "Submit" })).toHaveAttribute("type", "submit");
    });
});

describe("InputGroupText", () => {
    it("renders its children", () => {
        render(<InputGroupText className="custom">USD</InputGroupText>);

        expect(screen.getByText("USD")).toHaveClass("custom");
    });
});
