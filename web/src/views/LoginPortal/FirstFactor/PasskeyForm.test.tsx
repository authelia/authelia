import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { getWebAuthnPasskeyOptions, getWebAuthnResult, postWebAuthnPasskeyResponse } from "@services/WebAuthn";
import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: null, id: null, subflow: null }),
}));

vi.mock("@models/WebAuthn", () => ({
    AssertionResult: { Success: 0 },
    AssertionResultFailureString: () => "failure",
}));

vi.mock("@services/WebAuthn", () => ({
    getWebAuthnPasskeyOptions: vi.fn(),
    getWebAuthnResult: vi.fn(),
    postWebAuthnPasskeyResponse: vi.fn(),
}));

vi.mock("@components/PasskeyIcon", () => ({
    default: () => <span data-testid="passkey-icon" />,
}));

vi.mock("@simplewebauthn/browser", () => ({
    browserSupportsWebAuthnAutofill: vi.fn().mockResolvedValue(false),
}));

const mockGetWebAuthnPasskeyOptions = vi.mocked(getWebAuthnPasskeyOptions);
const mockGetWebAuthnResult = vi.mocked(getWebAuthnResult);
const mockPostWebAuthnPasskeyResponse = vi.mocked(postWebAuthnPasskeyResponse);

function mockSuccessfulAssertion() {
    mockGetWebAuthnPasskeyOptions.mockResolvedValue({ options: {}, status: 200 } as never);
    mockGetWebAuthnResult.mockResolvedValue({ response: {}, result: 0 } as never);
    mockPostWebAuthnPasskeyResponse.mockResolvedValue({
        data: { data: { redirect: undefined }, status: "OK" },
        status: 200,
    } as never);
}

function renderForm(rememberMe: boolean) {
    return render(
        <PasskeyForm
            disabled={false}
            rememberMe={rememberMe}
            onAuthenticationStart={vi.fn()}
            onAuthenticationStop={vi.fn()}
            onAuthenticationError={vi.fn()}
            onAuthenticationSuccess={vi.fn()}
        />,
    );
}

beforeEach(() => {
    vi.clearAllMocks();
});

it("renders passkey sign in button", () => {
    renderForm(false);
    expect(screen.getByText("Sign in with a passkey")).toBeInTheDocument();
    expect(screen.getByText("or")).toBeInTheDocument();
});

it("renders button as disabled when disabled prop is true", () => {
    render(
        <PasskeyForm
            disabled={true}
            rememberMe={false}
            onAuthenticationStart={vi.fn()}
            onAuthenticationStop={vi.fn()}
            onAuthenticationError={vi.fn()}
            onAuthenticationSuccess={vi.fn()}
        />,
    );
    expect(screen.getByText("Sign in with a passkey").closest("button")).toBeDisabled();
});

it("does not prompt and posts false when remember me is disabled", async () => {
    mockSuccessfulAssertion();
    renderForm(false);

    fireEvent.click(screen.getByText("Sign in with a passkey"));

    await waitFor(() => expect(mockPostWebAuthnPasskeyResponse).toHaveBeenCalled());
    expect(screen.queryByText("Remember me?")).not.toBeInTheDocument();
    expect(mockPostWebAuthnPasskeyResponse.mock.calls[0][1]).toBe(false);
});

it("posts true when remember me is enabled and the user answers yes", async () => {
    mockSuccessfulAssertion();
    renderForm(true);

    fireEvent.click(screen.getByText("Sign in with a passkey"));

    await waitFor(() => expect(screen.getByText("Remember me?")).toBeInTheDocument());
    expect(mockPostWebAuthnPasskeyResponse).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Yes" }));

    await waitFor(() => expect(mockPostWebAuthnPasskeyResponse).toHaveBeenCalled());
    expect(mockPostWebAuthnPasskeyResponse.mock.calls[0][1]).toBe(true);
});

it("posts false when remember me is enabled and the user answers no", async () => {
    mockSuccessfulAssertion();
    renderForm(true);

    fireEvent.click(screen.getByText("Sign in with a passkey"));

    await waitFor(() => expect(screen.getByText("Remember me?")).toBeInTheDocument());

    fireEvent.click(screen.getByRole("button", { name: "No" }));

    await waitFor(() => expect(mockPostWebAuthnPasskeyResponse).toHaveBeenCalled());
    expect(mockPostWebAuthnPasskeyResponse.mock.calls[0][1]).toBe(false);
});

it("posts false and completes when the prompt is dismissed", async () => {
    mockSuccessfulAssertion();
    renderForm(true);

    fireEvent.click(screen.getByText("Sign in with a passkey"));

    await waitFor(() => expect(screen.getByText("Remember me?")).toBeInTheDocument());

    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });

    await waitFor(() => expect(mockPostWebAuthnPasskeyResponse).toHaveBeenCalled());
    expect(mockPostWebAuthnPasskeyResponse.mock.calls[0][1]).toBe(false);
});
