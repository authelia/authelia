import { act, fireEvent, render, screen } from "@testing-library/react";

import SignOut from "@views/LoginPortal/SignOut/SignOut";

let mockSearchParams: URLSearchParams = new URLSearchParams();
let mockRedirectionURL: null | string = null;

const mockNavigate = vi.fn();
const mockRedirector = vi.fn();
const mockCreateError = vi.fn();
const mockSignOut = vi.fn();
const mockCheckSafeRedirection = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-router-dom", () => ({
    useSearchParams: () => [mockSearchParams, vi.fn()],
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

vi.mock("@constants/SearchParams", () => ({
    Confirm: "confirm",
    RedirectionRestoreURL: "rd_restore",
    RedirectionURL: "rd",
    State: "state",
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({ createErrorNotification: mockCreateError }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => mockRedirectionURL,
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => mockRedirector,
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => <div data-testid="layout">{props.children}</div>,
}));

vi.mock("@services/SafeRedirection", () => ({
    checkSafeRedirection: (...args: any[]) => mockCheckSafeRedirection(...args),
}));

vi.mock("@services/SignOut", () => ({
    signOut: (...args: any[]) => mockSignOut(...args),
}));

beforeEach(() => {
    vi.useFakeTimers();
    vi.spyOn(console, "log").mockImplementation(() => {});
    mockSearchParams = new URLSearchParams();
    mockRedirectionURL = null;
    mockNavigate.mockReset();
    mockRedirector.mockReset();
    mockCreateError.mockReset();
    mockSignOut.mockReset().mockResolvedValue({ safeTargetURL: false });
    mockCheckSafeRedirection.mockReset().mockResolvedValue({ ok: false });
});

afterEach(() => {
    vi.useRealTimers();
});

const renderAndSettle = async () => {
    await act(async () => {
        render(<SignOut />);
    });
};

const advanceRedirectTimeout = async () => {
    await act(async () => {
        vi.advanceTimersByTime(2000);
    });
};

it("renders the sign out message by default", async () => {
    await renderAndSettle();
    expect(screen.getByText(/You're being signed out and redirected/)).toBeInTheDocument();
});

it("scenario 1 (yes): logs out and redirects to safe target URL with state", async () => {
    mockRedirectionURL = "https://example.com/app";
    mockSearchParams = new URLSearchParams("rd=https://example.com/app&confirm=true&state=abc123");
    mockCheckSafeRedirection.mockResolvedValue({ ok: true });

    await renderAndSettle();

    expect(mockCheckSafeRedirection).toHaveBeenCalledWith("https://example.com/app");
    expect(screen.getByText("Are you sure you want to sign out?")).toBeInTheDocument();

    await act(async () => {
        fireEvent.click(screen.getByText("Yes"));
    });

    expect(mockSignOut).toHaveBeenCalledTimes(1);
    expect(mockSignOut.mock.calls[0][0]).toBe("https://example.com/app");

    await advanceRedirectTimeout();

    expect(mockRedirector).toHaveBeenCalledTimes(1);
    expect(mockRedirector.mock.calls[0][0]).toBe("https://example.com/app?state=abc123");
    expect(mockNavigate).not.toHaveBeenCalled();
});

it("scenario 1 (no): redirects to safe target URL with state without signing out", async () => {
    mockRedirectionURL = "https://example.com/app";
    mockSearchParams = new URLSearchParams("rd=https://example.com/app&confirm=true&state=abc123");
    mockCheckSafeRedirection.mockResolvedValue({ ok: true });

    await renderAndSettle();

    expect(screen.getByText("Are you sure you want to sign out?")).toBeInTheDocument();

    await act(async () => {
        fireEvent.click(screen.getByText("No"));
    });

    expect(mockSignOut).not.toHaveBeenCalled();
    expect(mockRedirector).toHaveBeenCalledTimes(1);
    expect(mockRedirector.mock.calls[0][0]).toBe("https://example.com/app?state=abc123");
    expect(mockNavigate).not.toHaveBeenCalled();
});

it("scenario 2: signs out without safety check and does not propagate state when redirect URL absent", async () => {
    mockRedirectionURL = null;
    mockSearchParams = new URLSearchParams("state=abc123");

    await renderAndSettle();

    expect(mockCheckSafeRedirection).not.toHaveBeenCalled();
    expect(mockSignOut).toHaveBeenCalledTimes(1);
    expect(mockSignOut.mock.calls[0][0]).toBeFalsy();

    await advanceRedirectTimeout();

    expect(mockRedirector).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledTimes(1);
    expect(mockNavigate.mock.calls[0][0]).toBe("/");
});

it("scenario 3 (yes): no redirect URL with confirm=true asks then signs out and goes to index", async () => {
    mockRedirectionURL = null;
    mockSearchParams = new URLSearchParams("confirm=true");

    await renderAndSettle();

    expect(mockCheckSafeRedirection).not.toHaveBeenCalled();
    expect(screen.getByText("Are you sure you want to sign out?")).toBeInTheDocument();

    await act(async () => {
        fireEvent.click(screen.getByText("Yes"));
    });

    expect(mockSignOut).toHaveBeenCalledTimes(1);

    await advanceRedirectTimeout();

    expect(mockRedirector).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledTimes(1);
    expect(mockNavigate.mock.calls[0][0]).toBe("/");
});

it("scenario 3 (no): no redirect URL with confirm=true asks then goes to index without signing out", async () => {
    mockRedirectionURL = null;
    mockSearchParams = new URLSearchParams("confirm=true");

    await renderAndSettle();

    expect(screen.getByText("Are you sure you want to sign out?")).toBeInTheDocument();

    await act(async () => {
        fireEvent.click(screen.getByText("No"));
    });

    expect(mockSignOut).not.toHaveBeenCalled();
    expect(mockRedirector).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledTimes(1);
    expect(mockNavigate.mock.calls[0][0]).toBe("/");
});

it("scenario 4: safe redirect URL without confirm or state signs out and redirects unconditionally", async () => {
    mockRedirectionURL = "https://example.com/app";
    mockSearchParams = new URLSearchParams("rd=https://example.com/app");
    mockCheckSafeRedirection.mockResolvedValue({ ok: true });

    await renderAndSettle();

    expect(screen.queryByText("Are you sure you want to sign out?")).not.toBeInTheDocument();
    expect(mockSignOut).toHaveBeenCalledTimes(1);
    expect(mockSignOut.mock.calls[0][0]).toBe("https://example.com/app");

    await advanceRedirectTimeout();

    expect(mockRedirector).toHaveBeenCalledTimes(1);
    expect(mockRedirector.mock.calls[0][0]).toBe("https://example.com/app");
    expect(mockNavigate).not.toHaveBeenCalled();
});
