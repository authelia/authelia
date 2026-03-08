import { render, screen } from "@testing-library/react";

import WebAuthnCredentialsPanel from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@services/UserSessionElevation", () => ({
    getUserSessionElevation: vi.fn(),
}));

vi.mock("@views/Settings/Common/IdentityVerificationDialog", () => ({
    default: () => <div data-testid="identity-dialog" />,
}));

vi.mock("@views/Settings/Common/SecondFactorDialog", () => ({
    default: () => <div data-testid="second-factor-dialog" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDeleteDialog", () => ({
    default: () => <div data-testid="delete-dialog" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialEditDialog", () => ({
    default: () => <div data-testid="edit-dialog" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialInformationDialog", () => ({
    default: () => <div data-testid="info-dialog" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialRegisterDialog", () => ({
    default: () => <div data-testid="register-dialog" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsGrid", () => ({
    default: () => <div data-testid="credentials-grid" />,
}));

it("renders panel with title and add button", () => {
    render(<WebAuthnCredentialsPanel info={undefined} credentials={undefined} handleRefreshState={vi.fn()} />);
    expect(screen.getByText("WebAuthn Credentials")).toBeInTheDocument();
    expect(screen.getByText("Add")).toBeInTheDocument();
});

it("renders no credentials message when credentials is empty", () => {
    render(<WebAuthnCredentialsPanel info={undefined} credentials={[]} handleRefreshState={vi.fn()} />);
    expect(
        screen.getByText("No WebAuthn Credentials have been registered if you'd like to register one click add"),
    ).toBeInTheDocument();
});

it("renders credentials grid when credentials are provided", () => {
    const credentials = [{ description: "Key 1", id: "1" }] as any;
    render(<WebAuthnCredentialsPanel info={undefined} credentials={credentials} handleRefreshState={vi.fn()} />);
    expect(screen.getByTestId("credentials-grid")).toBeInTheDocument();
});
