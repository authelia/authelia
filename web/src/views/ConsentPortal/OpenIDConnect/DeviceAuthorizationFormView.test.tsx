import { render, screen } from "@testing-library/react";

import { AuthenticationLevel } from "@services/State";
import DeviceAuthorizationFormView from "@views/ConsentPortal/OpenIDConnect/DeviceAuthorizationFormView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@mui/material", async () => {
    const actual = await vi.importActual("@mui/material");
    return {
        ...actual,
        useTheme: () => ({
            spacing: (n: number) => `${(n || 1) * 8}px`,
        }),
    };
});

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => vi.fn(),
}));

vi.mock("@constants/Routes", () => ({
    ConsentDecisionSubRoute: "/decision",
    ConsentOpenIDSubRoute: "/openid",
    ConsentRoute: "/consent",
    IndexRoute: "/",
}));

vi.mock("@constants/SearchParams", () => ({
    Flow: "flow",
    FlowNameOpenIDConnect: "openid-connect",
    SubFlow: "subflow",
    SubFlowNameDeviceAuthorization: "device-authorization",
    UserCode: "user_code",
}));

vi.mock("@layouts/LoginLayout", () => ({
    default: (props: any) => <div data-testid="login-layout">{props.children}</div>,
}));

vi.mock("@components/LogoutButton", () => ({
    default: () => <button data-testid="logout-button">Logout</button>,
}));

vi.mock("@components/SwitchUserButton", () => ({
    default: () => <button data-testid="switch-user-button">Switch User</button>,
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

it("renders loading page when unauthenticated", () => {
    render(
        <DeviceAuthorizationFormView state={{ authentication_level: AuthenticationLevel.Unauthenticated } as any} />,
    );
    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});

it("renders form when authenticated", () => {
    render(<DeviceAuthorizationFormView state={{ authentication_level: AuthenticationLevel.OneFactor } as any} />);
    expect(screen.getByTestId("login-layout")).toBeInTheDocument();
    expect(screen.getByText("Confirm")).toBeInTheDocument();
});
