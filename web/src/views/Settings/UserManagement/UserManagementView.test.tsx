import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";

import { postSendResetPasswordEmailForUser } from "@services/UserManagement";
import UserManagementView from "@views/Settings/UserManagement/UserManagementView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({
        t: (key: string, opts?: any) => (opts ? `${key}:${Object.values(opts).join(",")}` : key),
    }),
}));

const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

const mocks = vi.hoisted(() => ({
    error: undefined as Error | undefined,
    fetchUsers: vi.fn(),
    users: undefined as any[] | undefined,
}));

vi.mock("@hooks/UserManagement", () => ({
    useAllUserInfoGET: () => [mocks.users, mocks.fetchUsers, false, mocks.error],
}));

vi.mock("@services/UserManagement", () => ({
    postSendResetPasswordEmailForUser: vi.fn(),
}));

vi.mock("@views/Settings/UserManagement/EditUserDialog", () => ({
    default: (props: any) => (
        <div data-testid="edit-user-dialog" data-open={String(props.open)} data-username={props.user?.username ?? ""}>
            <button data-testid="edit-user-close" onClick={() => props.onClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/UserManagement/NewUserDialog", () => ({
    default: (props: any) => (
        <div data-testid="new-user-dialog" data-open={String(props.open)}>
            <button data-testid="new-user-close" onClick={() => props.onClose()} />
        </div>
    ),
}));

vi.mock("@views/Settings/UserManagement/VerifyDeleteUserDialog", () => ({
    default: (props: any) => (
        <div data-testid="verify-delete-user-dialog" data-open={String(props.open)} data-username={props.username}>
            <button data-testid="verify-delete-user-close" onClick={() => props.onCancel()} />
        </div>
    ),
}));

vi.mock("@views/Settings/UserManagement/SetPasswordDialog", () => ({
    default: (props: any) => (
        <div data-testid="set-password-dialog" data-open={String(props.open)} data-username={props.username}>
            <button data-testid="set-password-close" onClick={() => props.onCancel()} />
        </div>
    ),
}));

vi.mock("@views/Settings/Common/VerifyActionDialog", () => ({
    default: (props: any) => (
        <div
            data-testid="verify-action-dialog"
            data-open={String(props.open)}
            data-title={props.title}
            data-message={props.message}
            data-confirm={props.confirmText}
        >
            <button data-testid="verify-action-confirm" onClick={() => props.onConfirm()} />
            <button data-testid="verify-action-cancel" onClick={() => props.onCancel()} />
        </div>
    ),
}));

const users = [
    {
        display_name: "John Doe",
        has_duo: false,
        has_totp: true,
        has_webauthn: false,
        last_logged_in: "2024-01-02T03:04:05Z",
        mail: ["john@example.com"],
        method: "totp",
        user_created_at: "2023-12-01T00:00:00Z",
        username: "john",
    },
    {
        display_name: "Alice",
        has_duo: false,
        has_totp: false,
        has_webauthn: false,
        mail: [],
        method: "webauthn",
        username: "alice",
    },
];

const byId = (id: string) => document.getElementById(id) as HTMLElement;
const dialog = (testId: string) => screen.getByTestId(testId);
const rowOf = (username: string) => screen.getByText(username).closest('[role="row"]') as HTMLElement;

const renderView = async () => {
    await act(async () => {
        render(<UserManagementView />);
    });
};

beforeEach(() => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(postSendResetPasswordEmailForUser).mockReset();
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    mocks.error = undefined;
    mocks.fetchUsers.mockReset();
    mocks.users = users;
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("fetches the users on mount and renders the heading", async () => {
    await renderView();

    expect(mocks.fetchUsers).toHaveBeenCalledOnce();
    expect(screen.getByText("User Management")).toBeInTheDocument();
    expect(byId("user-management-add")).toHaveTextContent("Add a {{item}}:user");
});

it("notifies when the users cannot be fetched", async () => {
    mocks.users = undefined;
    mocks.error = new Error("boom");

    await renderView();

    expect(mockCreateError).toHaveBeenCalledWith("There was an issue retrieving user info");
});

it("renders one row per user with the formatted columns", async () => {
    await renderView();

    const john = rowOf("john");
    expect(within(john).getByText("John Doe")).toBeInTheDocument();
    expect(within(john).getByText("john@example.com")).toBeInTheDocument();
    expect(within(john).getByText("TOTP")).toBeInTheDocument();
    expect(within(john).getByText(new Date("2024-01-02T03:04:05Z").toLocaleString())).toBeInTheDocument();
    expect(within(john).getByText(new Date("2023-12-01T00:00:00Z").toLocaleString())).toBeInTheDocument();

    const alice = rowOf("alice");
    expect(within(alice).getByText("Alice")).toBeInTheDocument();
    expect(within(alice).queryByText("TOTP")).not.toBeInTheDocument();
    expect(within(alice).queryByText("WebAuthn")).not.toBeInTheDocument();
    expect(within(alice).getAllByText("-").length).toBeGreaterThanOrEqual(4);
});

it("sorts the rows by username ascending", async () => {
    await renderView();

    const alice = screen.getByText("alice");
    const john = screen.getByText("john");

    expect(alice.compareDocumentPosition(john) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
});

it("renders no rows while the users are not loaded", async () => {
    mocks.users = undefined;

    await renderView();

    expect(screen.queryByText("john")).not.toBeInTheDocument();
    expect(mockCreateError).not.toHaveBeenCalled();
});

it("opens the new user dialog and refetches when it closes", async () => {
    await renderView();

    expect(dialog("new-user-dialog")).toHaveAttribute("data-open", "false");

    fireEvent.click(byId("user-management-add"));
    expect(dialog("new-user-dialog")).toHaveAttribute("data-open", "true");

    fireEvent.click(screen.getByTestId("new-user-close"));
    expect(dialog("new-user-dialog")).toHaveAttribute("data-open", "false");
    expect(mocks.fetchUsers).toHaveBeenCalledTimes(2);
});

it("opens the edit dialog for the row's user from the edit action", async () => {
    await renderView();

    fireEvent.click(byId("user-row-john-edit"));

    expect(dialog("edit-user-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("edit-user-dialog")).toHaveAttribute("data-username", "john");

    fireEvent.click(screen.getByTestId("edit-user-close"));

    expect(dialog("edit-user-dialog")).toHaveAttribute("data-open", "false");
    expect(mocks.fetchUsers).toHaveBeenCalledTimes(2);
});

it("opens the edit dialog when a row is double clicked", async () => {
    await renderView();

    fireEvent.doubleClick(rowOf("alice"));

    expect(dialog("edit-user-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("edit-user-dialog")).toHaveAttribute("data-username", "alice");
});

it("opens the delete confirmation for the row's user from the delete action", async () => {
    await renderView();

    fireEvent.click(byId("user-row-alice-delete"));

    expect(dialog("verify-delete-user-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("verify-delete-user-dialog")).toHaveAttribute("data-username", "alice");

    fireEvent.click(screen.getByTestId("verify-delete-user-close"));

    expect(dialog("verify-delete-user-dialog")).toHaveAttribute("data-open", "false");
    expect(mocks.fetchUsers).toHaveBeenCalledTimes(2);
});

it("offers the extra actions in the row menu", async () => {
    await renderView();

    fireEvent.click(byId("user-row-john-more"));

    expect(screen.getByText("Send Password Reset Email")).toBeInTheDocument();
    expect(screen.getByText("Change Password")).toBeInTheDocument();
    expect(screen.getByText("Edit User")).toBeInTheDocument();
    expect(screen.getByText("Delete User")).toBeInTheDocument();
});

it("sends a password reset email after confirmation", async () => {
    vi.mocked(postSendResetPasswordEmailForUser).mockResolvedValue(undefined);

    await renderView();

    fireEvent.click(byId("user-row-john-more"));
    fireEvent.click(screen.getByText("Send Password Reset Email"));

    expect(dialog("verify-action-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("verify-action-dialog").getAttribute("data-message")).toContain("john");
    expect(postSendResetPasswordEmailForUser).not.toHaveBeenCalled();

    await act(async () => {
        fireEvent.click(screen.getByTestId("verify-action-confirm"));
    });

    expect(postSendResetPasswordEmailForUser).toHaveBeenCalledWith("john");
    expect(mockCreateSuccess).toHaveBeenCalledWith("Password reset email sent successfully");
    await waitFor(() => expect(dialog("verify-action-dialog")).toHaveAttribute("data-open", "false"));
    expect(mocks.fetchUsers).toHaveBeenCalledTimes(2);
});

it("keeps the confirmation open and notifies when the reset email fails", async () => {
    vi.mocked(postSendResetPasswordEmailForUser).mockRejectedValue(new Error("boom"));

    await renderView();

    fireEvent.click(byId("user-row-john-more"));
    fireEvent.click(screen.getByText("Send Password Reset Email"));

    await act(async () => {
        fireEvent.click(screen.getByTestId("verify-action-confirm"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("Error sending password reset email");
    expect(dialog("verify-action-dialog")).toHaveAttribute("data-open", "true");
});

it("cancels the password reset confirmation", async () => {
    await renderView();

    fireEvent.click(byId("user-row-john-more"));
    fireEvent.click(screen.getByText("Send Password Reset Email"));
    fireEvent.click(screen.getByTestId("verify-action-cancel"));

    expect(dialog("verify-action-dialog")).toHaveAttribute("data-open", "false");
    expect(postSendResetPasswordEmailForUser).not.toHaveBeenCalled();
});

it("opens the set password dialog from the row menu", async () => {
    await renderView();

    fireEvent.click(byId("user-row-alice-more"));
    fireEvent.click(screen.getByText("Change Password"));

    expect(dialog("set-password-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("set-password-dialog")).toHaveAttribute("data-username", "alice");

    fireEvent.click(screen.getByTestId("set-password-close"));

    expect(dialog("set-password-dialog")).toHaveAttribute("data-open", "false");
    expect(mocks.fetchUsers).toHaveBeenCalledTimes(2);
});

it("opens the edit dialog from the row menu", async () => {
    await renderView();

    fireEvent.click(byId("user-row-alice-more"));
    fireEvent.click(screen.getByText("Edit User"));

    expect(dialog("edit-user-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("edit-user-dialog")).toHaveAttribute("data-username", "alice");
});

it("opens the delete confirmation from the row menu", async () => {
    await renderView();

    fireEvent.click(byId("user-row-john-more"));
    fireEvent.click(screen.getByText("Delete User"));

    expect(dialog("verify-delete-user-dialog")).toHaveAttribute("data-open", "true");
    expect(dialog("verify-delete-user-dialog")).toHaveAttribute("data-username", "john");
});
