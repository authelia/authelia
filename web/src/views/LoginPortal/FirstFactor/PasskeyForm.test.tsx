// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { AssertionResult } from "@models/WebAuthn";
import {
    cancelWebAuthnCeremony,
    getWebAuthnAutofillSupported,
    getWebAuthnPasskeyOptions,
    getWebAuthnResult,
    postWebAuthnPasskeyResponse,
} from "@services/WebAuthn";
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
    cancelWebAuthnCeremony: vi.fn(),
    getWebAuthnAutofillSupported: vi.fn(),
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
const cancelMock = vi.mocked(cancelWebAuthnCeremony);
const autofillSupportedMock = vi.mocked(getWebAuthnAutofillSupported);

const options = { challenge: "challenge" } as any;
const assertionResponse = { id: "credential-id" };

const defaultProps = {
    conditionalMediation: false,
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
    autofillSupportedMock.mockResolvedValue(true);
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

describe("conditional mediation", () => {
    const conditionalOptions = { challenge: "conditional-challenge", timeout: 60000 } as any;

    // Simulates a pending conditional mediation request which only settles when the user selects a Passkey from the
    // autofill suggestions (resolve) or when the ceremony is cancelled.
    let pending: { resolve: (_value: unknown) => void }[];

    beforeEach(() => {
        pending = [];

        getOptionsMock.mockImplementation(async (_signal?: AbortSignal, mediation?: string) =>
            mediation === "conditional"
                ? ({ mediation: "conditional", options: conditionalOptions, status: 200 } as any)
                : ({ options, status: 200 } as any),
        );

        getResultMock.mockImplementation((_options: any, useBrowserAutofill?: boolean) => {
            if (!useBrowserAutofill) {
                return Promise.resolve({ response: assertionResponse, result: AssertionResult.Success } as any);
            }

            return new Promise((resolve) => pending.push({ resolve })) as any;
        });

        cancelMock.mockImplementation(() => {
            pending.splice(0).forEach((p) => p.resolve({ result: AssertionResult.FailureAborted }));
        });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    async function flush() {
        await act(async () => {
            await vi.advanceTimersByTimeAsync(0);
        });
    }

    it("does nothing on mount when disabled by configuration", async () => {
        renderForm();

        await act(async () => {});

        expect(autofillSupportedMock).not.toHaveBeenCalled();
        expect(getOptionsMock).not.toHaveBeenCalled();
    });

    it("does nothing when the browser does not support autofill", async () => {
        autofillSupportedMock.mockResolvedValue(false);

        renderForm({ conditionalMediation: true });

        await waitFor(() => expect(autofillSupportedMock).toHaveBeenCalled());

        expect(getOptionsMock).not.toHaveBeenCalled();
        expect(getButton()).toBeEnabled();
    });

    it("does nothing while the form is disabled and starts once it is enabled", async () => {
        const { props, rerender } = renderForm({ conditionalMediation: true, disabled: true });

        await act(async () => {});

        expect(getOptionsMock).not.toHaveBeenCalled();

        rerender(<PasskeyForm {...props} disabled={false} />);

        await waitFor(() => expect(getResultMock).toHaveBeenCalledWith(conditionalOptions, true));
    });

    it("starts a conditional request on mount without starting authentication", async () => {
        const { props } = renderForm({ conditionalMediation: true });

        await waitFor(() => expect(getResultMock).toHaveBeenCalledWith(conditionalOptions, true));

        expect(getOptionsMock).toHaveBeenCalledWith(expect.any(AbortSignal), "conditional");
        expect(props.onAuthenticationStart).not.toHaveBeenCalled();
        expect(postResponseMock).not.toHaveBeenCalled();
        expect(getButton()).toBeEnabled();
    });

    it("does not start the ceremony when the server did not honor conditional mediation", async () => {
        getOptionsMock.mockResolvedValue({ options: conditionalOptions, status: 200 } as any);

        renderForm({ conditionalMediation: true });

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());
        await act(async () => {});

        expect(getResultMock).not.toHaveBeenCalled();
    });

    it("does not start the ceremony when the options request fails", async () => {
        getOptionsMock.mockResolvedValue({ status: 403 } as any);

        const { props } = renderForm({ conditionalMediation: true });

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());
        await act(async () => {});

        expect(getResultMock).not.toHaveBeenCalled();
        expect(props.onAuthenticationError).not.toHaveBeenCalled();
    });

    it("signs in when the user selects the passkey", async () => {
        mocks.queryParams = { rd: "https://app.example.com", rm: "GET" };

        const { props } = renderForm({ conditionalMediation: true, rememberMe: true });

        await waitFor(() => expect(pending).toHaveLength(1));

        await act(async () => {
            pending[0].resolve({ response: assertionResponse, result: AssertionResult.Success });
        });

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com"));
        expect(props.onAuthenticationStart).toHaveBeenCalledTimes(1);
        expect(props.onAuthenticationStop).toHaveBeenCalledTimes(1);
        expect(postResponseMock).toHaveBeenCalledWith(
            assertionResponse,
            true,
            "https://app.example.com",
            "GET",
            "flow-id",
            "flow",
            "subflow",
            expect.any(AbortSignal),
        );
        expect(props.onAuthenticationError).not.toHaveBeenCalled();
    });

    it("reports a rejected passkey after the user selected it", async () => {
        postResponseMock.mockResolvedValue({ data: { status: "KO" }, status: 200 } as any);

        const { props } = renderForm({ conditionalMediation: true });

        await waitFor(() => expect(pending).toHaveLength(1));

        await act(async () => {
            pending[0].resolve({ response: assertionResponse, result: AssertionResult.Success });
        });

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The server rejected the security key" }),
            ),
        );
        expect(props.onAuthenticationSuccess).not.toHaveBeenCalled();
        expect(props.onAuthenticationStop).toHaveBeenCalled();
    });

    it("reports a failed post after the user selected the passkey", async () => {
        postResponseMock.mockRejectedValue(new Error("network"));

        const { props } = renderForm({ conditionalMediation: true });

        await waitFor(() => expect(pending).toHaveLength(1));

        await act(async () => {
            pending[0].resolve({ response: assertionResponse, result: AssertionResult.Success });
        });

        await waitFor(() =>
            expect(props.onAuthenticationError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Failed to initiate security key sign in process" }),
            ),
        );
        expect(props.onAuthenticationStop).toHaveBeenCalled();
    });

    it("stays silent when the conditional request is aborted", async () => {
        const { props } = renderForm({ conditionalMediation: true });

        await waitFor(() => expect(pending).toHaveLength(1));

        await act(async () => {
            pending[0].resolve({ result: AssertionResult.FailureAborted });
        });

        expect(props.onAuthenticationError).not.toHaveBeenCalled();
        expect(props.onAuthenticationStart).not.toHaveBeenCalled();
        expect(postResponseMock).not.toHaveBeenCalled();
    });

    it("quietly starts a new conditional request when the user declines", async () => {
        vi.useFakeTimers();

        const { props } = renderForm({ conditionalMediation: true });

        await flush();
        expect(pending).toHaveLength(1);
        expect(getOptionsMock).toHaveBeenCalledTimes(1);

        await act(async () => {
            pending.shift()?.resolve({ result: AssertionResult.FailureUserConsent });
        });

        await act(async () => {
            await vi.advanceTimersByTimeAsync(999);
        });
        expect(getOptionsMock).toHaveBeenCalledTimes(1);

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1);
        });
        expect(getOptionsMock).toHaveBeenCalledTimes(2);
        expect(pending).toHaveLength(1);
        expect(props.onAuthenticationError).not.toHaveBeenCalled();
    });

    it("does not start a new conditional request after an unexpected failure", async () => {
        vi.useFakeTimers();

        const { props } = renderForm({ conditionalMediation: true });

        await flush();

        await act(async () => {
            pending.shift()?.resolve({ result: AssertionResult.FailureUnknown });
        });

        await act(async () => {
            await vi.advanceTimersByTimeAsync(120000);
        });

        expect(getOptionsMock).toHaveBeenCalledTimes(1);
        expect(props.onAuthenticationError).not.toHaveBeenCalled();
    });

    it("refreshes the challenge before it expires on the server", async () => {
        vi.useFakeTimers();

        renderForm({ conditionalMediation: true });

        await flush();
        expect(getOptionsMock).toHaveBeenCalledTimes(1);

        await act(async () => {
            await vi.advanceTimersByTimeAsync(54999);
        });
        expect(cancelMock).not.toHaveBeenCalled();
        expect(getOptionsMock).toHaveBeenCalledTimes(1);

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1);
        });
        expect(cancelMock).toHaveBeenCalledTimes(1);
        expect(getOptionsMock).toHaveBeenCalledTimes(2);
        expect(getResultMock).toHaveBeenCalledTimes(2);
        expect(pending).toHaveLength(1);

        await act(async () => {
            await vi.advanceTimersByTimeAsync(55000);
        });
        expect(getOptionsMock).toHaveBeenCalledTimes(3);
    });

    it("uses a minimum refresh interval for short timeouts", async () => {
        vi.useFakeTimers();

        getOptionsMock.mockResolvedValue({
            mediation: "conditional",
            options: { ...conditionalOptions, timeout: 1000 },
            status: 200,
        } as any);

        renderForm({ conditionalMediation: true });

        await flush();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(14999);
        });
        expect(getOptionsMock).toHaveBeenCalledTimes(1);

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1);
        });
        expect(getOptionsMock).toHaveBeenCalledTimes(2);
    });

    it("never refreshes once the user selected the passkey", async () => {
        vi.useFakeTimers();

        let resolvePost: (_value: unknown) => void = () => {};
        postResponseMock.mockReturnValue(new Promise((r) => (resolvePost = r)) as any);

        const { props } = renderForm({ conditionalMediation: true });

        await flush();

        await act(async () => {
            pending.shift()?.resolve({ response: assertionResponse, result: AssertionResult.Success });
        });

        await act(async () => {
            await vi.advanceTimersByTimeAsync(120000);
        });

        expect(getOptionsMock).toHaveBeenCalledTimes(1);
        expect(cancelMock).not.toHaveBeenCalled();

        await act(async () => {
            resolvePost({ data: { data: { redirect: "https://example.com" }, status: "OK" }, status: 200 });
        });

        expect(props.onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com");
    });

    it("cancels the conditional request and timers on unmount", async () => {
        vi.useFakeTimers();

        const { unmount } = renderForm({ conditionalMediation: true });

        await flush();
        expect(pending).toHaveLength(1);

        unmount();

        expect(cancelMock).toHaveBeenCalledTimes(1);

        await act(async () => {
            await vi.advanceTimersByTimeAsync(120000);
        });

        expect(getOptionsMock).toHaveBeenCalledTimes(1);
    });

    it("cancels the conditional request when the form becomes disabled", async () => {
        const { props, rerender } = renderForm({ conditionalMediation: true });

        await waitFor(() => expect(pending).toHaveLength(1));

        rerender(<PasskeyForm {...props} disabled={true} />);

        expect(cancelMock).toHaveBeenCalledTimes(1);
        expect(pending).toHaveLength(0);
        expect(props.onAuthenticationError).not.toHaveBeenCalled();
    });

    it("runs the modal flow when the button is clicked while a conditional request is pending", async () => {
        const { props } = renderForm({ conditionalMediation: true });

        await waitFor(() => expect(pending).toHaveLength(1));

        fireEvent.click(getButton());

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com"));
        expect(cancelMock).toHaveBeenCalledTimes(1);
        expect(getOptionsMock).toHaveBeenLastCalledWith(expect.any(AbortSignal));
        expect(getResultMock).toHaveBeenLastCalledWith(options);
        expect(props.onAuthenticationSuccess).toHaveBeenCalledTimes(1);
        expect(props.onAuthenticationStart).toHaveBeenCalledTimes(1);
        expect(props.onAuthenticationError).not.toHaveBeenCalled();
        expect(postResponseMock).toHaveBeenCalledTimes(1);
    });
});
