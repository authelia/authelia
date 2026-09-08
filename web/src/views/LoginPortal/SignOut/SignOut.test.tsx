import { act, render, screen, waitFor } from "@testing-library/react";

import { signOut } from "@services/SignOut";
import SignOut from "@views/LoginPortal/SignOut/SignOut";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    navigate: vi.fn(),
    redirectionURL: null as null | string,
    redirector: vi.fn(),
    searchParams: new URLSearchParams(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-router", () => ({
    useSearchParams: () => [mocks.searchParams, vi.fn()],
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

vi.mock("@constants/SearchParams", () => ({
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
    signOut: vi.fn(),
}));

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
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders sign out message", async () => {
        render(<SignOut />);

        expect(screen.getByText(/You're being signed out and redirected/)).toBeInTheDocument();
        await waitFor(() => expect(signOutMock).toHaveBeenCalled());
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
