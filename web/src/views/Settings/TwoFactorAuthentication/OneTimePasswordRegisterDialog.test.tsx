import { render, screen } from "@testing-library/react";

import OneTimePasswordRegisterDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordRegisterDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        createSuccessNotification: vi.fn(),
    }),
}));

vi.mock("@constants/constants", () => ({
    GoogleAuthenticator: { link: "https://example.com", name: "Google Authenticator" },
}));

vi.mock("@models/TOTPConfiguration", () => ({
    toAlgorithmString: () => "SHA1",
}));

vi.mock("@services/OneTimePassword", () => ({
    completeTOTPRegister: vi.fn(),
    stopTOTPRegister: vi.fn(),
}));

vi.mock("@services/RegisterDevice", () => ({
    getTOTPSecret: vi.fn(),
}));

vi.mock("@services/UserInfoTOTPConfiguration", () => ({
    getTOTPOptions: vi.fn().mockResolvedValue({ algorithms: ["SHA1"], lengths: [6], periods: [30] }),
}));

vi.mock("@components/AppStoreBadges", () => ({
    default: () => <div data-testid="app-store-badges" />,
}));

vi.mock("@components/CopyButton", () => ({
    default: () => <button data-testid="copy-button">Copy</button>,
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/OTPDial", () => ({
    default: () => <div data-testid="otp-dial" />,
    State: { Failure: 3, Idle: 0, InProgress: 1, RateLimited: 4, Success: 2 },
}));

vi.mock("qrcode.react", () => ({
    QRCodeSVG: () => <div data-testid="qr-code" />,
}));

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders dialog with title when open", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(<OneTimePasswordRegisterDialog open={true} setClosed={vi.fn()} />);
    expect(screen.getByText("Register {{item}}")).toBeInTheDocument();
    expect(screen.getByText("Start")).toBeInTheDocument();
});

it("does not render content when closed", () => {
    render(<OneTimePasswordRegisterDialog open={false} setClosed={vi.fn()} />);
    expect(screen.queryByText("Register {{item}}")).not.toBeInTheDocument();
});
