import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { AssertionResult } from "@models/WebAuthn";
import { getWebAuthnPasskeyOptions, getWebAuthnResult, postWebAuthnPasskeyResponse } from "@services/WebAuthn";
import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";

const mocks = vi.hoisted(() => ({
    queryParams: {} as Record<string, null | string>,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: (name: string) => mocks.queryParams[name] ?? null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: "flow", id: "flow-id", subflow: "subflow" }),
}));

vi.mock("@services/WebAuthn", () => ({
    getWebAuthnPasskeyOptions: vi.fn(),
    getWebAuthnResult: vi.fn(),
    postWebAuthnPasskeyResponse: vi.fn(),
}));

vi.mock("@components/PasskeyIcon", () => ({
    default: () => <span data-testid="passkey-icon" />,
}));

const getOptionsMock = vi.mocked(getWebAuthnPasskeyOptions);
const getResultMock = vi.mocked(getWebAuthnResult);
const postResponseMock = vi.mocked(postWebAuthnPasskeyResponse);

const options = { challenge: "challenge" } as any;
const assertionResponse = { id: "credential-id" };

const defaultProps = {
    disabled: false,
    onAuthenticationError: vi.fn(),
    onAuthenticationStart: vi.fn(),
    onAuthenticationStop: vi.fn(),
    onAuthenticationSuccess: vi.fn(),
    rememberMe: false,
};

function renderForm(props: Partial<typeof defaultProps> = {}) {
    const merged = {
        ...defaultProps,
        onAuthenticationError: vi.fn(),
        onAuthenticationStart: vi.fn(),
        onAuthenticationStop: vi.fn(),
        onAuthenticationSuccess: vi.fn(),
        ...props,
    };

    return { ...render(<PasskeyForm {...merged} />), props: merged };
}

function getButton() {
    return document.getElementById("passkey-sign-in-button") as HTMLButtonElement;
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.queryParams = {};
    vi.spyOn(console, "error").mockImplementation(() => {});
    getOptionsMock.mockResolvedValue({ options, status: 200 } as any);
    getResultMock.mockResolvedValue({ response: assertionResponse, result: AssertionResult.Success } as any);
    postResponseMock.mockResolvedValue({
        data: { data: { redirect: "https://example.com" }, status: "OK" },
        status: 200,
    } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders passkey sign in button", () => {
        renderForm();
        expect(screen.getByText("Sign in with a passkey")).toBeInTheDocument();
        expect(screen.getByText("or")).toBeInTheDocument();
        expect(screen.getByTestId("passkey-icon")).toBeInTheDocument();
    });

    it("renders button as disabled when disabled prop is true", () => {
        renderForm({ disabled: true });
        expect(getButton()).toBeDisabled();
    });
});

describe("sign in", () => {
    it("completes the passkey assertion and reports success", async () => {
        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com"));
        expect(props.onAuthenticationStart).toHaveBeenCalled();
        expect(props.onAuthenticationStop).toHaveBeenCalled();
        expect(getResultMock).toHaveBeenCalledWith(options);
    });

    it("forwards the remember me flag, redirection URL and request method", async () => {
        mocks.queryParams = { rd: "https://app.example.com", rm: "GET" };

        renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(postResponseMock).toHaveBeenCalledWith(
                assertionResponse,
                true,
                "https://app.example.com",
                "GET",
                "flow-id",
                "flow",
                "subflow",
                expect.anything(),
            ),
        );
    });

    it("reports success with no redirect when the response carries no data", async () => {
        postResponseMock.mockResolvedValue({ data: { data: null, status: "OK" }, status: 200 } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith(undefined));
    });

    it("ignores a second click while a sign in is in flight", async () => {
        let resolve: (value: unknown) => void = () => {};
        getOptionsMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderForm();

        fireEvent.click(getButton());
        await waitFor(() => expect(getOptionsMock).toHaveBeenCalledTimes(1));

        fireEvent.click(getButton());
        expect(getOptionsMock).toHaveBeenCalledTimes(1);

        resolve({ options, status: 200 });
        await waitFor(() => expect(getResultMock).toHaveBeenCalled());
    });
});

describe("failures", () => {
    it("reports a non-200 options response", async () => {
        getOptionsMock.mockResolvedValue({ options: null, status: 500 } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Failed to initiate security key sign in process" }),
            ),
        );
        expect(props.onAuthenticationStop).toHaveBeenCalled();
        expect(getResultMock).not.toHaveBeenCalled();
    });

    it("reports an empty options payload", async () => {
        getOptionsMock.mockResolvedValue({ options: null, status: 200 } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Failed to initiate security key sign in process" }),
            ),
        );
    });

    it("reports a cancelled assertion", async () => {
        getResultMock.mockResolvedValue({ result: AssertionResult.FailureUserConsent } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "You cancelled the assertion request" }),
            ),
        );
        expect(postResponseMock).not.toHaveBeenCalled();
    });

    it("reports an unrecognized device", async () => {
        getResultMock.mockResolvedValue({ result: AssertionResult.FailureUnrecognized } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "This device is not registered" }),
            ),
        );
    });

    it("reports a missing attestation response", async () => {
        getResultMock.mockResolvedValue({ response: null, result: AssertionResult.Success } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({
                    message: "The browser did not respond with the expected attestation data",
                }),
            ),
        );
        expect(postResponseMock).not.toHaveBeenCalled();
    });

    it("reports a rejected security key on a non-OK payload", async () => {
        postResponseMock.mockResolvedValue({ data: { status: "KO" }, status: 200 } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The server rejected the security key" }),
            ),
        );
        expect(props.onAuthenticationSuccess).not.toHaveBeenCalled();
    });

    it("reports a rejected security key on a non-200 status", async () => {
        postResponseMock.mockResolvedValue({ data: { status: "OK" }, status: 401 } as any);

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The server rejected the security key" }),
            ),
        );
    });

    it("reports an unexpected error", async () => {
        getOptionsMock.mockRejectedValue(new Error("boom"));

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Failed to initiate security key sign in process" }),
            ),
        );
        expect(console.error).toHaveBeenCalled();
    });

    it("stays silent when the request is cancelled", async () => {
        getOptionsMock.mockRejectedValue({ __CANCEL__: true });

        const { props } = renderForm();

        fireEvent.click(getButton());

        await waitFor(() => expect(props.onAuthenticationStop).toHaveBeenCalled());
        expect(props.onAuthenticationError).not.toHaveBeenCalled();
        expect(console.error).not.toHaveBeenCalled();
    });
});
