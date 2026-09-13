import { render, screen } from "@testing-library/react";

import { Dialog, DialogContent, DialogTitle } from "@components/UI/Dialog";

it("renders the dialog content with the dialog-content data-slot", () => {
    render(
        <Dialog open modal={false}>
            <DialogContent>
                <DialogTitle>Title</DialogTitle>
            </DialogContent>
        </Dialog>,
    );
    expect(screen.getByText("Title").closest("[data-slot='dialog-content']")).toBeInTheDocument();
});

it("caps the dialog content height and scrolls instead of overflowing the viewport", () => {
    render(
        <Dialog open modal={false}>
            <DialogContent>
                <DialogTitle>Title</DialogTitle>
            </DialogContent>
        </Dialog>,
    );
    const content = screen.getByText("Title").closest("[data-slot='dialog-content']");
    expect(content).toHaveClass("max-h-[85vh]");
    expect(content).toHaveClass("overflow-y-auto");
});

it("merges a passed className onto the dialog content", () => {
    render(
        <Dialog open modal={false}>
            <DialogContent className="custom-dialog">
                <DialogTitle>Title</DialogTitle>
            </DialogContent>
        </Dialog>,
    );
    expect(screen.getByText("Title").closest("[data-slot='dialog-content']")).toHaveClass("custom-dialog");
});
