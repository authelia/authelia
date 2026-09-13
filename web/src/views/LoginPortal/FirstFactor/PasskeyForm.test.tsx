// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { AssertionResult } from "@models/WebAuthn";
import { getWebAuthnPasskeyOptions, getWebAuthnResult, postWebAuthnPasskeyResponse } from "@services/WebAuthn";
import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";

const mocks = vi.hoisted(() => ({
    autofillSupported: false,
    queryParams: {} as Record<string, null | string>,
}));

vi.mock("@simplewebauthn/browser", () => ({
    browserSupportsWebAuthnAutofill: () => Promise.resolve(mocks.autofillSupported),
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

function getRememberMeDialog() {
    return document.getElementById("remember-me-dialog");
}

async function answerRememberMe(rememberMe: boolean) {
    const id = rememberMe ? "dialog-remember-me-yes" : "dialog-remember-me-no";

    await waitFor(() => expect(document.getElementById(id)).toBeInTheDocument());

    fireEvent.click(document.getElementById(id) as HTMLButtonElement);
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.autofillSupported = false;
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
        expect(getResultMock).toHaveBeenCalledWith(options, false);
        expect(getOptionsMock).toHaveBeenCalledWith(expect.anything(), false);
    });

    it("forwards the remember me flag, redirection URL and request method", async () => {
        mocks.queryParams = { rd: "https://app.example.com", rm: "GET" };

        renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await answerRememberMe(true);

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

describe("conditional mediation", () => {
    it("does not start a ceremony when the browser does not support autofill", async () => {
        renderForm();

        await waitFor(() => expect(getButton()).toBeInTheDocument());

        expect(getOptionsMock).not.toHaveBeenCalled();
    });

    it("starts a conditional ceremony on mount when autofill is supported", async () => {
        mocks.autofillSupported = true;

        renderForm();

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalledWith(expect.anything(), true));
        expect(getResultMock).toHaveBeenCalledWith(options, true);
    });

    it("stays idle until the user picks a credential", async () => {
        mocks.autofillSupported = true;

        let resolve: (value: unknown) => void = () => {};
        getResultMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        const { props } = renderForm();

        await waitFor(() => expect(getResultMock).toHaveBeenCalled());

        expect(props.onAuthenticationStart).not.toHaveBeenCalled();

        resolve({ response: assertionResponse, result: AssertionResult.Success });

        await waitFor(() => expect(props.onAuthenticationStart).toHaveBeenCalled());
    });

    it("completes the sign in when the user picks a credential from autofill", async () => {
        mocks.autofillSupported = true;

        const { props } = renderForm();

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com"));
        expect(props.onAuthenticationStart).toHaveBeenCalled();
        expect(props.onAuthenticationStop).toHaveBeenCalled();
    });

    it("prompts for remember me once a credential is picked from autofill", async () => {
        mocks.autofillSupported = true;

        renderForm({ rememberMe: true });

        await answerRememberMe(true);

        await waitFor(() => expect(postResponseMock).toHaveBeenCalled());
        expect(postResponseMock.mock.calls[0][1]).toBe(true);
    });

    it("stays silent when the ceremony is superseded or dismissed", async () => {
        mocks.autofillSupported = true;
        getResultMock.mockResolvedValue({ result: AssertionResult.FailureUserConsent } as any);

        const { props } = renderForm();

        await waitFor(() => expect(getResultMock).toHaveBeenCalled());

        expect(props.onAuthenticationError).not.toHaveBeenCalled();
        expect(props.onAuthenticationStart).not.toHaveBeenCalled();
        expect(postResponseMock).not.toHaveBeenCalled();
    });

    it("stays silent when the options request fails", async () => {
        mocks.autofillSupported = true;
        getOptionsMock.mockResolvedValue({ options: null, status: 500 } as any);

        const { props } = renderForm();

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());

        expect(props.onAuthenticationError).not.toHaveBeenCalled();
        expect(getResultMock).not.toHaveBeenCalled();
    });

    it("stays silent when the options request throws", async () => {
        mocks.autofillSupported = true;
        getOptionsMock.mockRejectedValue(new Error("boom"));

        const { props } = renderForm();

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());

        expect(props.onAuthenticationError).not.toHaveBeenCalled();
    });

    it("reports a server rejection once the user has picked a credential", async () => {
        mocks.autofillSupported = true;
        postResponseMock.mockResolvedValue({ data: { status: "KO" }, status: 200 } as any);

        const { props } = renderForm();

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The server rejected the security key" }),
            ),
        );
    });

    it("does not clear the busy state of an explicit ceremony that superseded it", async () => {
        mocks.autofillSupported = true;

        let rejectConditional: (reason?: unknown) => void = () => {};
        let resolveExplicit: (value: unknown) => void = () => {};

        getResultMock
            .mockReturnValueOnce(new Promise((_, reject) => (rejectConditional = reject)) as any)
            .mockReturnValueOnce(new Promise((r) => (resolveExplicit = r)) as any);

        const { props } = renderForm();

        await waitFor(() => expect(getResultMock).toHaveBeenCalledTimes(1));

        fireEvent.click(getButton());

        await waitFor(() => expect(props.onAuthenticationStart).toHaveBeenCalled());

        // The browser aborts the conditional ceremony as soon as the explicit one starts.
        rejectConditional(new DOMException("aborted", "AbortError"));

        await waitFor(() => expect(getResultMock).toHaveBeenCalledTimes(2));

        expect(props.onAuthenticationStop).not.toHaveBeenCalled();

        resolveExplicit({ response: assertionResponse, result: AssertionResult.Success });

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalled());
    });

    it("does not start a ceremony when the component unmounts first", async () => {
        mocks.autofillSupported = true;

        const { unmount } = renderForm();

        unmount();

        await waitFor(() => expect(getOptionsMock).not.toHaveBeenCalled());
    });
});

describe("remember me", () => {
    it("does not prompt and never remembers when the feature is disabled", async () => {
        const { props } = renderForm({ rememberMe: false });

        fireEvent.click(getButton());

        await waitFor(() => expect(postResponseMock).toHaveBeenCalled());

        expect(getRememberMeDialog()).not.toBeInTheDocument();
        expect(postResponseMock.mock.calls[0][1]).toBe(false);
        expect(props.onAuthenticationSuccess).toHaveBeenCalled();
    });

    it("prompts only after the assertion has produced a credential", async () => {
        let resolve: (value: unknown) => void = () => {};
        getResultMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await waitFor(() => expect(getResultMock).toHaveBeenCalled());

        expect(getRememberMeDialog()).not.toBeInTheDocument();

        resolve({ response: assertionResponse, result: AssertionResult.Success });

        await waitFor(() => expect(getRememberMeDialog()).toBeInTheDocument());
        expect(postResponseMock).not.toHaveBeenCalled();
    });

    it("remembers the session when the user answers yes", async () => {
        renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await answerRememberMe(true);

        await waitFor(() => expect(postResponseMock).toHaveBeenCalled());
        expect(postResponseMock.mock.calls[0][1]).toBe(true);
    });

    it("does not remember the session when the user answers no", async () => {
        renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await answerRememberMe(false);

        await waitFor(() => expect(postResponseMock).toHaveBeenCalled());
        expect(postResponseMock.mock.calls[0][1]).toBe(false);
    });

    it("does not remember the session when the prompt is dismissed", async () => {
        renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await waitFor(() => expect(getRememberMeDialog()).toBeInTheDocument());

        fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });

        await waitFor(() => expect(postResponseMock).toHaveBeenCalled());
        expect(postResponseMock.mock.calls[0][1]).toBe(false);
    });

    it("closes the prompt once it has been answered", async () => {
        renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await answerRememberMe(true);

        await waitFor(() => expect(getRememberMeDialog()).not.toBeInTheDocument());
    });

    it("does not prompt when the ceremony fails before producing a credential", async () => {
        getResultMock.mockResolvedValue({ result: AssertionResult.FailureUserConsent } as any);

        const { props } = renderForm({ rememberMe: true });

        fireEvent.click(getButton());

        await waitFor(() => expect(props.onAuthenticationError).toHaveBeenCalled());

        expect(getRememberMeDialog()).not.toBeInTheDocument();
        expect(postResponseMock).not.toHaveBeenCalled();
    });
});
