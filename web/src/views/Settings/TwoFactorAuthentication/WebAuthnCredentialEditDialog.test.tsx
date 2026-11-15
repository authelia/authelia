// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

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
    vi.clearAllMocks();
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
        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "New Name" } });
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
    render(<WebAuthnCredentialEditDialog open={true} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "New Name" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Update"));
    });

    expect(mockCreateError).toHaveBeenCalledOnce();
});

describe("update failures", () => {
    async function update(value = "New Name") {
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);

        await act(async () => {
            fireEvent.change(screen.getByLabelText("Description"), { target: { value } });
        });

        await act(async () => {
            fireEvent.click(screen.getByText("Update"));
        });
    }

    it("reports an empty response", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue(null as any);

        await update();

        expect(mockCreateError).toHaveBeenCalledWith(
            "An error occurred when attempting to update the WebAuthn Credential",
        );
    });

    it("reports that elevation is required", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({
            data: { elevation: true, status: "KO" },
        } as any);

        await update();

        expect(mockCreateError).toHaveBeenCalledWith("You must be elevated to {{action}} a {{item}}");
    });

    it("reports that a higher authentication level is required", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({
            data: { authentication: true, elevation: false, status: "KO" },
        } as any);

        await update();

        expect(mockCreateError).toHaveBeenCalledWith(
            "You must have a higher authentication level to {{action}} a {{item}}",
        );
    });

    it("reports a generic failure", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({
            data: { authentication: false, elevation: false, status: "KO" },
        } as any);

        await update();

        expect(mockCreateError).toHaveBeenCalledWith("There was a problem {{action}} the {{item}}");
    });

    it("reports success", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({ data: { status: "OK" } } as any);

        await update();

        expect(mockCreateSuccess).toHaveBeenCalledWith("Successfully {{action}} the {{item}}");
    });

    it("sends the credential id and the new description", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({ data: { status: "OK" } } as any);

        await update("Renamed");

        expect(updateUserWebAuthnCredential).toHaveBeenCalledWith("abc123", "Renamed");
    });
});

describe("description field", () => {
    it("does not update or close on Enter with an empty description", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");

        const handleClose = vi.fn();
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={handleClose} />);

        fireEvent.keyDown(screen.getByLabelText("Description"), { key: "Enter" });

        await waitFor(() => expect(screen.getByLabelText("Description")).toHaveValue(""));
        expect(updateUserWebAuthnCredential).not.toHaveBeenCalled();
        expect(handleClose).not.toHaveBeenCalled();
    });

    it("marks the field invalid after an empty update attempt", async () => {
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);

        fireEvent.keyDown(screen.getByLabelText("Description"), { key: "Enter" });

        await waitFor(() => expect(screen.getByLabelText("Description")).toHaveAttribute("aria-invalid", "true"));
    });

    it("rejects a description that is only whitespace", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");

        const handleClose = vi.fn();
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={handleClose} />);

        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "   " } });
        fireEvent.keyDown(screen.getByLabelText("Description"), { key: "Enter" });

        await waitFor(() => expect(screen.getByLabelText("Description")).toHaveAttribute("aria-invalid", "true"));
        expect(updateUserWebAuthnCredential).not.toHaveBeenCalled();
        expect(handleClose).not.toHaveBeenCalled();
    });

    it("trims the description before updating", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({ data: { status: "OK" } } as any);

        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);

        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "  Renamed  " } });

        await act(async () => {
            fireEvent.keyDown(screen.getByLabelText("Description"), { key: "Enter" });
        });

        expect(updateUserWebAuthnCredential).toHaveBeenCalledWith("abc123", "Renamed");
    });

    it("truncates a description longer than 30 characters", async () => {
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);

        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "a".repeat(45) } });

        await waitFor(() => expect(screen.getByLabelText("Description")).toHaveValue("a".repeat(30)));
    });

    it("keeps update disabled until a description is entered", async () => {
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);

        expect(document.getElementById("dialog-update")).toBeDisabled();

        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "New" } });

        await waitFor(() => expect(document.getElementById("dialog-update")).not.toBeDisabled());
    });

    it("updates on Enter", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");
        vi.mocked(updateUserWebAuthnCredential).mockResolvedValue({ data: { status: "OK" } } as any);

        const handleClose = vi.fn();
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={handleClose} />);

        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Renamed" } });
        fireEvent.keyDown(screen.getByLabelText("Description"), { key: "Enter" });

        await waitFor(() => expect(updateUserWebAuthnCredential).toHaveBeenCalledWith("abc123", "Renamed"));
        expect(handleClose).toHaveBeenCalled();
    });

    it("ignores keys other than Enter", async () => {
        const { updateUserWebAuthnCredential } = await import("@services/WebAuthn");

        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);

        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Renamed" } });
        fireEvent.keyDown(screen.getByLabelText("Description"), { key: "a" });

        expect(updateUserWebAuthnCredential).not.toHaveBeenCalled();
    });

    it("clears the field after cancelling", async () => {
        render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={vi.fn()} />);

        fireEvent.change(screen.getByLabelText("Description"), { target: { value: "Renamed" } });
        await waitFor(() => expect(screen.getByLabelText("Description")).toHaveValue("Renamed"));

        fireEvent.click(screen.getByText("Cancel"));

        await waitFor(() => expect(screen.getByLabelText("Description")).toHaveValue(""));
    });
});

it("does not render content when closed", () => {
    render(<WebAuthnCredentialEditDialog open={false} credential={credential} handleClose={vi.fn()} />);
    expect(screen.queryByText("Edit WebAuthn Credential")).not.toBeInTheDocument();
});

it("closes when Escape is pressed", async () => {
    const handleClose = vi.fn();
    render(<WebAuthnCredentialEditDialog open={true} credential={credential} handleClose={handleClose} />);

    fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

    await waitFor(() => expect(handleClose).toHaveBeenCalled());
});
