// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import ExternalIdentityForm from "@views/LoginPortal/FirstFactor/ExternalIdentityForm";

const getExternalIdentityProviders = vi.fn();
const postExternalIdentityStart = vi.fn();

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@services/ExternalIdentity", () => ({
    getExternalIdentityProviders: (...args: unknown[]) => getExternalIdentityProviders(...args),
    postExternalIdentityStart: (...args: unknown[]) => postExternalIdentityStart(...args),
}));

beforeEach(() => {
    getExternalIdentityProviders.mockReset();
    postExternalIdentityStart.mockReset();
});

it("renders a button per provider", async () => {
    getExternalIdentityProviders.mockResolvedValue([
        { id: "google", name: "Google" },
        { id: "github", name: "GitHub" },
    ]);

    render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());
    expect(screen.getByText("Sign in with GitHub")).toBeInTheDocument();
});

it("renders nothing when no providers are returned", async () => {
    getExternalIdentityProviders.mockResolvedValue([]);

    const { container } = render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(getExternalIdentityProviders).toHaveBeenCalled());
    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when the provider request fails", async () => {
    getExternalIdentityProviders.mockRejectedValue(new Error("network error"));

    const { container } = render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(getExternalIdentityProviders).toHaveBeenCalled());
    expect(container).toBeEmptyDOMElement();
});

it("disables the buttons when disabled", async () => {
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    render(<ExternalIdentityForm disabled={true} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());
    expect(screen.getByText("Sign in with Google").closest("button")).toBeDisabled();
});

it("navigates to the authorization url on click", async () => {
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);
    postExternalIdentityStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize?x=1" });

    const assign = vi.fn();

    vi.stubGlobal("location", { ...window.location, assign });

    render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Sign in with Google"));

    await waitFor(() => expect(assign).toHaveBeenCalledWith("https://op.example.com/authorize?x=1"));

    vi.unstubAllGlobals();
});

it("disables the clicked button while the request is in flight", async () => {
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    let resolveStart: (value: { authorization_url: string }) => void = () => {};
    postExternalIdentityStart.mockImplementation(
        () =>
            new Promise((resolve) => {
                resolveStart = resolve;
            }),
    );

    render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Sign in with Google"));

    await waitFor(() => expect(screen.getByText("Sign in with Google").closest("button")).toBeDisabled());

    resolveStart({ authorization_url: "https://op.example.com/authorize?x=1" });
});
