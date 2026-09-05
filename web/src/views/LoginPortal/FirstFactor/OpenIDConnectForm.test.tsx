import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import OpenIDConnectForm from "@views/LoginPortal/FirstFactor/OpenIDConnectForm";

const getOpenIDConnectProviders = vi.fn();
const postOpenIDConnectStart = vi.fn();

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@services/OpenIDConnectRelyingParty", () => ({
    getOpenIDConnectProviders: (...args: unknown[]) => getOpenIDConnectProviders(...args),
    postOpenIDConnectStart: (...args: unknown[]) => postOpenIDConnectStart(...args),
}));

beforeEach(() => {
    getOpenIDConnectProviders.mockReset();
    postOpenIDConnectStart.mockReset();
});

it("renders a button per provider", async () => {
    getOpenIDConnectProviders.mockResolvedValue([
        { id: "google", name: "Google" },
        { id: "github", name: "GitHub" },
    ]);

    render(<OpenIDConnectForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());
    expect(screen.getByText("Sign in with GitHub")).toBeInTheDocument();
});

it("renders nothing when no providers are returned", async () => {
    getOpenIDConnectProviders.mockResolvedValue([]);

    const { container } = render(<OpenIDConnectForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(getOpenIDConnectProviders).toHaveBeenCalled());
    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when the provider request fails", async () => {
    getOpenIDConnectProviders.mockRejectedValue(new Error("network error"));

    const { container } = render(<OpenIDConnectForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(getOpenIDConnectProviders).toHaveBeenCalled());
    expect(container).toBeEmptyDOMElement();
});

it("disables the buttons when disabled", async () => {
    getOpenIDConnectProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    render(<OpenIDConnectForm disabled={true} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());
    expect(screen.getByText("Sign in with Google").closest("button")).toBeDisabled();
});

it("navigates to the authorization url on click", async () => {
    getOpenIDConnectProviders.mockResolvedValue([{ id: "google", name: "Google" }]);
    postOpenIDConnectStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize?x=1" });

    const assign = vi.fn();

    vi.stubGlobal("location", { ...window.location, assign });

    render(<OpenIDConnectForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Sign in with Google"));

    await waitFor(() => expect(assign).toHaveBeenCalledWith("https://op.example.com/authorize?x=1"));

    vi.unstubAllGlobals();
});

it("disables the clicked button while the request is in flight", async () => {
    getOpenIDConnectProviders.mockResolvedValue([{ id: "google", name: "Google" }]);

    let resolveStart: (value: { authorization_url: string }) => void = () => {};
    postOpenIDConnectStart.mockImplementation(
        () =>
            new Promise((resolve) => {
                resolveStart = resolve;
            }),
    );

    render(<OpenIDConnectForm disabled={false} rememberMe={false} />);

    await waitFor(() => expect(screen.getByText("Sign in with Google")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Sign in with Google"));

    await waitFor(() => expect(screen.getByText("Sign in with Google").closest("button")).toBeDisabled());

    resolveStart({ authorization_url: "https://op.example.com/authorize?x=1" });
});
