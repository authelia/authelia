import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import LoginPortal from "@views/LoginPortal/LoginPortal";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@constants/Routes", () => ({
    AuthenticatedRoute: "/authenticated",
    IndexRoute: "/",
    SecondFactorPasswordSubRoute: "/password",
    SecondFactorPushSubRoute: "/push",
    SecondFactorRoute: "/2fa",
    SecondFactorTOTPSubRoute: "/totp",
    SecondFactorWebAuthnSubRoute: "/webauthn",
}));

vi.mock("@contexts/LocalStorageMethodContext", () => ({
    useLocalStorageMethodContext: () => ({
        localStorageMethod: undefined,
        setLocalStorageMethod: vi.fn(),
    }),
}));

vi.mock("@hooks/Configuration", () => ({
    useConfiguration: () => [undefined, vi.fn(), false, null],
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => vi.fn(),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => vi.fn(),
}));

vi.mock("@hooks/State", () => ({
    useAutheliaState: () => [undefined, vi.fn(), false, null],
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoPOST: () => [undefined, vi.fn(), false, null],
}));

vi.mock("@services/SafeRedirection", () => ({
    checkSafeRedirection: vi.fn(),
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

it("renders loading page when state is not loaded", () => {
    render(
        <MemoryRouter>
            <LoginPortal
                duoSelfEnrollment={false}
                passkeyLogin={false}
                rememberMe={true}
                resetPassword={true}
                resetPasswordCustomURL=""
            />
        </MemoryRouter>,
    );
    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});
