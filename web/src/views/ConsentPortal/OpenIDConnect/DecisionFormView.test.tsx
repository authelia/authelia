import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";

import {
    getConsentResponse,
    postConsentResponseAccept,
    postConsentResponseReject,
    putDeviceCodeFlowUserCode,
} from "@services/ConsentOpenIDConnect";
import { postFirstFactorReauthenticate } from "@services/Password";
import { AuthenticationLevel } from "@services/State";
import DecisionFormView from "@views/ConsentPortal/OpenIDConnect/DecisionFormView";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    flow: { flow: "openid-connect", id: null as null | string, subflow: null as null | string },
    layoutProps: { current: null as any },
    navigate: vi.fn(),
    redirect: vi.fn(),
    resetNotification: vi.fn(),
    userCode: { current: null as null | string },
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("broadcast-channel", () => {
    class MockBroadcastChannel {
        addEventListener = vi.fn();
        postMessage = vi.fn();
    }

    return { BroadcastChannel: MockBroadcastChannel };
});

vi.mock("@hooks/Flow", () => ({
    useFlow: () => mocks.flow,
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => mocks.userCode.current,
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => mocks.redirect,
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mocks.navigate,
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
        resetNotification: mocks.resetNotification,
    }),
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: vi.fn(() => null),
}));

vi.mock("@services/ConsentOpenIDConnect", () => ({
    formatClaim: (translated: string, claim: string) => translated || claim,
    formatScope: (translated: string, scope: string) => translated || scope,
    getConsentResponse: vi.fn(),
    postConsentResponseAccept: vi.fn(),
    postConsentResponseReject: vi.fn(),
    putDeviceCodeFlowUserCode: vi.fn(),
}));

vi.mock("@services/Password", () => ({
    postFirstFactorReauthenticate: vi.fn(),
}));

vi.mock("@layouts/LoginLayout", () => ({
    default: (props: any) => {
        mocks.layoutProps.current = props;

        return <div data-testid="login-layout">{props.children}</div>;
    },
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

const userInfo = { display_name: "Test User", emails: ["test@example.com"], groups: [] } as any;

const response = (overrides: Partial<any> = {}) => ({
    audience: [],
    claims: null,
    client_description: "Test Client",
    client_id: "test-client",
    essential_claims: null,
    pre_configuration: false,
    require_login: false,
    resource: [],
    scopes: ["openid", "profile"],
    ...overrides,
});

const completionQuery = (decision: string) => {
    const query = new URLSearchParams();

    query.set("flow", "openid-connect");
    query.set("subflow", "device_authorization");
    query.set("decision", decision);

    return query;
};

const renderView = (level: AuthenticationLevel = AuthenticationLevel.TwoFactor) =>
    render(<DecisionFormView state={{ authentication_level: level } as any} userInfo={userInfo} />);

const renderLoaded = async (body: any = response(), level: AuthenticationLevel = AuthenticationLevel.TwoFactor) => {
    vi.mocked(getConsentResponse).mockResolvedValue(body);

    const utils = renderView(level);

    await waitFor(() => expect(screen.getByTestId("login-layout")).toBeInTheDocument());

    return utils;
};

beforeEach(() => {
    vi.clearAllMocks();

    mocks.flow.flow = "openid-connect";
    mocks.flow.id = "flow-1";
    mocks.flow.subflow = null;
    mocks.userCode.current = null;
    mocks.layoutProps.current = null;

    vi.spyOn(console, "error").mockImplementation(() => {});

    vi.mocked(getConsentResponse).mockResolvedValue(undefined as any);
    vi.mocked(postConsentResponseAccept).mockResolvedValue({ redirect_uri: "https://app.example.com/callback" });
    vi.mocked(postConsentResponseReject).mockResolvedValue({ redirect_uri: "https://app.example.com/denied" });
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the loading page while the consent response is not loaded", () => {
    renderView();

    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});

it("renders the loading page when the user info is not loaded", async () => {
    vi.mocked(getConsentResponse).mockResolvedValue(response());

    render(<DecisionFormView state={{ authentication_level: AuthenticationLevel.TwoFactor } as any} />);

    await waitFor(() => expect(getConsentResponse).toHaveBeenCalled());

    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});

it("navigates to the index route when unauthenticated", () => {
    renderView(AuthenticationLevel.Unauthenticated);

    expect(mocks.navigate).toHaveBeenCalledWith("/");
    expect(getConsentResponse).not.toHaveBeenCalled();
});

it("navigates to the index route when there is no flow id or user code", () => {
    mocks.flow.id = null;

    renderView();

    expect(mocks.navigate).toHaveBeenCalledWith("/");
    expect(getConsentResponse).not.toHaveBeenCalled();
});

it("requests the consent response with the flow id and user code", async () => {
    mocks.userCode.current = "ABCD-1234";

    await renderLoaded();

    expect(getConsentResponse).toHaveBeenCalledWith("flow-1", "ABCD-1234");
});

it("navigates to the index route when the consent response fails", async () => {
    vi.mocked(getConsentResponse).mockRejectedValue(new Error("boom"));

    renderView();

    await waitFor(() => expect(mocks.navigate).toHaveBeenCalledWith("/"));
});

it("renders the client description and client id", async () => {
    await renderLoaded();

    expect(screen.getByTestId("openid-consent-client-name")).toHaveTextContent("Test Client");
    expect(screen.getByText("test-client")).toBeInTheDocument();
});

it("renders the requested scopes", async () => {
    await renderLoaded();

    expect(document.getElementById("scope-openid")).toBeInTheDocument();
    expect(document.getElementById("scope-profile")).toBeInTheDocument();
});

it("renders the requested claims", async () => {
    await renderLoaded(response({ claims: ["name"], essential_claims: ["sub"] }));

    expect(document.getElementById("claim-sub-essential")).toBeInTheDocument();
    expect(document.getElementById("claim-name")).toBeInTheDocument();
});

it("renders the requested audience", async () => {
    await renderLoaded(response({ audience: ["https://api.example.com"] }));

    expect(document.getElementById("openid-consent-audience")).toBeInTheDocument();
    expect(screen.getByText("https://api.example.com")).toBeInTheDocument();
});

it("renders the requested resource", async () => {
    await renderLoaded(response({ resource: ["https://files.example.com"] }));

    expect(document.getElementById("openid-consent-resource")).toBeInTheDocument();
    expect(screen.getByText("https://files.example.com")).toBeInTheDocument();
});

it("omits the audience section when the audience is empty", async () => {
    await renderLoaded(response({ audience: [] }));

    expect(document.getElementById("openid-consent-audience")).not.toBeInTheDocument();
});

it("omits the resource section when the resource is null", async () => {
    await renderLoaded(response({ resource: null }));

    expect(document.getElementById("openid-consent-resource")).not.toBeInTheDocument();
});

it("omits the claims section when there are no claims", async () => {
    await renderLoaded();

    expect(document.getElementById("openid-consent-claims")).not.toBeInTheDocument();
});

it("renders the pre-configuration option when the client allows it", async () => {
    await renderLoaded(response({ pre_configuration: true }));

    expect(document.getElementById("pre-configure")).toBeInTheDocument();
});

it("omits the pre-configuration option when the client disallows it", async () => {
    await renderLoaded();

    expect(document.getElementById("pre-configure")).not.toBeInTheDocument();
});

it("omits the reauthentication prompt when a login is not required", async () => {
    await renderLoaded();

    expect(document.getElementById("openid-consent-prompt-login")).not.toBeInTheDocument();
});

it("uses a wider layout so the request details remain readable", async () => {
    await renderLoaded();

    expect(mocks.layoutProps.current.maxWidth).toBe("sm");
});

it("accepts the consent request and redirects", async () => {
    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(postConsentResponseAccept).toHaveBeenCalledWith(false, "test-client", [], "flow-1", null, null);
    expect(mocks.redirect).toHaveBeenCalledWith("https://app.example.com/callback");
});

it("accepts the consent request with the selected claims", async () => {
    await renderLoaded(response({ claims: ["name", "picture"], essential_claims: ["sub"] }));

    await act(async () => {
        fireEvent.click(document.getElementById("claim-picture") as HTMLElement);
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(postConsentResponseAccept).toHaveBeenCalledWith(false, "test-client", ["name"], "flow-1", null, null);
});

it("accepts the consent request with pre-configuration when requested", async () => {
    await renderLoaded(response({ pre_configuration: true }));

    await act(async () => {
        fireEvent.click(document.getElementById("pre-configure") as HTMLElement);
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(postConsentResponseAccept).toHaveBeenCalledWith(true, "test-client", [], "flow-1", null, null);
});

it("rejects the consent request and redirects", async () => {
    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Deny/ }));
    });

    expect(postConsentResponseReject).toHaveBeenCalledWith("test-client", "flow-1", null, null);
    expect(mocks.redirect).toHaveBeenCalledWith("https://app.example.com/denied");
});

it("shows a spinner on the accepting button only", async () => {
    let release: (value: any) => void = () => {};

    vi.mocked(postConsentResponseAccept).mockReturnValue(
        new Promise((resolve) => {
            release = resolve;
        }) as any,
    );

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(within(screen.getByRole("button", { name: /Accept/ })).getByTestId("spinner")).toBeInTheDocument();
    expect(within(screen.getByRole("button", { name: /Deny/ })).queryByTestId("spinner")).not.toBeInTheDocument();

    await act(async () => {
        release({ redirect_uri: "https://app.example.com/callback" });
    });
});

it("shows a spinner on the denying button only", async () => {
    let release: (value: any) => void = () => {};

    vi.mocked(postConsentResponseReject).mockReturnValue(
        new Promise((resolve) => {
            release = resolve;
        }) as any,
    );

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Deny/ }));
    });

    expect(within(screen.getByRole("button", { name: /Deny/ })).getByTestId("spinner")).toBeInTheDocument();
    expect(within(screen.getByRole("button", { name: /Accept/ })).queryByTestId("spinner")).not.toBeInTheDocument();

    await act(async () => {
        release({ redirect_uri: "https://app.example.com/denied" });
    });
});

it("disables both decisions while a decision is in flight", async () => {
    let release: (value: any) => void = () => {};

    vi.mocked(postConsentResponseAccept).mockReturnValue(
        new Promise((resolve) => {
            release = resolve;
        }) as any,
    );

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(screen.getByRole("button", { name: /Accept/ })).toBeDisabled();
    expect(screen.getByRole("button", { name: /Deny/ })).toBeDisabled();

    await act(async () => {
        release({ redirect_uri: "https://app.example.com/callback" });
    });
});

it("notifies the user when the acceptance has no redirect target", async () => {
    vi.mocked(postConsentResponseAccept).mockResolvedValue({});

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("Failed to redirect you");
});

it("notifies the user when the rejection has no redirect target", async () => {
    vi.mocked(postConsentResponseReject).mockResolvedValue({});

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Deny/ }));
    });

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("Failed to redirect you");
});

it("does not render the reauthentication prompt before a decision is made", async () => {
    await renderLoaded(response({ require_login: true }));

    expect(document.getElementById("openid-consent-prompt-login")).not.toBeInTheDocument();
});

it("keeps the request details expanded before a decision is made", async () => {
    await renderLoaded(response({ require_login: true }));

    expect(document.getElementById("openid-consent-scopes")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Show request details" })).not.toBeInTheDocument();
});

it("offers the acceptance without a password", async () => {
    await renderLoaded(response({ require_login: true }));

    expect(screen.getByRole("button", { name: /Accept/ })).toBeEnabled();
});

it("allows the rejection without a password", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Deny/ }));
    });

    expect(postConsentResponseReject).toHaveBeenCalled();
});

it("reveals the reauthentication prompt when the request is accepted", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(document.getElementById("openid-consent-prompt-login")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
});

it("does not post the consent when the acceptance only opens the reauthentication prompt", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(postConsentResponseAccept).not.toHaveBeenCalled();
    expect(postFirstFactorReauthenticate).not.toHaveBeenCalled();
});

it("collapses the request details when the reauthentication prompt opens", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(document.getElementById("openid-consent-scopes")).not.toBeInTheDocument();
    expect(screen.getByTestId("openid-consent-client-name")).toHaveTextContent("Test Client");
});

it("allows the collapsed request details to be reopened", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: "Show request details" }));
    });

    expect(document.getElementById("openid-consent-scopes")).toBeInTheDocument();
});

it("replaces the decisions with a submission and a cancellation", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(screen.getByRole("button", { name: /Submit/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Cancel/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Deny/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Accept/ })).not.toBeInTheDocument();
});

it("focuses the password field when the reauthentication prompt opens", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await waitFor(() => expect(screen.getByLabelText("Password")).toHaveFocus());
});

it("reauthenticates and then accepts when the password is submitted", async () => {
    vi.mocked(getConsentResponse)
        .mockResolvedValueOnce(response({ require_login: true }))
        .mockResolvedValueOnce(response({ require_login: false }));

    vi.mocked(postFirstFactorReauthenticate).mockResolvedValue(undefined as any);

    renderView();

    await waitFor(() => expect(screen.getByTestId("login-layout")).toBeInTheDocument());

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(postFirstFactorReauthenticate).toHaveBeenCalledWith(
        "password",
        undefined,
        undefined,
        "flow-1",
        "openid-connect",
        null,
        null,
    );
    expect(postConsentResponseAccept).toHaveBeenCalledWith(false, "test-client", [], "flow-1", null, null);
    expect(mocks.redirect).toHaveBeenCalledWith("https://app.example.com/callback");
});

it("reports an incorrect password inline and allows another attempt", async () => {
    vi.mocked(postFirstFactorReauthenticate).mockRejectedValue(new Error("bad password"));

    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrong" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(screen.getByRole("alert")).toHaveTextContent("Incorrect password");
    expect(screen.getByLabelText("Password")).toHaveValue("");
    expect(screen.getByRole("button", { name: /Submit/ })).toBeEnabled();
    expect(postConsentResponseAccept).not.toHaveBeenCalled();
});

it("does not notify separately when the password is incorrect", async () => {
    vi.mocked(postFirstFactorReauthenticate).mockRejectedValue(new Error("bad password"));

    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrong" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(mocks.createErrorNotification).not.toHaveBeenCalled();
});

it("accepts on a second attempt after an incorrect password", async () => {
    vi.mocked(getConsentResponse)
        .mockResolvedValueOnce(response({ require_login: true }))
        .mockResolvedValueOnce(response({ require_login: false }));

    vi.mocked(postFirstFactorReauthenticate)
        .mockRejectedValueOnce(new Error("bad password"))
        .mockResolvedValueOnce(undefined as any);

    renderView();

    await waitFor(() => expect(screen.getByTestId("login-layout")).toBeInTheDocument());

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrong" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(postConsentResponseAccept).toHaveBeenCalledWith(false, "test-client", [], "flow-1", null, null);
});

it("clears the failure once a new password is typed", async () => {
    vi.mocked(postFirstFactorReauthenticate).mockRejectedValue(new Error("bad password"));

    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrong" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "another" } });
    });

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
});

it("reports inline when a login is still required after reauthenticating", async () => {
    vi.mocked(getConsentResponse).mockResolvedValue(response({ require_login: true }));
    vi.mocked(postFirstFactorReauthenticate).mockResolvedValue(undefined as any);

    renderView();

    await waitFor(() => expect(screen.getByTestId("login-layout")).toBeInTheDocument());

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(screen.getByRole("alert")).toHaveTextContent("Failed to confirm your identity");
    expect(postConsentResponseAccept).not.toHaveBeenCalled();
});

it("returns to the decision when the reauthentication is cancelled", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Cancel/ }));
    });

    expect(screen.getByRole("button", { name: /Accept/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Deny/ })).toBeInTheDocument();
    expect(document.getElementById("openid-consent-prompt-login")).not.toBeInTheDocument();
    expect(document.getElementById("openid-consent-scopes")).toBeInTheDocument();
});

it("discards the password when the reauthentication is cancelled", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Cancel/ }));
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(screen.getByLabelText("Password")).toHaveValue("");
});

it("discards the failure when the reauthentication is cancelled", async () => {
    vi.mocked(postFirstFactorReauthenticate).mockRejectedValue(new Error("bad password"));

    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrong" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Cancel/ }));
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
});

it("does not submit the reauthentication without a password", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(postFirstFactorReauthenticate).not.toHaveBeenCalled();
    expect(screen.getByLabelText("Password")).toHaveAttribute("aria-invalid", "true");
});

it("shows a spinner on the submitting button while reauthenticating", async () => {
    let release: (value: any) => void = () => {};

    vi.mocked(postFirstFactorReauthenticate).mockReturnValue(
        new Promise((resolve) => {
            release = resolve;
        }) as any,
    );

    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(within(screen.getByRole("button", { name: /Submit/ })).getByTestId("spinner")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeDisabled();

    await act(async () => {
        release(undefined);
    });
});
it("submits the user code and navigates to completion when accepting a device authorization", async () => {
    mocks.flow.subflow = "device_authorization";
    mocks.userCode.current = "ABCD-1234";

    vi.mocked(postConsentResponseAccept).mockResolvedValue({ flow_id: "flow-2" });

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(putDeviceCodeFlowUserCode).toHaveBeenCalledWith("flow-2", "ABCD-1234");
    expect(mocks.navigate).toHaveBeenCalledWith(
        "/consent/completion",
        false,
        false,
        false,
        completionQuery("accepted"),
    );
});

it("navigates to completion when denying a device authorization", async () => {
    mocks.flow.subflow = "device_authorization";
    mocks.userCode.current = "ABCD-1234";

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Deny/ }));
    });

    expect(mocks.navigate).toHaveBeenCalledWith(
        "/consent/completion",
        false,
        false,
        false,
        completionQuery("rejected"),
    );
});

it("notifies the user when the device authorization acceptance has no flow id", async () => {
    mocks.flow.subflow = "device_authorization";
    mocks.userCode.current = "ABCD-1234";

    vi.mocked(postConsentResponseAccept).mockResolvedValue({});

    await renderLoaded();

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("Failed to submit the user code");
    expect(putDeviceCodeFlowUserCode).not.toHaveBeenCalled();
});

it("forwards the subflow when reauthenticating during a device authorization", async () => {
    mocks.flow.subflow = "device_authorization";
    mocks.userCode.current = "ABCD-1234";

    vi.mocked(getConsentResponse)
        .mockResolvedValueOnce(response({ require_login: true }))
        .mockResolvedValueOnce(response({ require_login: false }));

    vi.mocked(postFirstFactorReauthenticate).mockResolvedValue(undefined as any);
    vi.mocked(postConsentResponseAccept).mockResolvedValue({ flow_id: "flow-2" });

    renderView();

    await waitFor(() => expect(screen.getByTestId("login-layout")).toBeInTheDocument());

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Submit/ }));
    });

    expect(postFirstFactorReauthenticate).toHaveBeenCalledWith(
        "password",
        undefined,
        undefined,
        "flow-1",
        "openid-connect",
        "device_authorization",
        "ABCD-1234",
    );
    expect(putDeviceCodeFlowUserCode).toHaveBeenCalledWith("flow-2", "ABCD-1234");
});

it("gives the acceptance the filled treatment and the denial the outlined one", async () => {
    await renderLoaded();

    expect(screen.getByRole("button", { name: /Accept/ })).toHaveAttribute("data-variant", "default");
    expect(screen.getByRole("button", { name: /Deny/ })).toHaveAttribute("data-variant", "outline");
});

it("gives the submission the filled treatment and the cancellation the outlined one", async () => {
    await renderLoaded(response({ require_login: true }));

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Accept/ }));
    });

    expect(screen.getByRole("button", { name: /Submit/ })).toHaveAttribute("data-variant", "default");
    expect(screen.getByRole("button", { name: /Cancel/ })).toHaveAttribute("data-variant", "outline");
});

it("separates the account actions", async () => {
    const { container } = await renderLoaded();

    const separator = container.querySelector("[data-slot='separator']");

    expect(separator).toBeInTheDocument();
    expect(separator).toHaveAttribute("data-orientation", "vertical");
});
