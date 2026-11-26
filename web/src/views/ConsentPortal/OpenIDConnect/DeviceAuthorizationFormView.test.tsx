// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { AuthenticationLevel } from "@services/State";
import DeviceAuthorizationFormView from "@views/ConsentPortal/OpenIDConnect/DeviceAuthorizationFormView";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    navigate: vi.fn(),
    postConsentUserCode: vi.fn(),
    userCode: { current: null as null | string },
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({ createErrorNotification: mocks.createErrorNotification }),
}));

vi.mock("@services/ConsentOpenIDConnect", () => ({
    postConsentUserCode: mocks.postConsentUserCode,
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

    query.set("flow_id", "flow-1");
    query.set("user_code", code);
    query.set("flow", "openid_connect");
    query.set("subflow", "device_authorization");

    return query;
};

beforeEach(() => {
    vi.clearAllMocks();

    mocks.userCode.current = null;
    mocks.postConsentUserCode.mockResolvedValue({ flow_id: "flow-1" });
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

    expect(mocks.postConsentUserCode).toHaveBeenCalledWith("BGKMRTVX");
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

it("limits the code to the generated length", async () => {
    renderView();

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Code"), { target: { value: " BGKM RTVX ZZZZ" } });
    });

    expect(screen.getByLabelText("Code")).toHaveValue("BGKMRTVX");
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

it("notifies the user and stays on the page when the code is not accepted", async () => {
    mocks.postConsentUserCode.mockRejectedValue(new Error("failed"));

    renderView();

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Code"), { target: { value: "BGKMRTVX" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByRole("button", { name: /Confirm/ }));
    });

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("Failed to submit the user code");
    expect(mocks.navigate).not.toHaveBeenCalled();
});

it("submits a code supplied in the url without waiting for the user", async () => {
    mocks.userCode.current = "BGKMRTVX";

    renderView();

    await waitFor(() => expect(mocks.navigate).toHaveBeenCalled());

    expect(mocks.postConsentUserCode).toHaveBeenCalledWith("BGKMRTVX");
    expect(mocks.navigate).toHaveBeenCalledWith(
        "/consent/openid/decision",
        true,
        true,
        true,
        expectedQuery("BGKMRTVX"),
    );
});

it("submits a code supplied in the url only once", async () => {
    mocks.userCode.current = "BGKMRTVX";

    const { rerender } = renderView();

    rerender(<DeviceAuthorizationFormView state={{ authentication_level: AuthenticationLevel.OneFactor } as any} />);

    await waitFor(() => expect(mocks.navigate).toHaveBeenCalled());

    expect(mocks.postConsentUserCode).toHaveBeenCalledTimes(1);
    expect(mocks.navigate).toHaveBeenCalledTimes(1);
});

it("separates the account actions", () => {
    const { container } = renderView();

    const separator = container.querySelector("[data-slot='separator']");

    expect(separator).toBeInTheDocument();
    expect(separator).toHaveAttribute("data-orientation", "vertical");
});
