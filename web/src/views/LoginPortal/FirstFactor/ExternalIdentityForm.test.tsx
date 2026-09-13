// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import ExternalIdentityForm from "@views/LoginPortal/FirstFactor/ExternalIdentityForm";

const getExternalIdentityProviders = vi.fn();
const postExternalIdentityStart = vi.fn();
const createErrorNotification = vi.fn();

vi.mock("@views/LoginPortal/FirstFactor/ExternalIdentityIcon", () => ({
    default: (props: any) => <div data-testid={`icon-${props.type}`} data-logo-uri={props.logoURI ?? ""} />,
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({ createErrorNotification }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

const flowParameters: { flow?: string; id?: string; subflow?: string } = {};
let userCode: string | undefined;

vi.mock("@hooks/Flow", () => ({
    useFlow: () => flowParameters,
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => userCode,
}));

vi.mock("@services/ExternalIdentity", () => ({
    getExternalIdentityProviders: (...args: unknown[]) => getExternalIdentityProviders(...args),
    postExternalIdentityStart: (...args: unknown[]) => postExternalIdentityStart(...args),
}));

beforeEach(() => {
    getExternalIdentityProviders.mockReset();
    postExternalIdentityStart.mockReset();
    createErrorNotification.mockReset();
});

it("shows an error and enables the buttons again when the sign in cannot be started", async () => {
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);
    postExternalIdentityStart.mockRejectedValue(new Error("request failed"));

    const assign = vi.fn();

    vi.stubGlobal("location", { ...window.location, assign });

    render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Sign in with Google"));

    await waitFor(() =>
        expect(createErrorNotification).toHaveBeenCalledWith(
            "There was an issue signing in with the external provider",
        ),
    );

    expect(screen.getByText("Sign in with Google").closest("button")).toBeEnabled();
    expect(assign).not.toHaveBeenCalled();

    vi.unstubAllGlobals();
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

it("reports the loading state while the request is in flight and after it fails", async () => {
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);
    postExternalIdentityStart.mockRejectedValue(new Error("request failed"));

    const onLoadingChange = vi.fn();

    render(<ExternalIdentityForm disabled={false} rememberMe={false} onLoadingChange={onLoadingChange} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Sign in with Google"));

    await waitFor(() => expect(onLoadingChange).toHaveBeenLastCalledWith(false));
    expect(onLoadingChange).toHaveBeenNthCalledWith(1, true);
});

it("sends the flow the sign in was started from", async () => {
    getExternalIdentityProviders.mockResolvedValue([{ id: "google", name: "Google" }]);
    postExternalIdentityStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize?x=1" });

    flowParameters.flow = "openid_connect";
    flowParameters.id = "abc";
    flowParameters.subflow = "device_authorization";
    userCode = "ABCD-EFGH";

    vi.stubGlobal("location", { ...window.location, assign: vi.fn() });

    render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Sign in with Google"));

    await waitFor(() =>
        expect(postExternalIdentityStart).toHaveBeenCalledWith(
            "google",
            expect.objectContaining({
                flow: "openid_connect",
                flowID: "abc",
                subflow: "device_authorization",
                userCode: "ABCD-EFGH",
            }),
            expect.anything(),
        ),
    );

    delete flowParameters.flow;
    delete flowParameters.id;
    delete flowParameters.subflow;
    userCode = undefined;
});

it("renders an icon for each provider carrying its type and logo", async () => {
    getExternalIdentityProviders.mockResolvedValue([
        { id: "google", logo_uri: "https://cdn.example.com/google.png", name: "Google", type: "openid_connect" },
        { id: "discord", name: "Discord", type: "discord" },
    ]);

    render(<ExternalIdentityForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    expect(screen.getByTestId("icon-openid_connect")).toHaveAttribute(
        "data-logo-uri",
        "https://cdn.example.com/google.png",
    );
    expect(screen.getByTestId("icon-discord")).toHaveAttribute("data-logo-uri", "");
});
