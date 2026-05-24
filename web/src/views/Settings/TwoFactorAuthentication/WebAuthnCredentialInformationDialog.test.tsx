import { fireEvent, render, screen } from "@testing-library/react";

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
