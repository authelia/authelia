import { act, fireEvent, render, screen } from "@testing-library/react";

import { AuthenticationLevel } from "@services/State";
import DeviceAuthorizationFormView from "@views/ConsentPortal/OpenIDConnect/DeviceAuthorizationFormView";

const mocks = vi.hoisted(() => ({
    navigate: vi.fn(),
    userCode: { current: null as null | string },
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => mocks.userCode.current,
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mocks.navigate,
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

const renderView = (level: AuthenticationLevel = AuthenticationLevel.OneFactor) =>
    render(<DeviceAuthorizationFormView state={{ authentication_level: level } as any} />);

const expectedQuery = (code: string) => {
    const query = new URLSearchParams();

    query.set("user_code", code);
    query.set("flow", "openid-connect");
    query.set("subflow", "device-authorization");

    return query;
};

beforeEach(() => {
    vi.clearAllMocks();

    mocks.userCode.current = null;
});

it("renders loading page when unauthenticated", () => {
    renderView(AuthenticationLevel.Unauthenticated);

    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});

it("navigates to the index route when unauthenticated", () => {
    renderView(AuthenticationLevel.Unauthenticated);

    expect(mocks.navigate).toHaveBeenCalledWith("/", true, true, true, expect.any(URLSearchParams));
});

it("renders form when authenticated", () => {
    renderView();

    expect(screen.getByTestId("login-layout")).toBeInTheDocument();
    expect(screen.getByText("Confirm")).toBeInTheDocument();
});

it("explains where the code comes from", () => {
    renderView();

    expect(screen.getByText("Enter the code displayed on your device")).toBeInTheDocument();
});

it("renders the code entry inside a form", () => {
    const { container } = renderView();

    const form = container.querySelector("form");

    expect(form).toBeInTheDocument();
    expect(form).toContainElement(screen.getByLabelText("Code"));
    expect(form).toContainElement(screen.getByRole("button", { name: /Confirm/ }));
});

it("submits the form with the confirm button so a keyboard return works", () => {
    renderView();

    expect(screen.getByRole("button", { name: /Confirm/ })).toHaveAttribute("type", "submit");
});

it("navigates to the decision route when the code is submitted", async () => {
    renderView();

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Code"), { target: { value: "BGKMRTVX" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Confirm/ }));
    });

    expect(mocks.navigate).toHaveBeenCalledWith(
        "/consent/openid/decision",
        true,
        true,
        true,
        expectedQuery("BGKMRTVX"),
    );
});

it("uppercases the code as it is typed", async () => {
    renderView();

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Code"), { target: { value: "bgkmrtvx" } });
    });

    expect(screen.getByLabelText("Code")).toHaveValue("BGKMRTVX");
});

it("strips whitespace from a pasted code", async () => {
    renderView();

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Code"), { target: { value: " BGKM RTVX " } });
    });

    expect(screen.getByLabelText("Code")).toHaveValue("BGKMRTVX");
});

it("limits the code to the generated length", () => {
    renderView();

    expect(screen.getByLabelText("Code")).toHaveAttribute("maxlength", "8");
});

it("asks touch keyboards for the correct casing", () => {
    renderView();

    const input = screen.getByLabelText("Code");

    expect(input).toHaveAttribute("autocapitalize", "characters");
    expect(input).toHaveAttribute("autocomplete", "one-time-code");
    expect(input).toHaveAttribute("spellcheck", "false");
});

it("disables the confirmation until a code is entered", async () => {
    renderView();

    expect(screen.getByRole("button", { name: /Confirm/ })).toBeDisabled();

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Code"), { target: { value: "BGKMRTVX" } });
    });

    expect(screen.getByRole("button", { name: /Confirm/ })).toBeEnabled();
});

it("submits a code supplied in the url without waiting for the user", () => {
    mocks.userCode.current = "BGKMRTVX";

    renderView();

    expect(mocks.navigate).toHaveBeenCalledWith(
        "/consent/openid/decision",
        true,
        true,
        true,
        expectedQuery("BGKMRTVX"),
    );
});

it("submits a code supplied in the url only once", () => {
    mocks.userCode.current = "BGKMRTVX";

    const { rerender } = renderView();

    rerender(<DeviceAuthorizationFormView state={{ authentication_level: AuthenticationLevel.OneFactor } as any} />);

    expect(mocks.navigate).toHaveBeenCalledTimes(1);
});

it("separates the account actions", () => {
    const { container } = renderView();

    const separator = container.querySelector("[data-slot='separator']");

    expect(separator).toBeInTheDocument();
    expect(separator).toHaveAttribute("data-orientation", "vertical");
});
