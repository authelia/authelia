import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import { completeTOTPSignIn } from "@services/OneTimePassword";
import SecondFactorMethodOneTimePassword from "@views/Settings/Common/SecondFactorMethodOneTimePassword";

const mocks = vi.hoisted(() => ({
    config: { digits: 6, period: 30 } as any,
    configError: null as Error | null,
    createErrorNotification: vi.fn(),
    fetchConfig: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
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
            <button data-testid="dial-partial" onClick={() => props.onChange("123")} />
        </div>
    ),
    State: { Failure: 3, Idle: 0, InProgress: 1, RateLimited: 4, Success: 2 },
}));

const completeSignInMock = vi.mocked(completeTOTPSignIn);

function renderMethod(onSecondFactorSuccess = vi.fn()) {
    return {
        ...render(<SecondFactorMethodOneTimePassword onSecondFactorSuccess={onSecondFactorSuccess} />),
        onSecondFactorSuccess,
    };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.config = { digits: 6, period: 30 };
    mocks.configError = null;
    completeSignInMock.mockResolvedValue({} as any);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders OTP dial when config is loaded", () => {
        renderMethod();
        expect(screen.getByTestId("otp-dial")).toBeInTheDocument();
        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-digits", "6");
        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-period", "30");
    });

    it("renders loading page when config is not available", () => {
        mocks.config = undefined;

        renderMethod();

        expect(screen.getByTestId("loading-page")).toBeInTheDocument();
        expect(screen.queryByTestId("otp-dial")).not.toBeInTheDocument();
    });

    it("fetches the configuration on mount", () => {
        renderMethod();
        expect(mocks.fetchConfig).toHaveBeenCalled();
    });

    it("renders the loading page and logs when the configuration fails to load", () => {
        mocks.configError = new Error("boom");

        renderMethod();

        expect(screen.getByTestId("loading-page")).toBeInTheDocument();
        expect(console.error).toHaveBeenCalledWith(mocks.configError);
    });
});

describe("sign in", () => {
    it("submits once the full passcode is entered", async () => {
        const { onSecondFactorSuccess } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalledWith("123456"));
        expect(onSecondFactorSuccess).toHaveBeenCalled();
    });

    it("does not submit a partial passcode", async () => {
        renderMethod();

        fireEvent.click(screen.getByTestId("dial-partial"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123"));
        expect(completeSignInMock).not.toHaveBeenCalled();
    });

    it("clears the passcode after a submission", async () => {
        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", ""));
    });

    it("notifies and fails when the code is rejected", async () => {
        completeSignInMock.mockResolvedValue(null as any);

        const { onSecondFactorSuccess } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("The One-Time Password might be wrong"),
        );
        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "3"));
        expect(onSecondFactorSuccess).not.toHaveBeenCalled();
    });

    it("fails when the request throws", async () => {
        completeSignInMock.mockRejectedValue(new Error("boom"));

        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "3"));
        expect(console.error).toHaveBeenCalled();
    });

    it("reports rate limiting and recovers after the retry window", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 4 } as any);

        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await vi.waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You have made too many requests"),
        );
        await vi.waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "4"));

        await act(async () => {
            await vi.advanceTimersByTimeAsync(4000);
        });

        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "0");
    });

    it("ignores new submissions while rate limited", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 4 } as any);

        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));
        await vi.waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "4"));

        fireEvent.click(screen.getByTestId("dial-full"));

        expect(completeSignInMock).toHaveBeenCalledTimes(1);
    });

    it("ignores new submissions while one is in progress", async () => {
        let resolve: (value: unknown) => void = () => {};
        completeSignInMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));
        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "1"));

        fireEvent.click(screen.getByTestId("dial-full"));

        expect(completeSignInMock).toHaveBeenCalledTimes(1);

        resolve({});
        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "2"));
    });

    it("waits for the configured number of digits", async () => {
        mocks.config = { digits: 8, period: 30 };

        renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123456"));
        expect(completeSignInMock).not.toHaveBeenCalled();
    });

    it("clears the rate limit timer on unmount", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 4 } as any);

        const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout");

        const { unmount } = renderMethod();

        fireEvent.click(screen.getByTestId("dial-full"));
        await vi.waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "4"));

        // Identify the component's own rate limit timer rather than any timer React happens to clear.
        const index = setTimeoutSpy.mock.calls.findIndex(([, timeout]) => timeout === 4000);

        expect(index).toBeGreaterThanOrEqual(0);

        const timerID = setTimeoutSpy.mock.results[index].value;

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalledWith(timerID);
    });
});
