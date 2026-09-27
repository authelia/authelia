// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from "@testing-library/react";

import { useAutheliaState } from "@hooks/State";
import { useUserInfoGET } from "@hooks/UserInfo";
import ErrorView from "@views/Error/ErrorView";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    fetchState: vi.fn(),
    fetchUserInfo: vi.fn(),
    params: {} as Record<string, string | undefined>,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({
        t: (key: string, opts?: { host?: string }) => (opts?.host ? key.replace("{{host}}", opts.host) : key),
    }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({ createErrorNotification: mocks.createErrorNotification }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: (param: string) => mocks.params[param],
}));

vi.mock("@hooks/State", () => ({
    useAutheliaState: vi.fn(),
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoGET: vi.fn(),
}));

vi.mock("@components/ComponentOrLoading", () => ({
    default: (props: any) => (props.ready ? props.children : <div data-testid="loading" />),
}));

vi.mock("@components/SwitchUserButton", () => ({
    default: () => <button data-testid="switch-user-button" />,
}));

vi.mock("@components/LogoutButton", () => ({
    default: () => <button data-testid="logout-button" />,
}));

vi.mock("@components/LockIcon", () => ({
    default: () => <div data-testid="lock-icon" />,
}));

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => (
        <div data-testid="layout" data-id={props.id} data-title={props.title ?? ""}>
            {props.children}
        </div>
    ),
}));

const stateMock = vi.mocked(useAutheliaState);
const userInfoMock = vi.mocked(useUserInfoGET);

function setState(level?: number, error?: Error) {
    stateMock.mockReturnValue([
        level === undefined ? undefined : ({ authentication_level: level } as any),
        mocks.fetchState,
        false,
        error,
    ]);
}

function setUserInfo(displayName?: string, error?: Error) {
    userInfoMock.mockReturnValue([
        displayName === undefined ? undefined : ({ display_name: displayName } as any),
        mocks.fetchUserInfo,
        false,
        error,
    ]);
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.params = { ec: "forbidden", rd: "https://deny.example.com/secret.html" };
    setState(1);
    setUserInfo("John");
});

describe("access denied", () => {
    it("fetches the state on mount", () => {
        render(<ErrorView />);

        expect(mocks.fetchState).toHaveBeenCalled();
    });

    it("shows the loading page until the state is known", () => {
        setState(undefined);

        render(<ErrorView />);

        expect(screen.getByTestId("loading")).toBeInTheDocument();
        expect(screen.queryByTestId("layout")).not.toBeInTheDocument();
    });

    it("shows the loading page until an authenticated user's information is known", () => {
        setUserInfo(undefined);

        render(<ErrorView />);

        expect(mocks.fetchUserInfo).toHaveBeenCalled();
        expect(screen.getByTestId("loading")).toBeInTheDocument();
    });

    it("names the denied host to an authenticated user and offers to switch user", () => {
        render(<ErrorView />);

        expect(screen.getByTestId("layout")).toHaveAttribute("data-id", "access-denied-stage");
        expect(screen.getByTestId("layout")).toHaveAttribute("data-title", "Access Denied");
        expect(screen.getByText("You do not have permission to access deny.example.com")).toBeInTheDocument();
        expect(screen.getByTestId("lock-icon")).toBeInTheDocument();
        expect(screen.getByTestId("switch-user-button")).toBeInTheDocument();
        expect(screen.getByTestId("logout-button")).toBeInTheDocument();
    });

    it("does not offer to sign out an anonymous user", () => {
        setState(0);

        render(<ErrorView />);

        expect(mocks.fetchUserInfo).not.toHaveBeenCalled();
        expect(screen.getByTestId("layout")).toHaveAttribute("data-id", "access-denied-stage");
        expect(screen.queryByTestId("switch-user-button")).not.toBeInTheDocument();
        expect(screen.queryByTestId("logout-button")).not.toBeInTheDocument();
    });

    it.each([
        ["missing", undefined],
        ["not a URL", "deny.example.com is down, call 555-0100"],
        ["not http", "javascript:alert(1)"],
    ])("does not name the host when the redirection URL is %s", (_name, rd) => {
        mocks.params.rd = rd;

        render(<ErrorView />);

        expect(screen.getByText("You do not have permission to access this resource")).toBeInTheDocument();
    });
});

describe("other errors", () => {
    it("shows a generic error for an unknown error code", () => {
        mocks.params.ec = "unknown";

        render(<ErrorView />);

        expect(screen.getByTestId("layout")).toHaveAttribute("data-id", "error-stage");
        expect(screen.getByTestId("layout")).toHaveAttribute("data-title", "");
        expect(screen.getByText("An unexpected error occurred")).toBeInTheDocument();
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("notifies and stops loading when the state cannot be fetched", () => {
        setState(undefined, new Error("boom"));

        render(<ErrorView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith(
            "There was an issue retrieving the current user state",
        );
        expect(screen.getByTestId("layout")).toBeInTheDocument();
    });

    it("notifies and stops loading when the user information cannot be fetched", () => {
        setUserInfo(undefined, new Error("boom"));

        render(<ErrorView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving user preferences");
        expect(screen.getByTestId("layout")).toBeInTheDocument();
    });
});
