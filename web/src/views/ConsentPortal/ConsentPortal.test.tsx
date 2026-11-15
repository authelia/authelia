// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import ConsentPortal from "@views/ConsentPortal/ConsentPortal";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    fetchState: vi.fn(),
    fetchUserInfo: vi.fn(),
    resetNotification: vi.fn(),
    state: null as any,
    stateError: null as Error | null,
    userInfo: undefined as any,
    userInfoError: null as Error | null,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@constants/Routes", () => ({
    ConsentCompletionSubRoute: "/completion",
    ConsentOpenIDSubRoute: "/openid",
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
        resetNotification: mocks.resetNotification,
    }),
}));

vi.mock("@hooks/State", () => ({
    useAutheliaState: () => [mocks.state, mocks.fetchState, false, mocks.stateError],
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoGET: () => [mocks.userInfo, mocks.fetchUserInfo, false, mocks.userInfoError],
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

vi.mock("@views/ConsentPortal/OpenIDConnect/ConsentPortal", () => ({
    default: (props: any) => (
        <div
            data-testid="openid-consent"
            data-display-name={props.userInfo?.display_name ?? ""}
            data-level={props.state?.authentication_level}
        />
    ),
}));

vi.mock("@views/ConsentPortal/CompletionView", () => ({
    default: () => <div data-testid="completion-view" />,
}));

const userInfo = { display_name: "John Doe", emails: [] };

function renderPortal(route = "/openid") {
    return render(
        <MemoryRouter initialEntries={[route]}>
            <ConsentPortal />
        </MemoryRouter>,
    );
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.state = null;
    mocks.stateError = null;
    mocks.userInfo = undefined;
    mocks.userInfoError = null;
});

describe("loading", () => {
    it("renders loading page when state is null", () => {
        renderPortal();
        expect(screen.getByTestId("loading-page")).toBeInTheDocument();
    });

    it("fetches the state on mount", () => {
        renderPortal();
        expect(mocks.fetchState).toHaveBeenCalled();
    });

    it("keeps loading while the user info is still pending", () => {
        mocks.state = { authentication_level: 1, factor_knowledge: true, username: "john" };

        renderPortal();

        expect(screen.getByTestId("loading-page")).toBeInTheDocument();
        expect(screen.queryByTestId("openid-consent")).not.toBeInTheDocument();
    });

    it("does not fetch the user info while unauthenticated", async () => {
        mocks.state = { authentication_level: 0, factor_knowledge: false, username: "" };

        renderPortal();

        expect(await screen.findByTestId("openid-consent")).toBeInTheDocument();
        expect(mocks.fetchUserInfo).not.toHaveBeenCalled();
    });

    it("fetches the user info once authenticated", () => {
        mocks.state = { authentication_level: 1, factor_knowledge: true, username: "john" };

        renderPortal();

        expect(mocks.fetchUserInfo).toHaveBeenCalled();
    });
});

describe("routing", () => {
    it("renders the OpenID Connect consent portal", async () => {
        mocks.state = { authentication_level: 1, factor_knowledge: true, username: "john" };
        mocks.userInfo = userInfo;

        renderPortal("/openid");

        expect(await screen.findByTestId("openid-consent")).toHaveAttribute("data-display-name", "John Doe");
        expect(screen.getByTestId("openid-consent")).toHaveAttribute("data-level", "1");
    });

    it("renders the completion view", async () => {
        mocks.state = { authentication_level: 1, factor_knowledge: true, username: "john" };
        mocks.userInfo = userInfo;

        renderPortal("/completion");

        expect(await screen.findByTestId("completion-view")).toBeInTheDocument();
    });

    it("renders the consent portal for an unauthenticated user without user info", async () => {
        mocks.state = { authentication_level: 0, factor_knowledge: false, username: "" };

        renderPortal("/openid");

        expect(await screen.findByTestId("openid-consent")).toHaveAttribute("data-display-name", "");
    });
});

describe("fetch errors", () => {
    it("notifies when the state cannot be retrieved", async () => {
        mocks.stateError = new Error("boom");

        renderPortal();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "There was an issue retrieving the current user state",
            ),
        );
    });

    it("notifies when the user info cannot be retrieved", async () => {
        mocks.state = { authentication_level: 1, factor_knowledge: true, username: "john" };
        mocks.userInfoError = new Error("boom");

        renderPortal();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "There was an issue retrieving user preferences",
            ),
        );
    });
});
