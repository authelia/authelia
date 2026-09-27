// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { AssertionResult, WebAuthnTouchState } from "@models/WebAuthn";
import { AuthenticationLevel } from "@services/State";
import { getWebAuthnOptions, getWebAuthnResult, postWebAuthnResponse } from "@services/WebAuthn";
import WebAuthnMethod from "@views/LoginPortal/SecondFactor/WebAuthnMethod";

const mocks = vi.hoisted(() => ({
    queryParams: {} as Record<string, null | string>,
    userCode: null as null | string,
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

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => mocks.userCode,
}));

vi.mock("@services/WebAuthn", () => ({
    getWebAuthnOptions: vi.fn(),
    getWebAuthnResult: vi.fn(),
    postWebAuthnResponse: vi.fn(),
}));

vi.mock("@components/WebAuthnTryIcon", () => ({
    default: (props: any) => (
        <div data-testid="webauthn-try-icon" data-state={props.webauthnTouchState}>
            <button data-testid="retry" onClick={() => props.onRetryClick()} />
        </div>
    ),
}));

vi.mock("@views/LoginPortal/SecondFactor/MethodContainer", () => ({
    default: (props: any) => (
        <div data-testid="method-container" data-state={props.state} data-title={props.title}>
            <button data-testid="register" onClick={() => props.onRegisterClick()} />
            {props.children}
        </div>
    ),
    State: { ALREADY_AUTHENTICATED: "ALREADY_AUTHENTICATED", METHOD: "METHOD", NOT_REGISTERED: "NOT_REGISTERED" },
}));

const getOptionsMock = vi.mocked(getWebAuthnOptions);
const getResultMock = vi.mocked(getWebAuthnResult);
const postResponseMock = vi.mocked(postWebAuthnResponse);

const options = { challenge: "challenge" } as any;
const assertionResponse = { id: "credential-id" };

const defaultProps = {
    authenticationLevel: AuthenticationLevel.OneFactor,
    id: "webauthn-method",
    onRegisterClick: vi.fn(),
    onSignInError: vi.fn(),
    onSignInSuccess: vi.fn(),
    registered: true,
};

function renderMethod(props: Partial<typeof defaultProps> = {}) {
    const merged = {
        ...defaultProps,
        onRegisterClick: vi.fn(),
        onSignInError: vi.fn(),
        onSignInSuccess: vi.fn(),
        ...props,
    };

    return { ...render(<WebAuthnMethod {...merged} />), props: merged };
}

function expectState(state: WebAuthnTouchState) {
    return waitFor(() => expect(screen.getByTestId("webauthn-try-icon")).toHaveAttribute("data-state", String(state)));
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.queryParams = {};
    mocks.userCode = null;
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
    it("renders method container with security key title", async () => {
        renderMethod();
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-title", "Security Key");
        expect(screen.getByTestId("webauthn-try-icon")).toBeInTheDocument();
        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());
    });

    it("renders not registered state when not registered", () => {
        renderMethod({ registered: false });
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "NOT_REGISTERED");
    });

    it("renders already authenticated state at two factor level", () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "ALREADY_AUTHENTICATED");
    });

    it("forwards the register request", () => {
        const { props } = renderMethod({ registered: false });

        fireEvent.click(screen.getByTestId("register"));

        expect(props.onRegisterClick).toHaveBeenCalled();
    });
});

describe("initiation guards", () => {
    it("does not sign in when the user is not registered", async () => {
        renderMethod({ registered: false });
        await waitFor(() => expect(getOptionsMock).not.toHaveBeenCalled());
    });

    it("does not sign in when already at two factor level", async () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });
        await waitFor(() => expect(getOptionsMock).not.toHaveBeenCalled());
    });
});

describe("assertion", () => {
    it("completes the assertion and reports success", async () => {
        const { props } = renderMethod();

        await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalledWith("https://example.com"));
        expect(getResultMock).toHaveBeenCalledWith(options);
    });

    it("forwards the redirection URL, flow details and user code", async () => {
        mocks.queryParams = { rd: "https://app.example.com" };
        mocks.userCode = "ABCD-EFGH";

        renderMethod();

        await waitFor(() =>
            expect(postResponseMock).toHaveBeenCalledWith(
                assertionResponse,
                "https://app.example.com",
                "flow-id",
                "flow",
                "subflow",
                "ABCD-EFGH",
                expect.anything(),
            ),
        );
    });

    it("reports success with no redirect when the response carries no data", async () => {
        postResponseMock.mockResolvedValue({ data: { data: null, status: "OK" }, status: 200 } as any);

        const { props } = renderMethod();

        await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalledWith(undefined));
    });

    it("reports in progress while posting the response", async () => {
        let resolve: (value: unknown) => void = () => {};
        postResponseMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderMethod();

        await expectState(WebAuthnTouchState.InProgress);

        resolve({ data: { status: "OK" }, status: 200 });
        await waitFor(() => expect(postResponseMock).toHaveBeenCalled());
    });

    it("fails on a non-200 options response", async () => {
        getOptionsMock.mockResolvedValue({ options: null, status: 400 } as any);

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Failed to initiate security key sign in process" }),
            ),
        );
        await expectState(WebAuthnTouchState.Failure);
    });

    it("fails on an empty options payload", async () => {
        getOptionsMock.mockResolvedValue({ options: null, status: 200 } as any);

        renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(getResultMock).not.toHaveBeenCalled();
    });

    it("fails when the user cancels the assertion", async () => {
        getResultMock.mockResolvedValue({ result: AssertionResult.FailureUserConsent } as any);

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "You cancelled the assertion request" }),
            ),
        );
        await expectState(WebAuthnTouchState.Failure);
    });

    it("fails when the browser returns no attestation data", async () => {
        getResultMock.mockResolvedValue({ response: null, result: AssertionResult.Success } as any);

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({
                    message: "The browser did not respond with the expected attestation data",
                }),
            ),
        );
        expect(postResponseMock).not.toHaveBeenCalled();
    });

    it("fails when the server rejects the key", async () => {
        postResponseMock.mockResolvedValue({ data: { status: "KO" }, status: 200 } as any);

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The server rejected the security key" }),
            ),
        );
        expect(props.onSignInSuccess).not.toHaveBeenCalled();
    });

    it("fails on a non-200 response status", async () => {
        postResponseMock.mockResolvedValue({ data: { status: "OK" }, status: 401 } as any);

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The server rejected the security key" }),
            ),
        );
    });

    it("fails on an unexpected error", async () => {
        getOptionsMock.mockRejectedValue(new Error("boom"));

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Failed to initiate security key sign in process" }),
            ),
        );
        expect(console.error).toHaveBeenCalled();
    });

    it("stays silent when the request is cancelled", async () => {
        getOptionsMock.mockRejectedValue({ __CANCEL__: true });

        const { props } = renderMethod();

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());
        expect(props.onSignInError).not.toHaveBeenCalled();
        expect(screen.getByTestId("webauthn-try-icon")).toHaveAttribute(
            "data-state",
            String(WebAuthnTouchState.WaitTouch),
        );
    });
});

describe("retry", () => {
    it("restarts the assertion", async () => {
        getOptionsMock.mockResolvedValue({ options: null, status: 400 } as any);

        renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(getOptionsMock).toHaveBeenCalledTimes(1);

        fireEvent.click(screen.getByTestId("retry"));

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalledTimes(2));
    });
});
