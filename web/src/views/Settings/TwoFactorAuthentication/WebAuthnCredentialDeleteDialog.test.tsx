import { act, fireEvent, render, screen } from "@testing-library/react";

import WebAuthnCredentialDeleteDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDeleteDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

vi.mock("@services/WebAuthn", () => ({
    deleteUserWebAuthnCredential: vi.fn(),
}));

beforeEach(() => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
});

const credential = { description: "My Key", id: "abc123" } as any;

it("calls handleClose on successful deletion", async () => {
    const { deleteUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(deleteUserWebAuthnCredential).mockResolvedValue({ data: { status: "OK" } } as any);

    const handleClose = vi.fn();
    render(<WebAuthnCredentialDeleteDialog open={true} credential={credential} handleClose={handleClose} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateSuccess).toHaveBeenCalledOnce();
    expect(handleClose).toHaveBeenCalledOnce();
});

it("shows error notification when deletion fails with elevation required", async () => {
    const { deleteUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(deleteUserWebAuthnCredential).mockResolvedValue({ data: { elevation: true, status: "KO" } } as any);

    render(<WebAuthnCredentialDeleteDialog open={true} credential={credential} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateError).toHaveBeenCalledOnce();
});

it("does nothing when credential is undefined", async () => {
    const { deleteUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(deleteUserWebAuthnCredential).mockReset();

    render(<WebAuthnCredentialDeleteDialog open={true} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(deleteUserWebAuthnCredential).not.toHaveBeenCalled();
});
