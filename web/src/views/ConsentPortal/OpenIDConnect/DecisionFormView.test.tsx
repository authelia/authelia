import { render, screen, waitFor } from "@testing-library/react";

import { useRouterNavigate } from "@hooks/RouterNavigate";
import { getConsentResponse } from "@services/ConsentOpenIDConnect";
import { AuthenticationLevel } from "@services/State";
import DecisionFormView from "@views/ConsentPortal/OpenIDConnect/DecisionFormView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { appLogo: "", buttonRow: "", form: "", formField: "" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@mui/material", async () => {
    const actual = await vi.importActual("@mui/material");
    return {
        ...actual,
        useTheme: () => ({
            palette: { grey: { 600: "#999" } },
            spacing: (n: number) => `${(n || 1) * 8}px`,
        }),
    };
});

vi.mock("broadcast-channel", () => {
    class MockBroadcastChannel {
        addEventListener = vi.fn();
        postMessage = vi.fn();
    }
    return { BroadcastChannel: MockBroadcastChannel };
});

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: "openid-connect", id: "flow-1", subflow: null }),
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        resetNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => vi.fn(),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: vi.fn(),
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: () => null,
}));

vi.mock("@services/ConsentOpenIDConnect", () => ({
    getConsentResponse: vi.fn().mockResolvedValue(undefined),
    postConsentResponseAccept: vi.fn(),
    postConsentResponseReject: vi.fn(),
    putDeviceCodeFlowUserCode: vi.fn(),
}));

vi.mock("@services/Password", () => ({
    postFirstFactorReauthenticate: vi.fn(),
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

vi.mock("@views/ConsentPortal/OpenIDConnect/DecisionFormClaims", () => ({
    default: () => <div data-testid="claims-form" />,
}));

vi.mock("@views/ConsentPortal/OpenIDConnect/DecisionFormPreConfiguration", () => ({
    default: () => <div data-testid="pre-config-form" />,
}));

vi.mock("@views/ConsentPortal/OpenIDConnect/DecisionFormScopes", () => ({
    default: () => <div data-testid="scopes-form" />,
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

beforeEach(() => {
    vi.mocked(useRouterNavigate).mockReturnValue(vi.fn());
    vi.mocked(getConsentResponse).mockResolvedValue(undefined as any);
});

it("renders loading page when consent response is not loaded", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(<DecisionFormView state={{ authentication_level: AuthenticationLevel.TwoFactor } as any} />);
    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});

it("navigates to index route when unauthenticated", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});

    const mockNavigate = vi.fn();

    vi.mocked(useRouterNavigate).mockReturnValue(mockNavigate);

    render(<DecisionFormView state={{ authentication_level: AuthenticationLevel.Unauthenticated } as any} />);

    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("renders consent form with scopes, claims, and pre-configuration when consent response resolves", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});

    vi.mocked(getConsentResponse).mockResolvedValue({
        audience: [],
        claims: ["email"],
        client_description: "Test Client",
        client_id: "test-client",
        essential_claims: [],
        pre_configuration: true,
        require_login: false,
        scopes: ["openid", "profile"],
    });

    render(
        <DecisionFormView
            state={{ authentication_level: AuthenticationLevel.TwoFactor } as any}
            userInfo={{ display_name: "Test User", emails: ["test@example.com"], groups: [] } as any}
        />,
    );

    await waitFor(() => {
        expect(screen.getByTestId("login-layout")).toBeInTheDocument();
    });

    expect(screen.getByTestId("scopes-form")).toBeInTheDocument();
    expect(screen.getByTestId("claims-form")).toBeInTheDocument();
    expect(screen.getByTestId("pre-config-form")).toBeInTheDocument();
});
