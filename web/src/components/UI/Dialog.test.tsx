import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@components/UI/Dialog";

describe("trigger", () => {
    it("renders the trigger", () => {
        render(
            <Dialog>
                <DialogTrigger>Open</DialogTrigger>
                <DialogContent>
                    <DialogTitle>Title</DialogTitle>
                </DialogContent>
            </Dialog>,
        );

        expect(screen.getByText("Open")).toHaveAttribute("data-slot", "dialog-trigger");
        expect(screen.queryByText("Title")).not.toBeInTheDocument();
    });

    it("opens the dialog", async () => {
        render(
            <Dialog>
                <DialogTrigger>Open</DialogTrigger>
                <DialogContent>
                    <DialogTitle>Title</DialogTitle>
                </DialogContent>
            </Dialog>,
        );

        fireEvent.click(screen.getByText("Open"));

        expect(await screen.findByText("Title")).toBeInTheDocument();
    });
});

describe("content", () => {
    it("renders the overlay, title and description", async () => {
        render(
            <Dialog defaultOpen>
                <DialogContent className="custom">
                    <DialogHeader className="header">
                        <DialogTitle>Title</DialogTitle>
                        <DialogDescription>Description</DialogDescription>
                    </DialogHeader>
                </DialogContent>
            </Dialog>,
        );

        expect(await screen.findByText("Title")).toHaveAttribute("data-slot", "dialog-title");
        expect(screen.getByText("Description")).toHaveAttribute("data-slot", "dialog-description");
        expect(document.querySelector('[data-slot="dialog-overlay"]')).toBeInTheDocument();
        expect(document.querySelector('[data-slot="dialog-content"]')).toHaveClass("custom");
        expect(document.querySelector('[data-slot="dialog-header"]')).toHaveClass("header");
    });

    it("shows the corner close button by default", async () => {
        render(
            <Dialog defaultOpen>
                <DialogContent>
                    <DialogTitle>Title</DialogTitle>
                </DialogContent>
            </Dialog>,
        );

        await screen.findByText("Title");
        expect(document.querySelector('[data-slot="dialog-close"]')).toBeInTheDocument();
    });

    it("hides the corner close button when asked", async () => {
        render(
            <Dialog defaultOpen>
                <DialogContent showCloseButton={false}>
                    <DialogTitle>Title</DialogTitle>
                </DialogContent>
            </Dialog>,
        );

        await screen.findByText("Title");
        expect(document.querySelector('[data-slot="dialog-close"]')).toBeNull();
    });

    it("closes with the corner close button", async () => {
        const onOpenChange = vi.fn();

        render(
            <Dialog defaultOpen onOpenChange={onOpenChange}>
                <DialogContent>
                    <DialogTitle>Title</DialogTitle>
                </DialogContent>
            </Dialog>,
        );

        await screen.findByText("Title");
        fireEvent.click(document.querySelector('[data-slot="dialog-close"]') as HTMLElement);

        await waitFor(() => expect(onOpenChange).toHaveBeenCalled());
        expect(onOpenChange.mock.calls[0][0]).toBe(false);
    });
});

describe("footer", () => {
    it("omits the close button by default", async () => {
        render(
            <Dialog defaultOpen>
                <DialogContent showCloseButton={false}>
                    <DialogTitle>Title</DialogTitle>
                    <DialogFooter className="footer">Actions</DialogFooter>
                </DialogContent>
            </Dialog>,
        );

        await screen.findByText("Title");
        expect(document.querySelector('[data-slot="dialog-footer"]')).toHaveClass("footer");
        expect(screen.queryByText("Close")).not.toBeInTheDocument();
    });

    it("renders a close button when asked", async () => {
        render(
            <Dialog defaultOpen>
                <DialogContent showCloseButton={false}>
                    <DialogTitle>Title</DialogTitle>
                    <DialogFooter showCloseButton>Actions</DialogFooter>
                </DialogContent>
            </Dialog>,
        );

        expect(await screen.findByText("Close")).toBeInTheDocument();
    });
});

describe("explicit close", () => {
    it("closes the dialog", async () => {
        const onOpenChange = vi.fn();

        render(
            <Dialog defaultOpen onOpenChange={onOpenChange}>
                <DialogContent showCloseButton={false}>
                    <DialogTitle>Title</DialogTitle>
                    <DialogClose>Dismiss</DialogClose>
                </DialogContent>
            </Dialog>,
        );

        fireEvent.click(await screen.findByText("Dismiss"));

        await waitFor(() => expect(onOpenChange).toHaveBeenCalled());
        expect(onOpenChange.mock.calls[0][0]).toBe(false);
    });
});
