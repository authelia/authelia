import { render, screen } from "@testing-library/react";

import { SecondFactorMethod } from "@models/Methods";
import TwoFactorAuthenticationOptionsPanel from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsPanel";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/LocalStorageMethodContext", () => ({
    useLocalStorageMethodContext: () => ({
        localStorageMethod: SecondFactorMethod.TOTP,
        localStorageMethodAvailable: true,
        setLocalStorageMethod: vi.fn(),
    }),
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@services/UserInfo", () => ({
    isMethod2FA: () => true,
    Method2FA: {},
    setPreferred2FAMethod: vi.fn().mockResolvedValue(undefined),
    toSecondFactorMethod: () => SecondFactorMethod.TOTP,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsMethodsRadioGroup", () => ({
    default: (props: any) => <div data-testid={`radio-group-${props.id}`} data-name={props.name} />,
}));

const config = {
    available_methods: new Set([SecondFactorMethod.TOTP, SecondFactorMethod.WebAuthn]),
} as any;

const info = {
    has_duo: false,
    has_totp: true,
    has_webauthn: true,
    method: SecondFactorMethod.TOTP,
} as any;

it("renders options panel with radio groups", () => {
    render(<TwoFactorAuthenticationOptionsPanel config={config} info={info} refresh={vi.fn()} />);
    expect(screen.getByText("Options")).toBeInTheDocument();
    expect(screen.getByTestId("radio-group-account")).toBeInTheDocument();
    expect(screen.getByTestId("radio-group-local")).toBeInTheDocument();
});

it("renders nothing when user has no methods", () => {
    const noMethodsInfo = { ...info, has_duo: false, has_totp: false, has_webauthn: false };
    const { container } = render(
        <TwoFactorAuthenticationOptionsPanel config={config} info={noMethodsInfo} refresh={vi.fn()} />,
    );
    expect(container.textContent).toBe("");
});
