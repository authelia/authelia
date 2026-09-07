import { render, screen } from "@testing-library/react";

import WebAuthnCredentialItem from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialItem";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/CredentialItem", () => ({
    default: (props: any) => (
        <div data-testid="credential-item" data-description={props.description} data-problem={props.problem} />
    ),
}));

const credential = {
    attestation_format: "fido-u2f",
    attestation_type: "none",
    created_at: "2024-01-01T00:00:00Z",
    description: "Security Key",
    id: "abc123",
    legacy: false,
} as any;

it("renders credential item with correct description", () => {
    render(
        <WebAuthnCredentialItem
            index={0}
            credential={credential}
            handleInformation={vi.fn()}
            handleEdit={vi.fn()}
            handleDelete={vi.fn()}
        />,
    );
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-description", "Security Key");
});

it("passes legacy flag as problem prop", () => {
    render(
        <WebAuthnCredentialItem
            index={0}
            credential={{ ...credential, legacy: true }}
            handleInformation={vi.fn()}
            handleEdit={vi.fn()}
            handleDelete={vi.fn()}
        />,
    );
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-problem", "true");
});
