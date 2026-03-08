import { render, screen } from "@testing-library/react";

import OneTimePasswordPanel from "@views/Settings/TwoFactorAuthentication/OneTimePasswordPanel";

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

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordConfiguration", () => ({
    default: (props: any) => <div data-testid="otp-config" data-issuer={props.config?.issuer} />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordDeleteDialog", () => ({
    default: () => <div data-testid="otp-delete-dialog" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordInformationDialog", () => ({
    default: () => <div data-testid="otp-info-dialog" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordRegisterDialog", () => ({
    default: () => <div data-testid="otp-register-dialog" />,
}));

it("renders panel with title and add button", () => {
    render(<OneTimePasswordPanel info={undefined} config={undefined} handleRefreshState={vi.fn()} />);
    expect(screen.getByText("One-Time Password")).toBeInTheDocument();
    expect(screen.getByText("Add")).toBeInTheDocument();
});

it("renders not registered message when config is undefined", () => {
    render(<OneTimePasswordPanel info={undefined} config={undefined} handleRefreshState={vi.fn()} />);
    expect(
        screen.getByText("The One-Time Password has not been registered if you'd like to register it click add"),
    ).toBeInTheDocument();
});

it("renders OTP configuration when config is provided", () => {
    const config = { digits: 6, issuer: "Authelia", period: 30 } as any;
    render(<OneTimePasswordPanel info={undefined} config={config} handleRefreshState={vi.fn()} />);
    expect(screen.getByTestId("otp-config")).toBeInTheDocument();
});

it("disables add button when config is provided", () => {
    const config = { digits: 6, issuer: "Authelia", period: 30 } as any;
    render(<OneTimePasswordPanel info={undefined} config={config} handleRefreshState={vi.fn()} />);
    expect(screen.getByText("Add").closest("button")).toBeDisabled();
});
