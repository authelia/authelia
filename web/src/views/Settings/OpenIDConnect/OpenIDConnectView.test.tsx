import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";

import OpenIDConnectView from "@views/Settings/OpenIDConnect/OpenIDConnectView";

const getOpenIDConnectLinks = vi.fn();
const getOpenIDConnectProviders = vi.fn();
const postOpenIDConnectStart = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        createSuccessNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoPOST: () => [undefined, vi.fn(), false, undefined],
}));

vi.mock("@services/OpenIDConnectRelyingParty", () => ({
    deleteOpenIDConnectLink: vi.fn(),
    deleteOpenIDConnectLinkPending: vi.fn(),
    getOpenIDConnectLinks: (...args: unknown[]) => getOpenIDConnectLinks(...args),
    getOpenIDConnectProviders: (...args: unknown[]) => getOpenIDConnectProviders(...args),
    postOpenIDConnectStart: (...args: unknown[]) => postOpenIDConnectStart(...args),
    putOpenIDConnectLink: vi.fn(),
}));

vi.mock("@services/UserSessionElevation", () => ({
    getUserSessionElevation: vi.fn(),
}));

vi.mock("@views/Settings/Common/IdentityVerificationDialog", () => ({
    default: () => <div data-testid="identity-dialog" />,
}));

vi.mock("@views/Settings/Common/SecondFactorDialog", () => ({
    default: () => <div data-testid="second-factor-dialog" />,
}));

vi.mock("@views/Settings/OpenIDConnect/OpenIDConnectLinkDeleteDialog", () => ({
    default: () => <div data-testid="delete-dialog" />,
}));

beforeEach(() => {
    getOpenIDConnectLinks.mockReset();
    getOpenIDConnectProviders.mockReset();
    getOpenIDConnectProviders.mockResolvedValue([]);
    postOpenIDConnectStart.mockReset();
});

it("renders the links returned by the service", async () => {
    getOpenIDConnectLinks.mockResolvedValue({
        links: [
            {
                created_at: "2024-01-01T00:00:00Z",
                id: 1,
                issuer: "https://accounts.google.com",
                provider: "google",
                provider_name: "Google",
                remote_username: "john",
                subject: "abc123",
            },
        ],
    });

    render(<OpenIDConnectView />);

    await waitFor(() => expect(screen.getByText("Google")).toBeInTheDocument());
});

it("renders the pending panel when pending is present", async () => {
    getOpenIDConnectLinks.mockResolvedValue({
        links: [],
        pending: {
            issuer: "https://accounts.google.com",
            provider: "google",
            provider_name: "Google",
            subject: "abc123",
        },
    });

    render(<OpenIDConnectView />);

    await waitFor(() => expect(screen.getByText("Link your {{name}} account")).toBeInTheDocument());
});

it("does not render the pending panel when it is absent", async () => {
    getOpenIDConnectLinks.mockResolvedValue({ links: [] });

    render(<OpenIDConnectView />);

    await waitFor(() => expect(screen.getByText("No external accounts are linked")).toBeInTheDocument());

    expect(screen.queryByText("Link your {{name}} account")).not.toBeInTheDocument();
});

it("offers a link action for a configured provider which is not linked", async () => {
    getOpenIDConnectLinks.mockResolvedValue({ links: [] });
    getOpenIDConnectProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    render(<OpenIDConnectView />);

    await waitFor(() => expect(document.getElementById("openid-connect-link-start-google")).toBeInTheDocument());
});

it("does not offer a link action for a provider which is already linked", async () => {
    getOpenIDConnectLinks.mockResolvedValue({
        links: [
            {
                created_at: "2024-01-01T00:00:00Z",
                id: 1,
                issuer: "https://accounts.google.com",
                provider: "google",
                provider_name: "Google",
                subject: "abc123",
            },
        ],
    });
    getOpenIDConnectProviders.mockResolvedValue([
        { id: "google", name: "Google" },
        { id: "github", name: "GitHub" },
    ]);

    render(<OpenIDConnectView />);

    await waitFor(() => expect(document.getElementById("openid-connect-link-start-github")).toBeInTheDocument());

    expect(document.getElementById("openid-connect-link-start-google")).not.toBeInTheDocument();
});

it("does not offer a link action for a provider with a pending proposal", async () => {
    getOpenIDConnectLinks.mockResolvedValue({
        links: [],
        pending: {
            issuer: "https://accounts.google.com",
            provider: "google",
            provider_name: "Google",
            subject: "abc123",
        },
    });
    getOpenIDConnectProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    render(<OpenIDConnectView />);

    await waitFor(() => expect(screen.getByText("No external accounts are linked")).toBeInTheDocument());

    expect(document.getElementById("openid-connect-link-start-google")).not.toBeInTheDocument();
});

it("navigates to the authorization url when a link action is used", async () => {
    getOpenIDConnectLinks.mockResolvedValue({ links: [] });
    getOpenIDConnectProviders.mockResolvedValue([{ id: "google", name: "Google" }]);
    postOpenIDConnectStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize?x=1" });

    const assign = vi.fn();

    vi.stubGlobal("location", { ...window.location, assign });

    render(<OpenIDConnectView />);

    await waitFor(() => expect(document.getElementById("openid-connect-link-start-google")).toBeInTheDocument());

    fireEvent.click(document.getElementById("openid-connect-link-start-google")!);

    await waitFor(() => expect(assign).toHaveBeenCalledWith("https://op.example.com/authorize?x=1"));

    expect(postOpenIDConnectStart).toHaveBeenCalledWith("google", { keepMeLoggedIn: false });

    vi.unstubAllGlobals();
});
