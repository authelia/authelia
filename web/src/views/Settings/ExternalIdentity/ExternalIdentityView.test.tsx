// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";

import ExternalIdentityView from "@views/Settings/ExternalIdentity/ExternalIdentityView";

const getExternalIdentityLinks = vi.fn();
const getExternalIdentityProviders = vi.fn();
const postExternalIdentityStart = vi.fn();

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

vi.mock("@services/ExternalIdentity", () => ({
    deleteExternalIdentityLink: vi.fn(),
    deleteExternalIdentityLinkPending: vi.fn(),
    getExternalIdentityLinks: (...args: unknown[]) => getExternalIdentityLinks(...args),
    getExternalIdentityProviders: (...args: unknown[]) => getExternalIdentityProviders(...args),
    postExternalIdentityStart: (...args: unknown[]) => postExternalIdentityStart(...args),
    putExternalIdentityLink: vi.fn(),
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

vi.mock("@views/Settings/ExternalIdentity/ExternalIdentityLinkDeleteDialog", () => ({
    default: () => <div data-testid="delete-dialog" />,
}));

beforeEach(() => {
    getExternalIdentityLinks.mockReset();
    getExternalIdentityProviders.mockReset();
    getExternalIdentityProviders.mockResolvedValue([]);
    postExternalIdentityStart.mockReset();
});

it("renders the links returned by the service", async () => {
    getExternalIdentityLinks.mockResolvedValue({
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

    render(<ExternalIdentityView />);

    await waitFor(() => expect(screen.getByText("Google")).toBeInTheDocument());
});

it("renders the pending panel when pending is present", async () => {
    getExternalIdentityLinks.mockResolvedValue({
        links: [],
        pending: {
            issuer: "https://accounts.google.com",
            provider: "google",
            provider_name: "Google",
            subject: "abc123",
        },
    });

    render(<ExternalIdentityView />);

    await waitFor(() => expect(screen.getByText("Link your {{name}} account")).toBeInTheDocument());
});

it("does not render the pending panel when it is absent", async () => {
    getExternalIdentityLinks.mockResolvedValue({ links: [] });

    render(<ExternalIdentityView />);

    await waitFor(() => expect(screen.getByText("No external accounts are linked")).toBeInTheDocument());

    expect(screen.queryByText("Link your {{name}} account")).not.toBeInTheDocument();
});

it("offers a link action for a configured provider which is not linked", async () => {
    getExternalIdentityLinks.mockResolvedValue({ links: [] });
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    render(<ExternalIdentityView />);

    await waitFor(() => expect(document.getElementById("external-identity-link-start-google")).toBeInTheDocument());
});

it("does not offer a link action for a provider which is already linked", async () => {
    getExternalIdentityLinks.mockResolvedValue({
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
    getExternalIdentityProviders.mockResolvedValue([
        { id: "google", name: "Google" },
        { id: "github", name: "GitHub" },
    ]);

    render(<ExternalIdentityView />);

    await waitFor(() => expect(document.getElementById("external-identity-link-start-github")).toBeInTheDocument());

    expect(document.getElementById("external-identity-link-start-google")).not.toBeInTheDocument();
});

it("does not offer a link action for a provider with a pending proposal", async () => {
    getExternalIdentityLinks.mockResolvedValue({
        links: [],
        pending: {
            issuer: "https://accounts.google.com",
            provider: "google",
            provider_name: "Google",
            subject: "abc123",
        },
    });
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    render(<ExternalIdentityView />);

    await waitFor(() => expect(screen.getByText("No external accounts are linked")).toBeInTheDocument());

    expect(document.getElementById("external-identity-link-start-google")).not.toBeInTheDocument();
});

it("navigates to the authorization url when a link action is used", async () => {
    getExternalIdentityLinks.mockResolvedValue({ links: [] });
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);
    postExternalIdentityStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize?x=1" });

    const assign = vi.fn();

    vi.stubGlobal("location", { ...window.location, assign });

    render(<ExternalIdentityView />);

    await waitFor(() => expect(document.getElementById("external-identity-link-start-google")).toBeInTheDocument());

    fireEvent.click(document.getElementById("external-identity-link-start-google")!);

    await waitFor(() => expect(assign).toHaveBeenCalledWith("https://op.example.com/authorize?x=1"));

    expect(postExternalIdentityStart).toHaveBeenCalledWith("google", { keepMeLoggedIn: false });

    vi.unstubAllGlobals();
});
