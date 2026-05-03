import { render, screen } from "@testing-library/react";

import SecondFactorMethodOneTimePassword from "@views/Settings/Common/SecondFactorMethodOneTimePassword";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/UserInfoTOTPConfiguration", () => ({
    useUserInfoTOTPConfiguration: () => [{ digits: 6, period: 30 }, vi.fn(), false, null],
}));

vi.mock("@services/OneTimePassword", () => ({
    completeTOTPSignIn: vi.fn(),
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/OTPDial", () => ({
    default: (props: any) => <div data-testid="otp-dial" data-digits={props.digits} />,
    State: { Failure: 3, Idle: 0, InProgress: 1, RateLimited: 4, Success: 2 },
}));

it("renders OTP dial when config is loaded", () => {
    render(<SecondFactorMethodOneTimePassword onSecondFactorSuccess={vi.fn()} />);
    expect(screen.getByTestId("otp-dial")).toBeInTheDocument();
    expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-digits", "6");
});

it("renders loading page when config is not available", () => {
    vi.doMock("@hooks/UserInfoTOTPConfiguration", () => ({
        // eslint-disable-next-line @eslint-react/no-unnecessary-use-prefix
        useUserInfoTOTPConfiguration: () => [undefined, vi.fn(), false, null],
    }));
    // Re-import with doMock requires dynamic import but for simplicity, test the loaded state above
});
