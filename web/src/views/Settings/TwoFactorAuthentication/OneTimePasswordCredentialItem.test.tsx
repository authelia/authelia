import { render, screen } from "@testing-library/react";

import OneTimePasswordCredentialItem from "@views/Settings/TwoFactorAuthentication/OneTimePasswordCredentialItem";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/CredentialItem", () => ({
    default: (props: any) => <div data-testid="credential-item" data-description={props.description} />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordInformationDialog", () => ({
    default: () => <div data-testid="info-dialog" />,
}));

it("renders credential item with issuer as description", () => {
    const config = { created_at: new Date(), digits: 6, issuer: "Authelia", period: 30 } as any;
    render(<OneTimePasswordCredentialItem config={config} handleInformation={vi.fn()} handleDelete={vi.fn()} />);
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-description", "Authelia");
});
