// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import ExternalIdentityLinkNotice from "@views/LoginPortal/FirstFactor/ExternalIdentityLinkNotice";

const mocks = vi.hoisted(() => ({
    getExternalIdentityProviders: vi.fn(),
    navigate: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({
        t: (key: string, options?: { name?: string }) =>
            options?.name ? key.replaceAll("{{name}}", options.name) : key,
    }),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mocks.navigate,
}));

vi.mock("@services/ExternalIdentity", () => ({
    getExternalIdentityProviders: (...args: unknown[]) => mocks.getExternalIdentityProviders(...args),
}));

beforeEach(() => {
    mocks.getExternalIdentityProviders.mockReset();
    mocks.navigate.mockReset();
});

it("names the provider once the providers have loaded", async () => {
    mocks.getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    render(<ExternalIdentityLinkNotice provider="google" />);

    expect(await screen.findByText("Link your Google account")).toBeInTheDocument();
    expect(
        screen.getByText(
            "This Google account is not linked yet, sign in to link it and you will then return to Google to confirm the link",
            { exact: false },
        ),
    ).toBeInTheDocument();
});

it("falls back to the provider id when the provider is unknown", async () => {
    mocks.getExternalIdentityProviders.mockResolvedValue([{ id: "other", name: "Other" }]);

    render(<ExternalIdentityLinkNotice provider="google" />);

    expect(await screen.findByText("Link your google account")).toBeInTheDocument();
});

it("falls back to the provider id when the providers cannot be loaded", async () => {
    mocks.getExternalIdentityProviders.mockRejectedValue(new Error("failed"));

    render(<ExternalIdentityLinkNotice provider="google" />);

    expect(await screen.findByText("Link your google account")).toBeInTheDocument();
});

it("returns to the login page without the link", async () => {
    mocks.getExternalIdentityProviders.mockResolvedValue([]);

    render(<ExternalIdentityLinkNotice provider="google" />);

    fireEvent.click(await screen.findByText("Use another sign in method"));

    expect(mocks.navigate).toHaveBeenCalledWith("/", false);
});
