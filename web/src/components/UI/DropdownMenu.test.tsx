import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
    DropdownMenuSeparator,
    DropdownMenuShortcut,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger,
} from "@components/UI/DropdownMenu";

function renderMenu(children: React.ReactNode, open = true) {
    return render(
        <DropdownMenu defaultOpen={open}>
            <DropdownMenuTrigger>Open</DropdownMenuTrigger>
            <DropdownMenuContent>{children}</DropdownMenuContent>
        </DropdownMenu>,
    );
}

describe("trigger", () => {
    it("renders the trigger", () => {
        render(
            <DropdownMenu>
                <DropdownMenuTrigger>Open</DropdownMenuTrigger>
                <DropdownMenuContent>
                    <DropdownMenuItem>Item</DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>,
        );

        expect(screen.getByText("Open")).toHaveAttribute("data-slot", "dropdown-menu-trigger");
    });

    it("keeps the content closed until the trigger is used", () => {
        render(
            <DropdownMenu>
                <DropdownMenuTrigger>Open</DropdownMenuTrigger>
                <DropdownMenuContent>
                    <DropdownMenuItem>Item</DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>,
        );

        expect(screen.queryByText("Item")).not.toBeInTheDocument();
    });

    it("opens the content on click", async () => {
        render(
            <DropdownMenu>
                <DropdownMenuTrigger>Open</DropdownMenuTrigger>
                <DropdownMenuContent>
                    <DropdownMenuItem>Item</DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>,
        );

        fireEvent.click(screen.getByText("Open"));

        expect(await screen.findByText("Item")).toBeInTheDocument();
    });
});

describe("content", () => {
    it("renders the popup when open", async () => {
        renderMenu(<DropdownMenuItem>Item</DropdownMenuItem>);

        await waitFor(() => expect(document.querySelector('[data-slot="dropdown-menu-content"]')).toBeInTheDocument());
    });

    it("merges a custom class name", async () => {
        render(
            <DropdownMenu defaultOpen>
                <DropdownMenuTrigger>Open</DropdownMenuTrigger>
                <DropdownMenuContent className="custom" align="start" side="top" sideOffset={8}>
                    <DropdownMenuItem>Item</DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>,
        );

        await waitFor(() =>
            expect(document.querySelector('[data-slot="dropdown-menu-content"]')).toHaveClass("custom"),
        );
    });
});

describe("items", () => {
    it("renders a default item", async () => {
        renderMenu(<DropdownMenuItem>Item</DropdownMenuItem>);

        const item = await screen.findByText("Item");
        expect(item).toHaveAttribute("data-variant", "default");
    });

    it("renders a destructive item", async () => {
        renderMenu(<DropdownMenuItem variant="destructive">Delete</DropdownMenuItem>);

        expect(await screen.findByText("Delete")).toHaveAttribute("data-variant", "destructive");
    });

    it("renders an inset item", async () => {
        renderMenu(<DropdownMenuItem inset>Inset</DropdownMenuItem>);

        expect(await screen.findByText("Inset")).toHaveAttribute("data-inset", "true");
    });

    it("fires the click handler", async () => {
        const onClick = vi.fn();

        renderMenu(<DropdownMenuItem onClick={onClick}>Item</DropdownMenuItem>);

        fireEvent.click(await screen.findByText("Item"));

        expect(onClick).toHaveBeenCalled();
    });

    it("renders a group with a label", async () => {
        renderMenu(
            <DropdownMenuGroup>
                <DropdownMenuLabel>Section</DropdownMenuLabel>
                <DropdownMenuItem>Item</DropdownMenuItem>
            </DropdownMenuGroup>,
        );

        expect(await screen.findByText("Section")).toHaveAttribute("data-slot", "dropdown-menu-label");
        expect(document.querySelector('[data-slot="dropdown-menu-group"]')).toBeInTheDocument();
    });

    it("renders an inset label", async () => {
        renderMenu(
            <DropdownMenuGroup>
                <DropdownMenuLabel inset>Section</DropdownMenuLabel>
                <DropdownMenuItem>Item</DropdownMenuItem>
            </DropdownMenuGroup>,
        );

        expect(await screen.findByText("Section")).toHaveAttribute("data-inset", "true");
    });
});

describe("checkbox items", () => {
    it("renders a checked item", async () => {
        renderMenu(<DropdownMenuCheckboxItem checked>Checked</DropdownMenuCheckboxItem>);

        const item = await screen.findByText("Checked");
        expect(item).toHaveAttribute("data-slot", "dropdown-menu-checkbox-item");
        expect(item).toHaveAttribute("aria-checked", "true");
    });

    it("renders an unchecked item", async () => {
        renderMenu(<DropdownMenuCheckboxItem checked={false}>Unchecked</DropdownMenuCheckboxItem>);

        expect(await screen.findByText("Unchecked")).toHaveAttribute("aria-checked", "false");
    });

    it("reports a change", async () => {
        const onCheckedChange = vi.fn();

        renderMenu(
            <DropdownMenuCheckboxItem checked={false} onCheckedChange={onCheckedChange}>
                Toggle
            </DropdownMenuCheckboxItem>,
        );

        fireEvent.click(await screen.findByText("Toggle"));

        await waitFor(() => expect(onCheckedChange).toHaveBeenCalledWith(true, expect.anything()));
    });
});

describe("radio items", () => {
    it("renders the selected item", async () => {
        renderMenu(
            <DropdownMenuRadioGroup value="one">
                <DropdownMenuRadioItem value="one">One</DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="two">Two</DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>,
        );

        expect(await screen.findByText("One")).toHaveAttribute("aria-checked", "true");
        expect(screen.getByText("Two")).toHaveAttribute("aria-checked", "false");
        expect(document.querySelector('[data-slot="dropdown-menu-radio-group"]')).toBeInTheDocument();
    });

    it("reports a change", async () => {
        const onValueChange = vi.fn();

        renderMenu(
            <DropdownMenuRadioGroup value="one" onValueChange={onValueChange}>
                <DropdownMenuRadioItem value="one">One</DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="two">Two</DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>,
        );

        fireEvent.click(await screen.findByText("Two"));

        await waitFor(() => expect(onValueChange).toHaveBeenCalledWith("two", expect.anything()));
    });
});

describe("decorations", () => {
    it("renders a separator", async () => {
        renderMenu(
            <>
                <DropdownMenuItem>Item</DropdownMenuItem>
                <DropdownMenuSeparator />
            </>,
        );

        await screen.findByText("Item");
        expect(document.querySelector('[data-slot="dropdown-menu-separator"]')).toHaveAttribute("role", "separator");
    });

    it("renders a shortcut", async () => {
        renderMenu(
            <DropdownMenuItem>
                Item
                <DropdownMenuShortcut>⌘K</DropdownMenuShortcut>
            </DropdownMenuItem>,
        );

        expect(await screen.findByText("⌘K")).toHaveAttribute("data-slot", "dropdown-menu-shortcut");
    });

    it("merges a custom class name on the separator", async () => {
        renderMenu(
            <>
                <DropdownMenuItem>Item</DropdownMenuItem>
                <DropdownMenuSeparator className="custom" />
            </>,
        );

        await screen.findByText("Item");
        expect(document.querySelector('[data-slot="dropdown-menu-separator"]')).toHaveClass("custom");
    });
});

describe("submenus", () => {
    it("renders a submenu trigger", async () => {
        renderMenu(
            <DropdownMenuSub>
                <DropdownMenuSubTrigger>More</DropdownMenuSubTrigger>
                <DropdownMenuSubContent>
                    <DropdownMenuItem>Nested</DropdownMenuItem>
                </DropdownMenuSubContent>
            </DropdownMenuSub>,
        );

        expect(await screen.findByText("More")).toHaveAttribute("data-slot", "dropdown-menu-sub-trigger");
    });

    it("renders an inset submenu trigger", async () => {
        renderMenu(
            <DropdownMenuSub>
                <DropdownMenuSubTrigger inset>More</DropdownMenuSubTrigger>
                <DropdownMenuSubContent>
                    <DropdownMenuItem>Nested</DropdownMenuItem>
                </DropdownMenuSubContent>
            </DropdownMenuSub>,
        );

        expect(await screen.findByText("More")).toHaveAttribute("data-inset", "true");
    });

    it("opens the submenu content", async () => {
        renderMenu(
            <DropdownMenuSub defaultOpen>
                <DropdownMenuSubTrigger>More</DropdownMenuSubTrigger>
                <DropdownMenuSubContent className="custom">
                    <DropdownMenuItem>Nested</DropdownMenuItem>
                </DropdownMenuSubContent>
            </DropdownMenuSub>,
        );

        expect(await screen.findByText("Nested")).toBeInTheDocument();
        expect(document.querySelector('[data-slot="dropdown-menu-sub-content"]')).toHaveClass("custom");
    });
});
