import { render, screen } from "@testing-library/react";

import OneTimePasswordConfiguration from "@views/Settings/TwoFactorAuthentication/OneTimePasswordConfiguration";

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordCredentialItem", () => ({
    default: (props: any) => <div data-testid="credential-item" data-digits={props.config.digits} />,
}));

it("passes config and handlers to OneTimePasswordCredentialItem", () => {
    const config = { algorithm: 0, created_at: new Date(), digits: 6, issuer: "Authelia", period: 30 } as any;
    render(<OneTimePasswordConfiguration config={config} handleInformation={vi.fn()} handleDelete={vi.fn()} />);
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-digits", "6");
});
