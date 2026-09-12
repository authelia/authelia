// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { cancelSignOut, getSignOutPending, signOut } from "@services/SignOut";
import SignOut from "@views/LoginPortal/SignOut/SignOut";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    navigate: vi.fn(),
    redirectionURL: null as null | string,
    redirector: vi.fn(),
    searchParams: new URLSearchParams(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options ? key.replace(/{{(\w+)}}/g, (_, name: string) => options[name] ?? "") : key,
    }),
}));

vi.mock("react-router", () => ({
    useSearchParams: () => [mocks.searchParams, vi.fn()],
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

vi.mock("@constants/SearchParams", () => ({
    FlowID: "flow_id",
    RedirectionRestoreURL: "rd_restore",
    RedirectionURL: "rd",
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => mocks.redirectionURL,
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => mocks.redirector,
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mocks.navigate,
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => <div data-testid="layout">{props.children}</div>,
}));

vi.mock("@services/SignOut", () => ({
    cancelSignOut: vi.fn(),
    getSignOutPending: vi.fn(),
    signOut: vi.fn(),
}));

const cancelSignOutMock = vi.mocked(cancelSignOut);
const getSignOutPendingMock = vi.mocked(getSignOutPending);
const signOutMock = vi.mocked(signOut);

async function advance(ms: number) {
    await act(async () => {
        await vi.advanceTimersByTimeAsync(ms);
    });
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.redirectionURL = null;
    mocks.searchParams = new URLSearchParams();
    signOutMock.mockResolvedValue({ safeTargetURL: false } as any);
    getSignOutPendingMock.mockResolvedValue({ pending: false });
    cancelSignOutMock.mockResolvedValue(undefined);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders sign out message", async () => {
        render(<SignOut />);

        await waitFor(() => expect(signOutMock).toHaveBeenCalled());
        expect(screen.getByText(/You're being signed out and redirected/)).toBeInTheDocument();
    });

    it("signs out with the redirection URL", async () => {
        mocks.redirectionURL = "https://app.example.com";

        render(<SignOut />);

        await waitFor(() => expect(signOutMock).toHaveBeenCalledWith("https://app.example.com", expect.anything()));
    });
});

describe("redirection", () => {
    it("navigates to the index route after the delay", async () => {
        vi.useFakeTimers();

        render(<SignOut />);

        await vi.waitFor(() => expect(signOutMock).toHaveBeenCalled());
        expect(mocks.navigate).not.toHaveBeenCalled();

        await advance(2000);

        expect(mocks.navigate).toHaveBeenCalledWith("/");
        expect(mocks.redirector).not.toHaveBeenCalled();
    });

    it("redirects to a safe target URL", async () => {
        vi.useFakeTimers();
        mocks.redirectionURL = "https://app.example.com";
        signOutMock.mockResolvedValue({ safeTargetURL: true } as any);

        render(<SignOut />);

        await vi.waitFor(() => expect(signOutMock).toHaveBeenCalled());

        await advance(2000);

        expect(mocks.redirector).toHaveBeenCalledWith("https://app.example.com");
        expect(mocks.navigate).not.toHaveBeenCalled();
    });

    it("navigates to the index route when the target is not safe", async () => {
        vi.useFakeTimers();
        mocks.redirectionURL = "https://evil.example.com";
        signOutMock.mockResolvedValue({ safeTargetURL: false } as any);

        render(<SignOut />);

        await vi.waitFor(() => expect(signOutMock).toHaveBeenCalled());

        await advance(2000);

        expect(mocks.navigate).toHaveBeenCalledWith("/");
        expect(mocks.redirector).not.toHaveBeenCalled();
    });

    it("navigates to the index route when the response is empty", async () => {
        vi.useFakeTimers();
        signOutMock.mockResolvedValue(undefined as any);

        render(<SignOut />);

        await vi.waitFor(() => expect(signOutMock).toHaveBeenCalled());

        await advance(2000);

        expect(mocks.navigate).toHaveBeenCalledWith("/");
    });

    it("restores the redirection URL from the restore parameter", async () => {
        vi.useFakeTimers();
        mocks.searchParams = new URLSearchParams({ rd_restore: "https://app.example.com", rm: "GET" });

        render(<SignOut />);

        await vi.waitFor(() => expect(signOutMock).toHaveBeenCalled());

        await advance(2000);

        expect(mocks.navigate).toHaveBeenCalledWith("/", false, false, false, expect.any(URLSearchParams));

        const search = mocks.navigate.mock.calls[0][4] as URLSearchParams;
        expect(search.get("rd")).toBe("https://app.example.com");
        expect(search.get("rm")).toBe("GET");
        expect(search.has("rd_restore")).toBe(false);
    });
});

describe("failures", () => {
    it("notifies when signing out fails", async () => {
        signOutMock.mockRejectedValue(new Error("boom"));

        render(<SignOut />);

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue signing out"),
        );
        expect(console.error).toHaveBeenCalled();
    });

    it("stays silent when the request is cancelled", async () => {
        signOutMock.mockRejectedValue({ __CANCEL__: true });

        render(<SignOut />);

        await waitFor(() => expect(signOutMock).toHaveBeenCalled());
        expect(mocks.createErrorNotification).not.toHaveBeenCalled();
        expect(console.error).not.toHaveBeenCalled();
    });

    it("does not redirect after unmounting", async () => {
        vi.useFakeTimers();

        const { unmount } = render(<SignOut />);

        await vi.waitFor(() => expect(signOutMock).toHaveBeenCalled());

        unmount();

        await advance(2000);

        expect(mocks.navigate).not.toHaveBeenCalled();
    });
});

describe("confirmation without a flow", () => {
    it("signs out immediately without checking for a pending logout", async () => {
        render(<SignOut />);

        await waitFor(() => expect(signOutMock).toHaveBeenCalledWith(null, expect.anything()));
        expect(getSignOutPendingMock).not.toHaveBeenCalled();
        expect(screen.queryByText(/Are you sure you want to sign out\?/)).not.toBeInTheDocument();
    });

    it("never asks for confirmation even when the server has a logout pending", async () => {
        getSignOutPendingMock.mockResolvedValue({ clientID: "app", pending: true });

        render(<SignOut />);

        await waitFor(() => expect(signOutMock).toHaveBeenCalled());
        expect(screen.queryByText("Cancel")).not.toBeInTheDocument();
    });
});

describe("confirmation with a flow", () => {
    beforeEach(() => {
        mocks.searchParams = new URLSearchParams({ flow_id: "flow-1" });
    });

    it("signs out immediately when the flow's logout is no longer pending", async () => {
        render(<SignOut />);

        await waitFor(() => expect(signOutMock).toHaveBeenCalledWith(null, expect.anything()));
        expect(getSignOutPendingMock).toHaveBeenCalledWith("flow-1", expect.anything());
        expect(screen.queryByText(/Are you sure you want to sign out\?/)).not.toBeInTheDocument();
    });

    it("does not sign out before it knows whether the logout is pending", async () => {
        getSignOutPendingMock.mockReturnValue(new Promise(() => {}));

        render(<SignOut />);

        await waitFor(() => expect(getSignOutPendingMock).toHaveBeenCalled());
        expect(signOutMock).not.toHaveBeenCalled();
        expect(screen.queryByText("Sign out")).not.toBeInTheDocument();
    });

    it("asks for confirmation when the logout is pending", async () => {
        getSignOutPendingMock.mockResolvedValue({ clientID: "app", clientName: "My App", pending: true });

        render(<SignOut />);

        expect(await screen.findByText(/Are you sure you want to sign out\?/)).toBeInTheDocument();
        expect(screen.getByText(/My App has requested that you sign out/)).toBeInTheDocument();
        expect(signOutMock).not.toHaveBeenCalled();
    });

    it("names the client by its identifier when it has no name", async () => {
        getSignOutPendingMock.mockResolvedValue({ clientID: "app", pending: true });

        render(<SignOut />);

        expect(await screen.findByText(/app has requested that you sign out/)).toBeInTheDocument();
    });

    it("asks for confirmation when it can't determine whether the logout is pending", async () => {
        getSignOutPendingMock.mockRejectedValue(new Error("boom"));

        render(<SignOut />);

        expect(await screen.findByText(/Are you sure you want to sign out\?/)).toBeInTheDocument();
        expect(signOutMock).not.toHaveBeenCalled();
    });

    it("signs out with the flow once the confirmation is accepted", async () => {
        getSignOutPendingMock.mockResolvedValue({ clientID: "app", pending: true });

        render(<SignOut />);

        const button = await screen.findByText("Sign out");

        await act(async () => {
            fireEvent.click(button);
        });

        await waitFor(() => expect(signOutMock).toHaveBeenCalledWith(null, expect.anything(), "flow-1"));
        expect(screen.getByText(/You're being signed out and redirected/)).toBeInTheDocument();
    });

    it("redirects to the URL the server returns once signed out", async () => {
        vi.useFakeTimers();
        getSignOutPendingMock.mockResolvedValue({ clientID: "app", pending: true });
        signOutMock.mockResolvedValue({
            redirectURL: "https://app.example.com/logged-out?state=abc",
            safeTargetURL: false,
        });

        render(<SignOut />);

        const button = await vi.waitFor(() => screen.getByText("Sign out"));

        await act(async () => {
            fireEvent.click(button);
        });

        await vi.waitFor(() => expect(signOutMock).toHaveBeenCalled());

        await advance(2000);

        expect(mocks.redirector).toHaveBeenCalledWith("https://app.example.com/logged-out?state=abc");
        expect(mocks.navigate).not.toHaveBeenCalled();
    });

    it("cancels the flow's logout and stays signed in when the confirmation is rejected", async () => {
        getSignOutPendingMock.mockResolvedValue({ clientID: "app", pending: true });

        render(<SignOut />);

        const button = await screen.findByText("Cancel");

        await act(async () => {
            fireEvent.click(button);
        });

        await waitFor(() => expect(cancelSignOutMock).toHaveBeenCalledWith("flow-1"));
        expect(signOutMock).not.toHaveBeenCalled();
        expect(mocks.navigate).toHaveBeenCalledWith("/");
        expect(mocks.redirector).not.toHaveBeenCalled();
    });

    it("notifies when cancelling fails but still leaves the page", async () => {
        getSignOutPendingMock.mockResolvedValue({ clientID: "app", pending: true });
        cancelSignOutMock.mockRejectedValue(new Error("boom"));

        render(<SignOut />);

        const button = await screen.findByText("Cancel");

        await act(async () => {
            fireEvent.click(button);
        });

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue cancelling the sign out"),
        );
        expect(mocks.navigate).toHaveBeenCalledWith("/");
    });
});
