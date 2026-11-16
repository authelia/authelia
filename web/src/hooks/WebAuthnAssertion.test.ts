// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from "@testing-library/react";

import { useWebAuthnAssertion } from "@hooks/WebAuthnAssertion";
import { AssertionResult, WebAuthnTouchState } from "@models/WebAuthn";
import { cancelWebAuthnCeremony, getWebAuthnResult, isWebAuthnSupported } from "@services/WebAuthn";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@services/WebAuthn", () => ({
    cancelWebAuthnCeremony: vi.fn(),
    getWebAuthnResult: vi.fn(),
    isWebAuthnSupported: vi.fn(),
}));

const getResultMock = vi.mocked(getWebAuthnResult);
const isSupportedMock = vi.mocked(isWebAuthnSupported);
const cancelCeremonyMock = vi.mocked(cancelWebAuthnCeremony);

const options = { challenge: "challenge", timeout: 60000 } as any;
const assertionResponse = { id: "credential-id" } as any;

function setup() {
    const callbacks = {
        getOptions: vi.fn().mockResolvedValue({ options, status: 200 }),
        onError: vi.fn(),
        onSuccess: vi.fn(),
        postResponse: vi.fn().mockResolvedValue({
            data: { data: { redirect: "https://example.com" }, status: "OK" },
            status: 200,
        }),
    };

    const hook = renderHook(() => useWebAuthnAssertion(callbacks));

    return { callbacks, hook };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    isSupportedMock.mockReturnValue(true);
    getResultMock.mockResolvedValue({ response: assertionResponse, result: AssertionResult.Success } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("starts in the waiting state without a ceremony", () => {
    const { hook } = setup();

    expect(hook.result.current.state).toBe(WebAuthnTouchState.WaitTouch);
    expect(hook.result.current.ceremony).toBeUndefined();
});

it("completes the assertion and reports the redirect", async () => {
    const { callbacks, hook } = setup();

    await act(() => hook.result.current.start());

    expect(callbacks.getOptions).toHaveBeenCalledWith(expect.any(AbortSignal));
    expect(getResultMock).toHaveBeenCalledWith(options);
    expect(callbacks.postResponse).toHaveBeenCalledWith(assertionResponse, expect.any(AbortSignal));
    expect(callbacks.onSuccess).toHaveBeenCalledWith("https://example.com");
    expect(callbacks.onError).not.toHaveBeenCalled();
    expect(hook.result.current.state).toBe(WebAuthnTouchState.InProgress);
    expect(hook.result.current.ceremony).toEqual({ timeout: 60000 });
});

it("reports success without a redirect when the response carries no data", async () => {
    const { callbacks, hook } = setup();
    callbacks.postResponse.mockResolvedValue({ data: { status: "OK" }, status: 200 });

    await act(() => hook.result.current.start());

    expect(callbacks.onSuccess).toHaveBeenCalledWith(undefined);
});

it("reports an unsupported browser without requesting a challenge", async () => {
    const { callbacks, hook } = setup();
    isSupportedMock.mockReturnValue(false);

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.Unsupported);
    expect(callbacks.onError).toHaveBeenCalledWith(new Error("Your browser does not support the WebAuthn protocol"));
    expect(callbacks.getOptions).not.toHaveBeenCalled();
});

it("fails when the challenge cannot be retrieved", async () => {
    const { callbacks, hook } = setup();
    callbacks.getOptions.mockResolvedValue({ status: 403 });

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.Failure);
    expect(callbacks.onError).toHaveBeenCalledWith(
        new Error("Unable to start sign in with your passkey or security key"),
    );
    expect(getResultMock).not.toHaveBeenCalled();
});

it("fails with the assertion failure message", async () => {
    const { callbacks, hook } = setup();
    getResultMock.mockResolvedValue({ result: AssertionResult.FailureUserConsent } as any);

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.Failure);
    expect(callbacks.onError).toHaveBeenCalledWith(new Error("The request timed out or was cancelled"));
    expect(callbacks.postResponse).not.toHaveBeenCalled();
});

it("fails silently when the ceremony is aborted", async () => {
    const { callbacks, hook } = setup();
    getResultMock.mockResolvedValue({ result: AssertionResult.FailureAborted } as any);

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.Failure);
    expect(callbacks.onError).not.toHaveBeenCalled();
});

it("fails when the browser returns no assertion response", async () => {
    const { callbacks, hook } = setup();
    getResultMock.mockResolvedValue({ result: AssertionResult.Success } as any);

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.Failure);
    expect(callbacks.onError).toHaveBeenCalledWith(
        new Error("Your browser did not return a response from your passkey or security key"),
    );
    expect(callbacks.postResponse).not.toHaveBeenCalled();
});

it("fails when the server rejects the assertion", async () => {
    const { callbacks, hook } = setup();
    callbacks.postResponse.mockResolvedValue({ data: { status: "KO" }, status: 403 });

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.Failure);
    expect(callbacks.onError).toHaveBeenCalledWith(new Error("Your passkey or security key could not be verified"));
    expect(callbacks.onSuccess).not.toHaveBeenCalled();
});

it("fails on an unexpected error", async () => {
    const { callbacks, hook } = setup();
    callbacks.getOptions.mockRejectedValue(new Error("boom"));

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.Failure);
    expect(callbacks.onError).toHaveBeenCalledWith(
        new Error("Unable to start sign in with your passkey or security key"),
    );
    expect(console.error).toHaveBeenCalled();
});

it("stays silent when the request is cancelled", async () => {
    const { callbacks, hook } = setup();
    callbacks.getOptions.mockRejectedValue({ __CANCEL__: true });

    await act(() => hook.result.current.start());

    expect(hook.result.current.state).toBe(WebAuthnTouchState.WaitTouch);
    expect(callbacks.onError).not.toHaveBeenCalled();
});

it("ignores a start while a ceremony is in flight", async () => {
    const { callbacks, hook } = setup();
    let resolve: (value: unknown) => void = () => {};
    getResultMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

    let first: Promise<void> = Promise.resolve();

    await act(async () => {
        first = hook.result.current.start();
        await Promise.resolve();
    });

    await act(() => hook.result.current.start());

    expect(callbacks.getOptions).toHaveBeenCalledTimes(1);

    await act(async () => {
        resolve({ response: assertionResponse, result: AssertionResult.Success });
        await first;
    });

    expect(callbacks.onSuccess).toHaveBeenCalledTimes(1);
});

it("allows a new start after the previous one completes", async () => {
    const { callbacks, hook } = setup();
    callbacks.getOptions.mockResolvedValueOnce({ status: 403 });

    await act(() => hook.result.current.start());
    await act(() => hook.result.current.start());

    expect(callbacks.getOptions).toHaveBeenCalledTimes(2);
    expect(callbacks.onSuccess).toHaveBeenCalledTimes(1);
});

it("keeps start stable across renders", () => {
    const { hook } = setup();
    const start = hook.result.current.start;

    hook.rerender();

    expect(hook.result.current.start).toBe(start);
});

it("cancels the ceremony on unmount", () => {
    const { hook } = setup();

    expect(cancelCeremonyMock).not.toHaveBeenCalled();

    hook.unmount();

    expect(cancelCeremonyMock).toHaveBeenCalledTimes(1);
});
