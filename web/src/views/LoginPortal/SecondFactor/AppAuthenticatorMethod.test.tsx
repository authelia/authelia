// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import { getAppAuthenticatorSignInStatus, initiateAppAuthenticatorSignIn } from "@services/AppAuthenticator";
import { RateLimitedError } from "@services/RateLimit";
import { AuthenticationLevel } from "@services/State";
import AppAuthenticatorMethod from "@views/LoginPortal/SecondFactor/AppAuthenticatorMethod";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => "https://target.example.com",
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: "openid_connect", id: "flow-id", subflow: undefined }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => "user-code",
}));

vi.mock("@services/AppAuthenticator", () => ({
    getAppAuthenticatorSignInStatus: vi.fn(),
    initiateAppAuthenticatorSignIn: vi.fn(),
}));

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/MethodContainer", () => ({
    default: (props: any) => (
        <div
            data-testid="method-container"
            data-state={props.state}
            data-title={props.title}
            data-registered={String(props.registered)}
        >
            <button data-testid="method-register" onClick={() => props.onRegisterClick()} />
            {props.children}
        </div>
    ),
    State: { ALREADY_AUTHENTICATED: "ALREADY_AUTHENTICATED", METHOD: "METHOD", NOT_REGISTERED: "NOT_REGISTERED" },
}));

const initiateMock = vi.mocked(initiateAppAuthenticatorSignIn);
const statusMock = vi.mocked(getAppAuthenticatorSignInStatus);

const signInBody = {
    flow: "openid_connect",
    flowID: "flow-id",
    subflow: undefined,
    targetURL: "https://target.example.com",
    userCode: "user-code",
};

function challenge(number = "07") {
    return { expires_at: Math.floor(Date.now() / 1000) + 60, nonce_length: number.length, number };
}

function renderMethod(props: Partial<{ authenticationLevel: AuthenticationLevel; registered: boolean }> = {}) {
    const merged = {
        authenticationLevel: AuthenticationLevel.OneFactor,
        id: "app-authenticator-method",
        onRegisterClick: vi.fn(),
        onSignInError: vi.fn(),
        onSignInSuccess: vi.fn(),
        registered: true,
        ...props,
    };

    return { ...render(<AppAuthenticatorMethod {...merged} />), props: merged };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    vi.spyOn(console, "error").mockImplementation(() => {});
    initiateMock.mockResolvedValue(challenge());
    statusMock.mockResolvedValue({ state: "pending" });
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders the number as a string", async () => {
        renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        expect(screen.getByText("Open the Authelia app and enter this number")).toBeInTheDocument();
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-title", "Authelia app");
    });

    it("sends the sign in body when creating the challenge", async () => {
        renderMethod();

        await vi.waitFor(() => expect(initiateMock).toHaveBeenCalledWith(signInBody, expect.any(AbortSignal)));
    });

    it("shows the remaining time", async () => {
        renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        expect(document.getElementById("app-authenticator-countdown")).toBeInTheDocument();
    });

    it("does not create a challenge when already authenticated", async () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });

        await vi.advanceTimersByTimeAsync(5000);

        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "ALREADY_AUTHENTICATED");
        expect(initiateMock).not.toHaveBeenCalled();
    });

    it("asks the user to register a device when they have none", async () => {
        const { props } = renderMethod({ registered: false });

        await vi.advanceTimersByTimeAsync(5000);

        expect(screen.getByText("Register a device in settings")).toBeInTheDocument();
        expect(initiateMock).not.toHaveBeenCalled();

        fireEvent.click(screen.getByTestId("method-register"));
        expect(props.onRegisterClick).toHaveBeenCalled();
    });
});

describe("polling", () => {
    it("polls every 2 seconds with the sign in body", async () => {
        renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        expect(statusMock).not.toHaveBeenCalled();

        await vi.advanceTimersByTimeAsync(2000);
        expect(statusMock).toHaveBeenCalledTimes(1);
        expect(statusMock).toHaveBeenCalledWith(signInBody, expect.any(AbortSignal));

        await vi.advanceTimersByTimeAsync(2000);
        expect(statusMock).toHaveBeenCalledTimes(2);
    });

    it("redirects after a pending then approved result", async () => {
        statusMock.mockResolvedValueOnce({ state: "pending" }).mockResolvedValueOnce({
            redirect: "https://target.example.com",
        });

        const { props } = renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        await vi.advanceTimersByTimeAsync(2000);
        await vi.advanceTimersByTimeAsync(2000);

        expect(screen.getByTestId("success-icon")).toBeInTheDocument();
        expect(props.onSignInSuccess).not.toHaveBeenCalled();

        await vi.advanceTimersByTimeAsync(1500);

        expect(props.onSignInSuccess).toHaveBeenCalledWith("https://target.example.com");

        await vi.advanceTimersByTimeAsync(10000);
        expect(statusMock).toHaveBeenCalledTimes(2);
    });

    it("reports success without a redirect when the response has no data", async () => {
        statusMock.mockResolvedValueOnce(undefined);

        const { props } = renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        await vi.advanceTimersByTimeAsync(3500);

        expect(props.onSignInSuccess).toHaveBeenCalledWith(undefined);
    });

    it("shows the denied message and stops polling", async () => {
        statusMock.mockResolvedValueOnce({ state: "denied" });

        renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        await vi.advanceTimersByTimeAsync(2000);

        expect(screen.getByText("Request denied on your device.")).toBeInTheDocument();
        expect(screen.queryByText("07")).not.toBeInTheDocument();
        expect(screen.queryByText("Try again")).not.toBeInTheDocument();

        await vi.advanceTimersByTimeAsync(10000);
        expect(statusMock).toHaveBeenCalledTimes(1);
    });

    it.each(["rejected", "expired"] as const)("shows try again for %s which creates a new challenge", async (state) => {
        statusMock.mockResolvedValueOnce({ state });

        renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        await vi.advanceTimersByTimeAsync(2000);

        expect(screen.getByText("The request was not approved.")).toBeInTheDocument();

        await vi.advanceTimersByTimeAsync(10000);
        expect(statusMock).toHaveBeenCalledTimes(1);

        initiateMock.mockResolvedValueOnce(challenge("42"));
        fireEvent.click(screen.getByText("Try again"));

        await vi.waitFor(() => expect(screen.getByText("42")).toBeInTheDocument());
        expect(initiateMock).toHaveBeenCalledTimes(2);

        await vi.advanceTimersByTimeAsync(2000);
        expect(statusMock).toHaveBeenCalledTimes(2);
    });

    it("stops polling on unmount", async () => {
        const { unmount } = renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());

        unmount();

        await vi.advanceTimersByTimeAsync(10000);
        expect(statusMock).not.toHaveBeenCalled();
    });
});

describe("errors", () => {
    it("reports a failure to create the challenge and offers try again", async () => {
        initiateMock.mockRejectedValueOnce(new Error("failed"));

        const { props } = renderMethod();

        await vi.waitFor(() => expect(screen.getByText("Try again")).toBeInTheDocument());
        expect(props.onSignInError).toHaveBeenCalledWith(
            expect.objectContaining({ message: "There was an issue completing sign in process" }),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("reports a failure while polling and stops", async () => {
        statusMock.mockRejectedValueOnce(new Error("failed"));

        const { props } = renderMethod();

        await vi.waitFor(() => expect(screen.getByText("07")).toBeInTheDocument());
        await vi.advanceTimersByTimeAsync(2000);

        expect(props.onSignInError).toHaveBeenCalledWith(
            expect.objectContaining({ message: "There was an issue completing sign in process" }),
        );
        expect(screen.getByText("Try again")).toBeInTheDocument();

        await vi.advanceTimersByTimeAsync(10000);
        expect(statusMock).toHaveBeenCalledTimes(1);
    });

    it("reports rate limiting", async () => {
        initiateMock.mockRejectedValueOnce(new RateLimitedError(5));

        const { props } = renderMethod();

        await vi.waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "You have made too many requests" }),
            ),
        );
        expect(document.getElementById("rate-limit-countdown")).toBeInTheDocument();
    });
});
