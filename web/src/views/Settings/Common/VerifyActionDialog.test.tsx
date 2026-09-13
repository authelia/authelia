import { fireEvent, render, screen } from "@testing-library/react";

import VerifyActionDialog from "@views/Settings/Common/VerifyActionDialog";

const defaultProps = {
    cancelText: "Keep",
    confirmText: "Do it",
    message: "Are you sure you want to do it?",
    onCancel: vi.fn(),
    onConfirm: vi.fn(),
    open: true,
    title: "Confirm action",
};

beforeEach(() => {
    defaultProps.onCancel.mockReset();
    defaultProps.onConfirm.mockReset();
});

it("renders the title, message and button labels", () => {
    render(<VerifyActionDialog {...defaultProps} />);

    expect(screen.getByText("Confirm action")).toBeInTheDocument();
    expect(screen.getByText("Are you sure you want to do it?")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Keep" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Do it" })).toBeInTheDocument();
});

it("renders nothing when closed", () => {
    render(<VerifyActionDialog {...defaultProps} open={false} />);

    expect(screen.queryByText("Confirm action")).not.toBeInTheDocument();
});

it("calls onConfirm when the confirm button is clicked", () => {
    render(<VerifyActionDialog {...defaultProps} />);

    fireEvent.click(screen.getByRole("button", { name: "Do it" }));

    expect(defaultProps.onConfirm).toHaveBeenCalledOnce();
    expect(defaultProps.onCancel).not.toHaveBeenCalled();
});

it("calls onCancel when the cancel button is clicked", () => {
    render(<VerifyActionDialog {...defaultProps} />);

    fireEvent.click(screen.getByRole("button", { name: "Keep" }));

    expect(defaultProps.onCancel).toHaveBeenCalledOnce();
    expect(defaultProps.onConfirm).not.toHaveBeenCalled();
});
