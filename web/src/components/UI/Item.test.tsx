import { render, screen } from "@testing-library/react";

import {
    Item,
    ItemActions,
    ItemContent,
    ItemDescription,
    ItemFooter,
    ItemGroup,
    ItemHeader,
    ItemMedia,
    ItemSeparator,
    ItemTitle,
} from "@components/UI/Item";

describe("Item", () => {
    it("defaults to the default size and variant", () => {
        const { container } = render(<Item>Body</Item>);

        const el = container.querySelector('[data-slot="item"]');
        expect(el).toHaveAttribute("data-size", "default");
        expect(el).toHaveAttribute("data-variant", "default");
    });

    it.each(["default", "sm"] as const)("supports the %s size", (size) => {
        const { container } = render(<Item size={size}>Body</Item>);
        expect(container.querySelector('[data-slot="item"]')).toHaveAttribute("data-size", size);
    });

    it.each(["default", "muted", "outline"] as const)("supports the %s variant", (variant) => {
        const { container } = render(<Item variant={variant}>Body</Item>);
        expect(container.querySelector('[data-slot="item"]')).toHaveAttribute("data-variant", variant);
    });

    it("merges a custom class name", () => {
        const { container } = render(<Item className="custom">Body</Item>);
        expect(container.querySelector('[data-slot="item"]')).toHaveClass("custom");
    });

    it("renders through a custom element", () => {
        const { container } = render(<Item render={<a href="https://example.com" />}>Link</Item>);
        expect(container.querySelector("a")).toHaveAttribute("href", "https://example.com");
    });
});

describe("ItemGroup", () => {
    it("renders a group with a separator", () => {
        const { container } = render(
            <ItemGroup className="group">
                <Item>One</Item>
                <ItemSeparator className="separator" />
                <Item>Two</Item>
            </ItemGroup>,
        );

        expect(container.querySelector('[data-slot="item-group"]')).toHaveClass("group");
        expect(container.querySelector('[data-slot="item-separator"]')).toHaveClass("separator");
    });
});

describe("Item parts", () => {
    it("renders the media, content, title and description", () => {
        const { container } = render(
            <Item>
                <ItemMedia className="media">icon</ItemMedia>
                <ItemContent className="content">
                    <ItemTitle>Title</ItemTitle>
                    <ItemDescription>Description</ItemDescription>
                </ItemContent>
                <ItemActions className="actions">action</ItemActions>
            </Item>,
        );

        expect(container.querySelector('[data-slot="item-media"]')).toHaveClass("media");
        expect(container.querySelector('[data-slot="item-content"]')).toHaveClass("content");
        expect(screen.getByText("Title")).toHaveAttribute("data-slot", "item-title");
        expect(screen.getByText("Description")).toHaveAttribute("data-slot", "item-description");
        expect(container.querySelector('[data-slot="item-actions"]')).toHaveClass("actions");
    });

    it.each(["default", "icon", "image"] as const)("supports the %s media variant", (variant) => {
        const { container } = render(
            <Item>
                <ItemMedia variant={variant}>icon</ItemMedia>
            </Item>,
        );

        expect(container.querySelector('[data-slot="item-media"]')).toHaveAttribute("data-variant", variant);
    });

    it("renders the header and footer", () => {
        const { container } = render(
            <Item>
                <ItemHeader className="header">Header</ItemHeader>
                <ItemContent>Body</ItemContent>
                <ItemFooter className="footer">Footer</ItemFooter>
            </Item>,
        );

        expect(container.querySelector('[data-slot="item-header"]')).toHaveClass("header");
        expect(container.querySelector('[data-slot="item-footer"]')).toHaveClass("footer");
    });
});
