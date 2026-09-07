import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
    Sheet,
    SheetClose,
    SheetContent,
    SheetDescription,
    SheetFooter,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from "@components/UI/Sheet";

function renderSheet(content: React.ReactNode = null, contentProps: Record<string, any> = {}) {
    return render(
        <Sheet defaultOpen>
            <SheetTrigger>Open</SheetTrigger>
            <SheetContent {...contentProps}>
                <SheetHeader>
                    <SheetTitle>Title</SheetTitle>
                    <SheetDescription>Description</SheetDescription>
                </SheetHeader>
                {content}
                <SheetFooter>Footer</SheetFooter>
            </SheetContent>
        </Sheet>,
    );
}

describe("rendering", () => {
    it("renders the title and description when open", async () => {
        renderSheet();

        expect(await screen.findByText("Title")).toHaveAttribute("data-slot", "sheet-title");
        expect(screen.getByText("Description")).toHaveAttribute("data-slot", "sheet-description");
    });

    it("renders the header and footer", async () => {
        renderSheet();

        await screen.findByText("Title");
        expect(document.querySelector('[data-slot="sheet-header"]')).toBeInTheDocument();
        expect(document.querySelector('[data-slot="sheet-footer"]')).toHaveTextContent("Footer");
    });

    it("renders the overlay", async () => {
        renderSheet();

        await screen.findByText("Title");
        expect(document.querySelector('[data-slot="sheet-overlay"]')).toBeInTheDocument();
    });

    it("shows the close button by default", async () => {
        renderSheet();

        expect(await screen.findByText("Close")).toBeInTheDocument();
    });

    it("hides the close button when asked", async () => {
        renderSheet(null, { showCloseButton: false });

        await screen.findByText("Title");
        expect(screen.queryByText("Close")).not.toBeInTheDocument();
    });

    it.each(["left", "right", "top", "bottom"] as const)("renders on the %s side", async (side) => {
        renderSheet(null, { side });

        await screen.findByText("Title");
        expect(document.querySelector('[data-slot="sheet-content"]')).toBeInTheDocument();
    });

    it("merges a custom class name", async () => {
        renderSheet(null, { className: "custom" });

        await screen.findByText("Title");
        expect(document.querySelector('[data-slot="sheet-content"]')).toHaveClass("custom");
    });
});

describe("opening and closing", () => {
    it("stays closed until the trigger is used", () => {
        render(
            <Sheet>
                <SheetTrigger>Open</SheetTrigger>
                <SheetContent>
                    <SheetTitle>Title</SheetTitle>
                </SheetContent>
            </Sheet>,
        );

        expect(screen.queryByText("Title")).not.toBeInTheDocument();
    });

    it("opens on the trigger", async () => {
        render(
            <Sheet>
                <SheetTrigger>Open</SheetTrigger>
                <SheetContent>
                    <SheetTitle>Title</SheetTitle>
                </SheetContent>
            </Sheet>,
        );

        fireEvent.click(screen.getByText("Open"));

        expect(await screen.findByText("Title")).toBeInTheDocument();
    });

    it("closes with an explicit close control", async () => {
        const onOpenChange = vi.fn();

        render(
            <Sheet defaultOpen onOpenChange={onOpenChange}>
                <SheetTrigger>Open</SheetTrigger>
                <SheetContent showCloseButton={false}>
                    <SheetTitle>Title</SheetTitle>
                    <SheetClose>Dismiss</SheetClose>
                </SheetContent>
            </Sheet>,
        );

        fireEvent.click(await screen.findByText("Dismiss"));

        await waitFor(() => expect(onOpenChange).toHaveBeenCalled());
        expect(onOpenChange.mock.calls[0][0]).toBe(false);
    });

    it("closes with the built in close button", async () => {
        const onOpenChange = vi.fn();

        render(
            <Sheet defaultOpen onOpenChange={onOpenChange}>
                <SheetTrigger>Open</SheetTrigger>
                <SheetContent>
                    <SheetTitle>Title</SheetTitle>
                </SheetContent>
            </Sheet>,
        );

        fireEvent.click(await screen.findByText("Close"));

        await waitFor(() => expect(onOpenChange).toHaveBeenCalled());
        expect(onOpenChange.mock.calls[0][0]).toBe(false);
    });
});
