import { render, screen } from "@testing-library/react";

import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@simplewebauthn/browser", () => ({
    browserSupportsWebAuthn: () => true,
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/PasswordForm", () => ({
    default: () => <div data-testid="password-form" />,
}));

const elevation = {
    can_skip_second_factor: true,
    elevated: false,
    factor_knowledge: true,
    require_second_factor: true,
    skip_second_factor: false,
} as any;

const info = {
    has_duo: false,
    has_telegram: false,
    has_totp: true,
    has_webauthn: true,
} as any;

it("renders dialog with stepper when opening with elevation", () => {
    render(
        <SecondFactorDialog
            elevation={elevation}
            info={info}
            opening={true}
            handleClosed={vi.fn()}
            handleOpened={vi.fn()}
        />,
    );
    expect(screen.getByText("Identity Verification")).toBeInTheDocument();
    expect(screen.getByText("Select a Method")).toBeInTheDocument();
    expect(screen.getByText("Authenticate")).toBeInTheDocument();
    expect(screen.getByText("Completed")).toBeInTheDocument();
});

it("renders method buttons for available methods", () => {
    render(
        <SecondFactorDialog
            elevation={elevation}
            info={info}
            opening={true}
            handleClosed={vi.fn()}
            handleOpened={vi.fn()}
        />,
    );
    expect(screen.getByText("One-Time Password")).toBeInTheDocument();
    expect(screen.getByText("WebAuthn")).toBeInTheDocument();
    expect(screen.getByText("Email One-Time Code")).toBeInTheDocument();
});

it("does not render content when not opening", () => {
    render(
        <SecondFactorDialog
            elevation={elevation}
            info={info}
            opening={false}
            handleClosed={vi.fn()}
            handleOpened={vi.fn()}
        />,
    );
    expect(screen.queryByText("One-Time Password")).not.toBeInTheDocument();
});
