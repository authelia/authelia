// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen } from "@testing-library/react";

import WebAuthnCredentialDeleteDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDeleteDialog";

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

it("reports that a higher authentication level is required", async () => {
    const { deleteUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(deleteUserWebAuthnCredential).mockResolvedValue({
        data: { authentication: true, elevation: false, status: "KO" },
    } as any);

    const handleClose = vi.fn();
    render(<WebAuthnCredentialDeleteDialog open={true} credential={credential} handleClose={handleClose} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateError).toHaveBeenCalledWith(
        "You must have a higher authentication level to {{action}} a {{item}}",
    );
    expect(handleClose).not.toHaveBeenCalled();
});

it("reports a generic failure", async () => {
    const { deleteUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(deleteUserWebAuthnCredential).mockResolvedValue({
        data: { authentication: false, elevation: false, status: "KO" },
    } as any);

    render(<WebAuthnCredentialDeleteDialog open={true} credential={credential} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("There was a problem {{action}} the {{item}}");
});

it("sends the credential id", async () => {
    const { deleteUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(deleteUserWebAuthnCredential).mockResolvedValue({ data: { status: "OK" } } as any);

    render(<WebAuthnCredentialDeleteDialog open={true} credential={credential} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(deleteUserWebAuthnCredential).toHaveBeenCalledWith("abc123");
});

it("closes when cancelled", () => {
    const handleClose = vi.fn();
    render(<WebAuthnCredentialDeleteDialog open={true} credential={credential} handleClose={handleClose} />);

    fireEvent.click(screen.getByText("Cancel"));

    expect(handleClose).toHaveBeenCalledOnce();
});

it("logs a rejected deletion", async () => {
    const { deleteUserWebAuthnCredential } = await import("@services/WebAuthn");
    vi.mocked(deleteUserWebAuthnCredential).mockRejectedValue(new Error("boom"));

    render(<WebAuthnCredentialDeleteDialog open={true} credential={credential} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(console.error).toHaveBeenCalled();
});
