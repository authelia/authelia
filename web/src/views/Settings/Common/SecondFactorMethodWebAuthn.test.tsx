import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { AssertionResult, WebAuthnTouchState } from "@models/WebAuthn";
import { getWebAuthnOptions, getWebAuthnResult, postWebAuthnResponse } from "@services/WebAuthn";
import SecondFactorMethodWebAuthn from "@views/Settings/Common/SecondFactorMethodWebAuthn";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
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

const getOptionsMock = vi.mocked(getWebAuthnOptions);
const getResultMock = vi.mocked(getWebAuthnResult);
const postResponseMock = vi.mocked(postWebAuthnResponse);

const options = { challenge: "challenge" } as any;
const assertionResponse = { id: "credential-id" };

function renderMethod(onSecondFactorSuccess = vi.fn()) {
    return {
        ...render(<SecondFactorMethodWebAuthn onSecondFactorSuccess={onSecondFactorSuccess} />),
        onSecondFactorSuccess,
    };
}

function expectState(state: WebAuthnTouchState) {
    return waitFor(() => expect(screen.getByTestId("webauthn-try-icon")).toHaveAttribute("data-state", String(state)));
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    getOptionsMock.mockResolvedValue({ options, status: 200 } as any);
    getResultMock.mockResolvedValue({ response: assertionResponse, result: AssertionResult.Success } as any);
    postResponseMock.mockResolvedValue({ data: { status: "OK" }, status: 200 } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders WebAuthnTryIcon", async () => {
        renderMethod();
        expect(screen.getByTestId("webauthn-try-icon")).toBeInTheDocument();
        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());
    });

    it("starts the assertion automatically", async () => {
        renderMethod();
        await waitFor(() => expect(getOptionsMock).toHaveBeenCalledTimes(1));
    });
});

describe("assertion", () => {
    it("completes the assertion and reports success", async () => {
        const { onSecondFactorSuccess } = renderMethod();

        await waitFor(() => expect(onSecondFactorSuccess).toHaveBeenCalled());
        expect(getResultMock).toHaveBeenCalledWith(options);
        expect(postResponseMock).toHaveBeenCalledWith(
            assertionResponse,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            expect.anything(),
        );
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

        const { onSecondFactorSuccess } = renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(console.error).toHaveBeenCalledWith(
            expect.objectContaining({ message: "Failed to initiate security key sign in process" }),
        );
        expect(onSecondFactorSuccess).not.toHaveBeenCalled();
    });

    it("fails on an empty options payload", async () => {
        getOptionsMock.mockResolvedValue({ options: null, status: 200 } as any);

        renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(getResultMock).not.toHaveBeenCalled();
    });

    it("fails when the user cancels the assertion", async () => {
        getResultMock.mockResolvedValue({ result: AssertionResult.FailureUserConsent } as any);

        renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(console.error).toHaveBeenCalledWith(
            expect.objectContaining({ message: "You cancelled the assertion request" }),
        );
    });

    it("fails when the browser returns no attestation data", async () => {
        getResultMock.mockResolvedValue({ response: null, result: AssertionResult.Success } as any);

        renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(console.error).toHaveBeenCalledWith(
            expect.objectContaining({ message: "The browser did not respond with the expected attestation data" }),
        );
        expect(postResponseMock).not.toHaveBeenCalled();
    });

    it("fails when the server rejects the key", async () => {
        postResponseMock.mockResolvedValue({ data: { status: "KO" }, status: 200 } as any);

        const { onSecondFactorSuccess } = renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(console.error).toHaveBeenCalledWith(
            expect.objectContaining({ message: "The server rejected the security key" }),
        );
        expect(onSecondFactorSuccess).not.toHaveBeenCalled();
    });

    it("fails on a non-200 response status", async () => {
        postResponseMock.mockResolvedValue({ data: { status: "OK" }, status: 401 } as any);

        renderMethod();

        await expectState(WebAuthnTouchState.Failure);
    });

    it("fails on an unexpected error", async () => {
        getOptionsMock.mockRejectedValue(new Error("boom"));

        renderMethod();

        await expectState(WebAuthnTouchState.Failure);
        expect(console.error).toHaveBeenCalled();
    });

    it("stays silent when the request is cancelled", async () => {
        getOptionsMock.mockRejectedValue({ __CANCEL__: true });

        renderMethod();

        await waitFor(() => expect(getOptionsMock).toHaveBeenCalled());
        expect(console.error).not.toHaveBeenCalled();
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
