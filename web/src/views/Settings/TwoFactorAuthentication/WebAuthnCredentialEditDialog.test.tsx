import { act, fireEvent, render, screen } from "@testing-library/react";

import WebAuthnCredentialEditDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialEditDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

vi.mock("@services/WebAuthn", () => ({
    updateUserWebAuthnCredential: vi.fn(),
}));

beforeEach(() => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

const credential = { description: "My Key", id: "abc123" } as any;

it("renders the edit dialog when open", () => {
    render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);
    expect(screen.getByText("Edit WebAuthn Credential")).toBeInTheDocument();
});

it("does not call update when description is empty", async () => {
    const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(updateUserWebAuthnCredential).mockReset();

    const handleClose = vi.fn();
    render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={handleClose} />);

    fireEvent.click(screen.getByText("Update"));

    expect(updateUserWebAuthnCredential).not.toHaveBeenCalled();
    expect(handleClose).not.toHaveBeenCalled();
});

it("calls handleClose on successful update", async () => {
    const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({ data: { status: "OK" } } as any);

    const handleClose = vi.fn();
    render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={handleClose} />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Description *"), { target: { value: "New Name" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Update"));
    });

    expect(handleClose).toHaveBeenCalledOnce();
});

it("calls handleClose when cancel is clicked", () => {
    const handleClose = vi.fn();
    render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={handleClose} />);
    fireEvent.click(screen.getByText("Cancel"));
    expect(handleClose).toHaveBeenCalledOnce();
});

it("shows error when credential is undefined", async () => {
    const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(updateUserWebAuthnCredential).mockReset();

    render(<WebAuthnCredentialEditDialog open={true} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Description *"), { target: { value: "New Name" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Update"));
    });

    expect(mockCreateError).toHaveBeenCalledOnce();
});
