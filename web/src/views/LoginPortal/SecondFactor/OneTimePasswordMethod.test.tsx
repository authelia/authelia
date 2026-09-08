import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { completeTOTPSignIn } from "@services/OneTimePassword";
import { AuthenticationLevel } from "@services/State";
import OneTimePasswordMethod from "@views/LoginPortal/SecondFactor/OneTimePasswordMethod";

const mocks = vi.hoisted(() => ({
    config: { digits: 6, period: 30 } as any,
    configError: undefined as Error | undefined,
    fetchConfig: vi.fn(),
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

vi.mock("@hooks/UserInfoTOTPConfiguration", () => ({
    useUserInfoTOTPConfiguration: () => [mocks.config, mocks.fetchConfig, false, mocks.configError],
}));

vi.mock("@services/OneTimePassword", () => ({
    completeTOTPSignIn: vi.fn(),
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/OTPDial", () => ({
    default: (props: any) => (
        <div
            data-testid="otp-dial"
            data-digits={props.digits}
            data-period={props.period}
            data-state={props.state}
            data-passcode={props.passcode}
        >
            <button data-testid="dial-full" onClick={() => props.onChange("123456")} />
            <button data-testid="dial-full-8" onClick={() => props.onChange("12345678")} />
            <button data-testid="dial-partial" onClick={() => props.onChange("123")} />
        </div>
    ),
    State: { Failure: 3, Idle: 0, InProgress: 1, RateLimited: 4, Success: 2 },
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

const completeSignInMock = vi.mocked(completeTOTPSignIn);

const defaultProps = {
    authenticationLevel: AuthenticationLevel.OneFactor,
    id: "otp-method",
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

    return { ...render(<OneTimePasswordMethod {...merged} />), props: merged };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.config = { digits: 6, period: 30 };
    mocks.configError = undefined;
    mocks.queryParams = {};
    mocks.userCode = null;
    completeSignInMock.mockResolvedValue({ data: { redirect: "https://example.com" } } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders method container with OTP dial when registered", () => {
        renderMethod();
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-title", "One-Time Password");
        expect(screen.getByTestId("otp-dial")).toBeInTheDocument();
        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-digits", "6");
        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-period", "30");
    });

    it("renders not registered state when not registered", () => {
        renderMethod({ registered: false });
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "NOT_REGISTERED");
    });

    it("renders already authenticated state at two factor level", () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "ALREADY_AUTHENTICATED");
        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "2");
    });

    it("renders a loading page while the configuration is unavailable", () => {
        mocks.config = undefined;

        renderMethod();

        expect(screen.getByTestId("loading-page")).toBeInTheDocument();
        expect(screen.queryByTestId("otp-dial")).not.toBeInTheDocument();
    });

    it("falls back to the default digits and period", () => {
        mocks.config = {};

        renderMethod();

        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-digits", "6");
        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-period", "30");
    });

    it("forwards the register request", () => {
        const { props } = renderMethod({ registered: false });

        fireEvent.click(screen.getByTestId("register"));

        expect(props.onRegisterClick).toHaveBeenCalled();
    });
});

describe("configuration", () => {
    it("fetches the configuration when registered at one factor level", () => {
        renderMethod();
        expect(mocks.fetchConfig).toHaveBeenCalled();
    });

    it("does not fetch the configuration when not registered", () => {
        renderMethod({ registered: false });
        expect(mocks.fetchConfig).not.toHaveBeenCalled();
    });

    it("does not fetch the configuration when already at two factor level", () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });
        expect(mocks.fetchConfig).not.toHaveBeenCalled();
    });

    it("reports a failure to load the configuration", async () => {
        mocks.configError = new Error("boom");

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Could not obtain user settings" }),
            ),
        );
        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "3"));
    });

    it("shows the dial when only an error is available", () => {
        mocks.config = undefined;
        mocks.configError = new Error("boom");

        renderMethod();

        expect(screen.getByTestId("otp-dial")).toBeInTheDocument();
    });
});

describe("sign in", () => {
    it("submits once the full passcode is entered", async () => {
        const { props } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() =>
            expect(completeSignInMock).toHaveBeenCalledWith("123456", null, "flow-id", "flow", "subflow", null),
        );
        expect(props.onSignInSuccess).toHaveBeenCalledWith("https://example.com");
    });

    it("forwards the redirection URL and user code", async () => {
        mocks.queryParams = { rd: "https://app.example.com" };
        mocks.userCode = "ABCD-EFGH";

        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() =>
            expect(completeSignInMock).toHaveBeenCalledWith(
                "123456",
                "https://app.example.com",
                "flow-id",
                "flow",
                "subflow",
                "ABCD-EFGH",
            ),
        );
    });

    it("does not submit a partial passcode", async () => {
        renderMethod();

        fireEvent.click(screen.getByTestId("dial-partial"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123"));
        expect(completeSignInMock).not.toHaveBeenCalled();
    });

    it("waits for the configured number of digits", async () => {
        mocks.config = { digits: 8, period: 30 };

        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));
        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123456"));
        expect(completeSignInMock).not.toHaveBeenCalled();

        fireEvent.click(screen.getByTestId("dial-full-8"));
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());
    });

    it("does not submit when not registered", async () => {
        renderMethod({ registered: false });

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123456"));
        expect(completeSignInMock).not.toHaveBeenCalled();
    });

    it("does not submit when already at two factor level", async () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123456"));
        expect(completeSignInMock).not.toHaveBeenCalled();
    });

    it("reports success with no redirect when the response carries no data", async () => {
        completeSignInMock.mockResolvedValue({} as any);

        const { props } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalledWith(undefined));
    });

    it("reports a wrong code when the response is empty", async () => {
        completeSignInMock.mockResolvedValue(null as any);

        const { props } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The One-Time Password might be wrong" }),
            ),
        );
        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "3"));
    });

    it("reports a wrong code when the request throws", async () => {
        completeSignInMock.mockRejectedValue(new Error("boom"));

        const { props } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "The One-Time Password might be wrong" }),
            ),
        );
        expect(console.error).toHaveBeenCalled();
    });

    it("clears the passcode after a submission", async () => {
        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", ""));
    });

    it("reports rate limiting and recovers after the retry window", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 5 } as any);

        const { props } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await vi.waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "You have made too many requests" }),
            ),
        );
        await vi.waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "4"));

        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "0");
    });

    it("clears the rate limit timer on unmount", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 5 } as any);

        const { unmount } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));
        await vi.waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "4"));

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalled();
    });
});
