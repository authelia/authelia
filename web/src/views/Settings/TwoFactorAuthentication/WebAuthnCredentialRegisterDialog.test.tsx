import { render, screen } from "@testing-library/react";

import WebAuthnCredentialRegisterDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialRegisterDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: {},
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        createSuccessNotification: vi.fn(),
    }),
}));

vi.mock("@models/WebAuthn", () => ({
    AttestationResult: { Success: 0 },
    AttestationResultFailureString: () => "failure",
    WebAuthnTouchState: { Failure: 2, InProgress: 1, WaitTouch: 0 },
}));

vi.mock("@services/WebAuthn", () => ({
    finishWebAuthnRegistration: vi.fn(),
    getWebAuthnRegistrationOptions: vi.fn(),
    startWebAuthnRegistration: vi.fn(),
}));

vi.mock("@components/InformationIcon", () => ({
    default: () => <div data-testid="info-icon" />,
}));

vi.mock("@components/WebAuthnRegisterIcon", () => ({
    default: () => <div data-testid="webauthn-register-icon" />,
}));

it("renders dialog with stepper when open", () => {
    render(<WebAuthnCredentialRegisterDialog open={true} setClosed={vi.fn()} />);
    expect(screen.getByText("Register {{item}}")).toBeInTheDocument();
    expect(screen.getAllByText("Description").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("Verification")).toBeInTheDocument();
});

it("does not render content when closed", () => {
    render(<WebAuthnCredentialRegisterDialog open={false} setClosed={vi.fn()} />);
    expect(screen.queryByText("Register {{item}}")).not.toBeInTheDocument();
});
