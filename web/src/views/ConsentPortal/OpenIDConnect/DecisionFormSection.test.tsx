import { render, screen } from "@testing-library/react";

import DecisionFormSection, { DecisionFormSectionItem } from "@views/ConsentPortal/OpenIDConnect/DecisionFormSection";

it("renders the title and the children within a list", () => {
    render(
        <DecisionFormSection id={"section-example"} title={"Permissions"}>
            <li>Example</li>
        </DecisionFormSection>,
    );

    expect(screen.getByRole("heading", { name: "Permissions" })).toBeInTheDocument();
    expect(screen.getByRole("list")).toBeInTheDocument();
    expect(screen.getByText("Example")).toBeInTheDocument();
});

it("assigns the id to the section element", () => {
    const { container } = render(
        <DecisionFormSection id={"section-example"} title={"Permissions"}>
            <li>Example</li>
        </DecisionFormSection>,
    );

    expect(container.querySelector("section")).toHaveAttribute("id", "section-example");
});

it("renders the description when provided", () => {
    render(
        <DecisionFormSection title={"Permissions"} description={"What the application may do"}>
            <li>Example</li>
        </DecisionFormSection>,
    );

    expect(screen.getByText("What the application may do")).toBeInTheDocument();
});

it("does not render a description element when it is not provided", () => {
    const { container } = render(
        <DecisionFormSection title={"Permissions"}>
            <li>Example</li>
        </DecisionFormSection>,
    );

    expect(container.querySelector("p")).not.toBeInTheDocument();
});

it("renders an item with its icon and content", () => {
    render(
        <DecisionFormSection title={"Permissions"}>
            <DecisionFormSectionItem id={"item-example"} icon={<span data-testid={"icon"} />}>
                {"Content"}
            </DecisionFormSectionItem>
        </DecisionFormSection>,
    );

    expect(screen.getByTestId("icon")).toBeInTheDocument();
    expect(screen.getByRole("listitem")).toHaveAttribute("id", "item-example");
    expect(screen.getByText("Content")).toBeInTheDocument();
});

it("renders an item without an icon", () => {
    const { container } = render(
        <DecisionFormSection title={"Permissions"}>
            <DecisionFormSectionItem id={"item-example"}>{"Content"}</DecisionFormSectionItem>
        </DecisionFormSection>,
    );

    expect(container.querySelectorAll("svg")).toHaveLength(0);
    expect(screen.getByText("Content")).toBeInTheDocument();
});

it("renders the list as an item group", () => {
    const { container } = render(
        <DecisionFormSection title={"Permissions"}>
            <li>Example</li>
        </DecisionFormSection>,
    );

    expect(container.querySelector("[data-slot='item-group']")).toBeInTheDocument();
});

it("renders items as list items", () => {
    render(
        <DecisionFormSection title={"Permissions"}>
            <DecisionFormSectionItem id={"item-example"}>{"Content"}</DecisionFormSectionItem>
        </DecisionFormSection>,
    );

    expect(screen.getByRole("listitem")).toHaveAttribute("data-slot", "item");
});

it("renders an item icon as item media", () => {
    const { container } = render(
        <DecisionFormSection title={"Permissions"}>
            <DecisionFormSectionItem id={"item-example"} icon={<span data-testid={"icon"} />}>
                {"Content"}
            </DecisionFormSectionItem>
        </DecisionFormSection>,
    );

    expect(container.querySelector("[data-slot='item-media']")).toContainElement(screen.getByTestId("icon"));
});
