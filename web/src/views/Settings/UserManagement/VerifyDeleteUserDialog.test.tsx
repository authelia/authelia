import { act, fireEvent, render, screen } from "@testing-library/react";

import { deleteUser } from "@services/UserManagement";
import VerifyDeleteUserDialog from "@views/Settings/UserManagement/VerifyDeleteUserDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string, opts?: any) => (opts?.item ? `${key}:${opts.item}` : key) }),
}));

const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

vi.mock("@services/UserManagement", () => ({
    deleteUser: vi.fn(),
}));

const onCancel = vi.fn();

const input = () => document.getElementById("verify-delete-user-input") as HTMLInputElement;
const confirmButton = () => document.getElementById("verify-delete-user-confirm") as HTMLButtonElement;

beforeEach(() => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.mocked(deleteUser).mockReset();
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    onCancel.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the title and the username in the message", () => {
    render(<VerifyDeleteUserDialog open={true} username="john" onCancel={onCancel} />);

    expect(screen.getAllByText("Delete User").length).toBeGreaterThanOrEqual(1);
    expect(
        screen.getByText("You are about to delete user {{item}}, enter their username to continue.:john"),
    ).toBeInTheDocument();
});

it("keeps the delete button disabled until the exact username is typed", () => {
    render(<VerifyDeleteUserDialog open={true} username="john" onCancel={onCancel} />);

    expect(confirmButton()).toBeDisabled();

    fireEvent.change(input(), { target: { value: "joh" } });
    expect(confirmButton()).toBeDisabled();

    fireEvent.change(input(), { target: { value: "John" } });
    expect(confirmButton()).toBeDisabled();

    fireEvent.change(input(), { target: { value: "john" } });
    expect(confirmButton()).toBeEnabled();

    fireEvent.change(input(), { target: { value: "johnn" } });
    expect(confirmButton()).toBeDisabled();
});

it("deletes the user, notifies and closes on success", async () => {
    vi.mocked(deleteUser).mockResolvedValue(undefined);

    render(<VerifyDeleteUserDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(input(), { target: { value: "john" } });

    await act(async () => {
        fireEvent.click(confirmButton());
    });

    expect(deleteUser).toHaveBeenCalledWith("john");
    expect(mockCreateSuccess).toHaveBeenCalledWith("User deleted successfully.");
    expect(mockCreateError).not.toHaveBeenCalled();
    expect(onCancel).toHaveBeenCalledOnce();
});

it("notifies an error and still closes when deletion fails", async () => {
    vi.mocked(deleteUser).mockRejectedValue(new Error("boom"));

    render(<VerifyDeleteUserDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(input(), { target: { value: "john" } });

    await act(async () => {
        fireEvent.click(confirmButton());
    });

    expect(mockCreateError).toHaveBeenCalledWith("Error deleting user.");
    expect(mockCreateSuccess).not.toHaveBeenCalled();
    expect(onCancel).toHaveBeenCalledOnce();
});

it("closes without deleting when cancel is clicked", () => {
    render(<VerifyDeleteUserDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(input(), { target: { value: "john" } });
    fireEvent.click(document.getElementById("verify-delete-user-cancel")!);

    expect(deleteUser).not.toHaveBeenCalled();
    expect(onCancel).toHaveBeenCalledOnce();
});

it("resets the typed name when the dialog is closed", () => {
    const { rerender } = render(<VerifyDeleteUserDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(input(), { target: { value: "john" } });
    fireEvent.click(document.getElementById("verify-delete-user-cancel")!);

    rerender(<VerifyDeleteUserDialog open={true} username="john" onCancel={onCancel} />);

    expect(input().value).toBe("");
    expect(confirmButton()).toBeDisabled();
});
