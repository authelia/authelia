// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import WebAuthnCredentialInformationDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialInformationDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@components/CopyButton", () => ({
    default: () => <button data-testid="copy-button">Copy</button>,
}));

vi.mock("@i18n/formats", () => ({
    FormatDateHumanReadable: {},
}));

vi.mock("@models/WebAuthn", () => ({
    toAttachmentName: (v: string) => v || "unknown",
    toTransportName: (v: string) => v || "unknown",
}));

const credential = {
    aaguid: "test-aaguid",
    attestation_format: "packed",
    attestation_type: "basic_full",
    backup_eligible: true,
    backup_state: true,
    clone_warning: false,
    created_at: "2024-01-01T00:00:00Z",
    description: "My Key",
    discoverable: true,
    id: "cred-1",
    kid: "kid-123",
    last_used_at: "2024-06-01T00:00:00Z",
    legacy: false,
    public_key: "pk-123",
    rpid: "example.com",
    transports: ["usb"],
} as any;

it("renders credential information when open", () => {
    render(<WebAuthnCredentialInformationDialog open={true} credential={credential} handleClose={vi.fn()} />);
    expect(screen.getByText("WebAuthn Credential Information")).toBeInTheDocument();
    expect(screen.getByText("My Key")).toBeInTheDocument();
});

it("renders not loaded message when credential is undefined", () => {
    render(<WebAuthnCredentialInformationDialog open={true} handleClose={vi.fn()} />);
    expect(screen.getByText("The WebAuthn Credential information is not loaded")).toBeInTheDocument();
});

it("calls handleClose when close button is clicked", () => {
    const handleClose = vi.fn();
    render(<WebAuthnCredentialInformationDialog open={true} credential={credential} handleClose={handleClose} />);
    fireEvent.click(screen.getByText("Close"));
    expect(handleClose).toHaveBeenCalledOnce();
});

describe("credential variations", () => {
    function renderWith(overrides: Record<string, unknown>) {
        return render(
            <WebAuthnCredentialInformationDialog
                open={true}
                credential={{ ...credential, ...overrides }}
                handleClose={vi.fn()}
            />,
        );
    }

    it("warns about a legacy credential", () => {
        renderWith({ legacy: true });

        expect(
            screen.getByText(
                "This is a legacy WebAuthn Credential if it's not operating normally you may need to delete it and register it again",
            ),
        ).toBeInTheDocument();
    });

    it("does not warn about a modern credential", () => {
        renderWith({ legacy: false });

        expect(
            screen.queryByText(
                "This is a legacy WebAuthn Credential if it's not operating normally you may need to delete it and register it again",
            ),
        ).not.toBeInTheDocument();
    });

    it("falls back to Unknown for a missing authenticator GUID", () => {
        renderWith({ aaguid: undefined, attestation_type: "basic_full", transports: ["usb"] });

        expect(screen.getAllByText("Unknown").length).toBeGreaterThanOrEqual(1);
    });

    it("falls back to Unknown for an empty attestation type", () => {
        renderWith({ attestation_type: "" });

        expect(screen.getAllByText("Unknown").length).toBeGreaterThanOrEqual(1);
    });

    it("reports a backed up credential", () => {
        renderWith({ backup_eligible: true, backup_state: true });

        expect(screen.getByText("Backed Up")).toBeInTheDocument();
    });

    it("reports a credential that is eligible but not backed up", () => {
        renderWith({ backup_eligible: true, backup_state: false });

        expect(screen.getByText("Eligible")).toBeInTheDocument();
    });

    it("reports a credential that is not eligible for backup", () => {
        renderWith({ backup_eligible: false, backup_state: false });

        expect(screen.getByText("Not Eligible")).toBeInTheDocument();
    });

    it("lists the transports", () => {
        renderWith({ transports: ["usb", "nfc"] });

        expect(screen.getByText("usb, nfc")).toBeInTheDocument();
    });

    it("falls back to Unknown for null transports", () => {
        renderWith({ transports: null });

        expect(screen.getAllByText("Unknown").length).toBeGreaterThanOrEqual(1);
    });

    it("falls back to Unknown for empty transports", () => {
        renderWith({ transports: [] });

        expect(screen.getAllByText("Unknown").length).toBeGreaterThanOrEqual(1);
    });

    it("reports Yes for a discoverable, verified, clone warned credential", () => {
        renderWith({ clone_warning: true, discoverable: true, verified: true });

        expect(screen.getAllByText("Yes").length).toBe(3);
    });

    it("reports No for a non-discoverable, unverified credential without a clone warning", () => {
        renderWith({ clone_warning: false, discoverable: false, verified: false });

        expect(screen.getAllByText("No").length).toBe(3);
    });

    it("reports Never for a credential that has not been used", () => {
        renderWith({ last_used_at: undefined });

        expect(screen.getByText("Never")).toBeInTheDocument();
    });

    it("renders the copy buttons for the KID and public key", () => {
        renderWith({});

        expect(screen.getAllByTestId("copy-button")).toHaveLength(2);
    });

    it("omits the copy buttons without a credential", () => {
        render(<WebAuthnCredentialInformationDialog open={true} handleClose={vi.fn()} />);

        expect(screen.queryByTestId("copy-button")).not.toBeInTheDocument();
    });
});

it("does not render content when closed", () => {
    render(<WebAuthnCredentialInformationDialog open={false} credential={credential} handleClose={vi.fn()} />);
    expect(screen.queryByText("WebAuthn Credential Information")).not.toBeInTheDocument();
});

it("closes when Escape is pressed", async () => {
    const handleClose = vi.fn();
    render(<WebAuthnCredentialInformationDialog open={true} credential={credential} handleClose={handleClose} />);

    fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

    await waitFor(() => expect(handleClose).toHaveBeenCalled());
});
