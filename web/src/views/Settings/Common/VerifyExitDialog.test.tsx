import { fireEvent, render, screen } from "@testing-library/react";

import VerifyExitDialog from "@views/Settings/Common/VerifyExitDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

const onCancel = vi.fn();
const onConfirm = vi.fn();

beforeEach(() => {
    onCancel.mockReset();
    onConfirm.mockReset();
});

it("renders the unsaved changes warning", () => {
    render(<VerifyExitDialog open={true} onConfirm={onConfirm} onCancel={onCancel} />);

    expect(screen.getByText("Unsaved Changes")).toBeInTheDocument();
    expect(
        screen.getByText("You have unsaved changes. Are you sure you want to exit without saving?"),
    ).toBeInTheDocument();
});

it("renders nothing when closed", () => {
    render(<VerifyExitDialog open={false} onConfirm={onConfirm} onCancel={onCancel} />);

    expect(screen.queryByText("Unsaved Changes")).not.toBeInTheDocument();
});

it("calls onConfirm when exit without saving is clicked", () => {
    render(<VerifyExitDialog open={true} onConfirm={onConfirm} onCancel={onCancel} />);

    fireEvent.click(screen.getByRole("button", { name: "Exit Without Saving" }));

    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onCancel).not.toHaveBeenCalled();
});

it("calls onCancel when cancel is clicked", () => {
    render(<VerifyExitDialog open={true} onConfirm={onConfirm} onCancel={onCancel} />);

    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onCancel).toHaveBeenCalledOnce();
    expect(onConfirm).not.toHaveBeenCalled();
});
