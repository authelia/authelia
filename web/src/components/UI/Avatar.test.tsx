import { render, screen } from "@testing-library/react";

import { Avatar, AvatarBadge, AvatarFallback, AvatarGroup, AvatarGroupCount, AvatarImage } from "@components/UI/Avatar";

describe("Avatar", () => {
    it("defaults to the default size", () => {
        const { container } = render(
            <Avatar>
                <AvatarFallback>JD</AvatarFallback>
            </Avatar>,
        );

        expect(container.querySelector('[data-slot="avatar"]')).toHaveAttribute("data-size", "default");
    });

    it.each(["default", "lg", "sm"] as const)("supports the %s size", (size) => {
        const { container } = render(
            <Avatar size={size}>
                <AvatarFallback>JD</AvatarFallback>
            </Avatar>,
        );

        expect(container.querySelector('[data-slot="avatar"]')).toHaveAttribute("data-size", size);
    });

    it("merges a custom class name", () => {
        const { container } = render(
            <Avatar className="custom">
                <AvatarFallback>JD</AvatarFallback>
            </Avatar>,
        );

        expect(container.querySelector('[data-slot="avatar"]')).toHaveClass("custom");
    });
});

describe("AvatarFallback", () => {
    it("renders the fallback content", () => {
        render(
            <Avatar>
                <AvatarFallback>JD</AvatarFallback>
            </Avatar>,
        );

        expect(screen.getByText("JD")).toHaveAttribute("data-slot", "avatar-fallback");
    });
});

describe("AvatarImage", () => {
    it("renders inside an avatar", () => {
        const { container } = render(
            <Avatar>
                <AvatarImage src="https://example.com/avatar.png" alt="John" />
                <AvatarFallback>JD</AvatarFallback>
            </Avatar>,
        );

        expect(container.querySelector('[data-slot="avatar"]')).toBeInTheDocument();
    });
});

describe("AvatarBadge", () => {
    it("renders a badge", () => {
        const { container } = render(
            <Avatar>
                <AvatarFallback>JD</AvatarFallback>
                <AvatarBadge className="custom" />
            </Avatar>,
        );

        const badge = container.querySelector('[data-slot="avatar-badge"]');
        expect(badge).toBeInTheDocument();
        expect(badge).toHaveClass("custom");
    });
});

describe("AvatarGroup", () => {
    it("renders a group with a count", () => {
        const { container } = render(
            <AvatarGroup className="custom">
                <Avatar>
                    <AvatarFallback>A</AvatarFallback>
                </Avatar>
                <Avatar>
                    <AvatarFallback>B</AvatarFallback>
                </Avatar>
                <AvatarGroupCount>+3</AvatarGroupCount>
            </AvatarGroup>,
        );

        expect(container.querySelector('[data-slot="avatar-group"]')).toHaveClass("custom");
        expect(screen.getByText("+3")).toHaveAttribute("data-slot", "avatar-group-count");
    });
});
