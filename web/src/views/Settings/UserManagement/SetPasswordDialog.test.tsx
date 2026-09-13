import { act, fireEvent, render, screen } from "@testing-library/react";

import { postChangePasswordForUser } from "@services/UserManagement";
import SetUserPasswordDialog from "@views/Settings/UserManagement/SetPasswordDialog";

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
    postChangePasswordForUser: vi.fn(),
}));

const onCancel = vi.fn();

const password = () => document.getElementById("set-password-password") as HTMLInputElement;
const confirm = () => document.getElementById("set-password-confirm") as HTMLInputElement;
const submit = () => document.getElementById("set-password-submit") as HTMLButtonElement;
const cancel = () => document.getElementById("set-password-cancel") as HTMLButtonElement;

beforeEach(() => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(postChangePasswordForUser).mockReset();
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    onCancel.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the title and the username in the message", () => {
    render(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);

    expect(screen.getByText("Set User Password")).toBeInTheDocument();
    expect(screen.getByText("Set a new password for user {{item}}.:john")).toBeInTheDocument();
    expect(password()).toHaveAttribute("type", "password");
    expect(confirm()).toHaveAttribute("type", "password");
});

it("keeps submit disabled until both passwords are non-empty and match", () => {
    render(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);

    expect(submit()).toBeDisabled();

    fireEvent.change(password(), { target: { value: "secret1" } });
    expect(submit()).toBeDisabled();

    fireEvent.change(confirm(), { target: { value: "secret2" } });
    expect(submit()).toBeDisabled();

    fireEvent.change(confirm(), { target: { value: "secret1" } });
    expect(submit()).toBeEnabled();
});

it("only shows the mismatch message once the confirmation is non-empty", () => {
    render(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(password(), { target: { value: "secret1" } });
    expect(screen.queryByText("Passwords do not match")).not.toBeInTheDocument();

    fireEvent.change(confirm(), { target: { value: "x" } });
    expect(screen.getByText("Passwords do not match")).toBeInTheDocument();

    fireEvent.change(confirm(), { target: { value: "secret1" } });
    expect(screen.queryByText("Passwords do not match")).not.toBeInTheDocument();
});

it("submits the password, notifies and closes on success", async () => {
    vi.mocked(postChangePasswordForUser).mockResolvedValue(undefined);

    render(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(password(), { target: { value: "secret1" } });
    fireEvent.change(confirm(), { target: { value: "secret1" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    expect(postChangePasswordForUser).toHaveBeenCalledWith("john", "secret1");
    expect(mockCreateSuccess).toHaveBeenCalledWith("Password updated successfully.");
    expect(onCancel).toHaveBeenCalledOnce();
});

it("notifies an error and keeps the dialog open with its values on failure", async () => {
    vi.mocked(postChangePasswordForUser).mockRejectedValue(new Error("boom"));

    render(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(password(), { target: { value: "secret1" } });
    fireEvent.change(confirm(), { target: { value: "secret1" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    expect(mockCreateError).toHaveBeenCalledWith("Error updating password.");
    expect(onCancel).not.toHaveBeenCalled();
    expect(password().value).toBe("secret1");
    expect(confirm().value).toBe("secret1");
    expect(submit()).toBeEnabled();
});

it("disables the inputs and buttons while the request is in flight", async () => {
    let resolve: () => void = () => {};
    vi.mocked(postChangePasswordForUser).mockReturnValue(
        new Promise<undefined>((r) => {
            resolve = () => r(undefined);
        }),
    );

    render(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(password(), { target: { value: "secret1" } });
    fireEvent.change(confirm(), { target: { value: "secret1" } });

    await act(async () => {
        fireEvent.click(submit());
    });

    expect(password()).toBeDisabled();
    expect(confirm()).toBeDisabled();
    expect(submit()).toBeDisabled();
    expect(cancel()).toBeDisabled();

    await act(async () => {
        resolve();
    });

    expect(onCancel).toHaveBeenCalledOnce();
});

it("clears the fields and closes when cancel is clicked", () => {
    const { rerender } = render(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);

    fireEvent.change(password(), { target: { value: "secret1" } });
    fireEvent.click(cancel());

    expect(onCancel).toHaveBeenCalledOnce();
    expect(postChangePasswordForUser).not.toHaveBeenCalled();

    rerender(<SetUserPasswordDialog open={true} username="john" onCancel={onCancel} />);
    expect(password().value).toBe("");
});
