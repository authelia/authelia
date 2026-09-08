import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";

const mocks = vi.hoisted(() => ({
    supportsWebAuthn: true,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@simplewebauthn/browser", () => ({
    browserSupportsWebAuthn: () => mocks.supportsWebAuthn,
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/PasswordForm", () => ({
    default: (props: any) => (
        <div data-testid="password-form">
            <button data-testid="password-success" onClick={() => props.onAuthenticationSuccess()} />
        </div>
    ),
}));

vi.mock("@views/Settings/Common/SecondFactorMethodOneTimePassword", () => ({
    default: (props: any) => (
        <div data-testid="method-otp">
            <button data-testid="otp-success" onClick={() => props.onSecondFactorSuccess()} />
        </div>
    ),
}));

vi.mock("@views/Settings/Common/SecondFactorMethodWebAuthn", () => ({
    default: (props: any) => (
        <div data-testid="method-webauthn">
            <button data-testid="webauthn-success" onClick={() => props.onSecondFactorSuccess()} />
        </div>
    ),
}));

vi.mock("@views/Settings/Common/SecondFactorMethodMobilePush", () => ({
    default: (props: any) => (
        <div data-testid="method-push">
            <button data-testid="push-success" onClick={() => props.onSecondFactorSuccess()} />
        </div>
    ),
}));

const elevation = {
    can_skip_second_factor: true,
    elevated: false,
    factor_knowledge: true,
    require_second_factor: true,
    skip_second_factor: false,
} as any;

const info = {
    has_duo: false,
    has_totp: true,
    has_webauthn: true,
} as any;

function renderDialog(
    props: Partial<{
        elevation: any;
        handleClosed: (ok: boolean, changed: boolean) => void;
        handleOpened: () => void;
        info: any;
        opening: boolean;
    }> = {},
) {
    const handleClosed = props.handleClosed ?? vi.fn();
    const handleOpened = props.handleOpened ?? vi.fn();
    const merged = { elevation, info, opening: true, ...props, handleClosed, handleOpened };

    return { ...render(<SecondFactorDialog {...merged} />), handleClosed, handleOpened };
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.supportsWebAuthn = true;
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders dialog with stepper when opening with elevation", () => {
        renderDialog();
        expect(screen.getByText("Identity Verification")).toBeInTheDocument();
        expect(screen.getByText("Select a Method")).toBeInTheDocument();
        expect(screen.getByText("Authenticate")).toBeInTheDocument();
        expect(screen.getByText("Completed")).toBeInTheDocument();
    });

    it("renders method buttons for available methods", () => {
        renderDialog();
        expect(screen.getByText("One-Time Password")).toBeInTheDocument();
        expect(screen.getByText("WebAuthn")).toBeInTheDocument();
        expect(screen.queryByText("Mobile Push")).not.toBeInTheDocument();
    });

    it("does not render content when not opening", () => {
        renderDialog({ opening: false });
        expect(screen.queryByText("Identity Verification")).not.toBeInTheDocument();
    });

    it("notifies the parent that it opened", () => {
        const { handleOpened } = renderDialog();
        expect(handleOpened).toHaveBeenCalled();
    });

    it("renders the mobile push option when Duo is registered", () => {
        renderDialog({ info: { has_duo: true, has_totp: false, has_webauthn: false } });
        expect(screen.getByText("Mobile Push")).toBeInTheDocument();
        expect(screen.queryByText("One-Time Password")).not.toBeInTheDocument();
    });

    it("hides WebAuthn when the browser does not support it", () => {
        mocks.supportsWebAuthn = false;

        renderDialog();

        expect(screen.queryByText("WebAuthn")).not.toBeInTheDocument();
        expect(screen.getByText("One-Time Password")).toBeInTheDocument();
    });

    it("hides the one-time code option when it cannot be skipped", () => {
        renderDialog({ elevation: { ...elevation, can_skip_second_factor: false } });
        expect(screen.queryByText("Email One-Time Code")).not.toBeInTheDocument();
    });

    it("shows a loading page while the user info is unavailable", () => {
        renderDialog({ info: undefined });
        expect(screen.getByTestId("loading-page")).toBeInTheDocument();
    });

    it("stays closed without an elevation", () => {
        renderDialog({ elevation: undefined });
        expect(screen.queryByText("Identity Verification")).not.toBeInTheDocument();
    });
});

describe("skipping", () => {
    it("closes immediately when the second factor can be skipped", () => {
        const { handleClosed } = renderDialog({
            elevation: { ...elevation, can_skip_second_factor: false, skip_second_factor: true },
        });

        expect(handleClosed).toHaveBeenCalledWith(true, false);
        expect(screen.queryByText("Identity Verification")).not.toBeInTheDocument();
    });

    it("closes immediately when a second factor is not required", () => {
        const { handleClosed } = renderDialog({
            elevation: { ...elevation, can_skip_second_factor: false, require_second_factor: false },
        });

        expect(handleClosed).toHaveBeenCalledWith(true, false);
    });

    it("does not skip when the user may choose to skip", () => {
        const { handleClosed } = renderDialog({
            elevation: { ...elevation, can_skip_second_factor: true, skip_second_factor: true },
        });

        expect(handleClosed).not.toHaveBeenCalled();
        expect(screen.getByText("Identity Verification")).toBeInTheDocument();
    });

    it("jumps straight to authentication when the knowledge factor is missing", () => {
        renderDialog({ elevation: { ...elevation, factor_knowledge: false } });

        expect(screen.getByTestId("password-form")).toBeInTheDocument();
        expect(screen.queryByText("One-Time Password")).not.toBeInTheDocument();
    });
});

describe("method selection", () => {
    it("shows the one-time password method", async () => {
        renderDialog();

        fireEvent.click(screen.getByText("One-Time Password"));

        expect(await screen.findByTestId("method-otp")).toBeInTheDocument();
    });

    it("shows the WebAuthn method", async () => {
        renderDialog();

        fireEvent.click(screen.getByText("WebAuthn"));

        expect(await screen.findByTestId("method-webauthn")).toBeInTheDocument();
    });

    it("shows the mobile push method", async () => {
        renderDialog({ info: { has_duo: true, has_totp: false, has_webauthn: false } });

        fireEvent.click(screen.getByText("Mobile Push"));

        expect(await screen.findByTestId("method-push")).toBeInTheDocument();
    });

    it("closes with a successful, unchanged result for the email one-time code", () => {
        const { handleClosed } = renderDialog();

        fireEvent.click(screen.getByText("Email One-Time Code"));

        expect(handleClosed).toHaveBeenCalledWith(true, false);
    });
});

describe("authentication result", () => {
    it("reports success after the completion delay", async () => {
        vi.useFakeTimers();

        const handleClosed = vi.fn();
        renderDialog({ handleClosed });

        await act(async () => {
            fireEvent.click(screen.getByText("One-Time Password"));
        });

        expect(screen.getByTestId("method-otp")).toBeInTheDocument();

        await act(async () => {
            fireEvent.click(screen.getByTestId("otp-success"));
        });

        expect(screen.getByTestId("success-icon")).toBeInTheDocument();
        expect(handleClosed).not.toHaveBeenCalled();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1500);
        });

        expect(handleClosed).toHaveBeenCalledWith(true, true);
    });

    it("reports success from the password form", async () => {
        vi.useFakeTimers();

        const handleClosed = vi.fn();
        renderDialog({ elevation: { ...elevation, factor_knowledge: false }, handleClosed });

        fireEvent.click(screen.getByTestId("password-success"));

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1500);
        });

        expect(handleClosed).toHaveBeenCalledWith(true, true);
    });

    it("reports success from the WebAuthn method", async () => {
        vi.useFakeTimers();

        const handleClosed = vi.fn();
        renderDialog({ handleClosed });

        await act(async () => {
            fireEvent.click(screen.getByText("WebAuthn"));
        });

        expect(screen.getByTestId("method-webauthn")).toBeInTheDocument();

        fireEvent.click(screen.getByTestId("webauthn-success"));

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1500);
        });

        expect(handleClosed).toHaveBeenCalledWith(true, true);
    });

    it("ignores further method selections while closing", async () => {
        vi.useFakeTimers();

        renderDialog();

        await act(async () => {
            fireEvent.click(screen.getByText("One-Time Password"));
        });

        expect(screen.getByTestId("method-otp")).toBeInTheDocument();
        await act(async () => {
            fireEvent.click(screen.getByTestId("otp-success"));
        });

        expect(screen.getByTestId("success-icon")).toBeInTheDocument();

        expect(screen.queryByText("WebAuthn")).not.toBeInTheDocument();
    });
});

describe("cancelling", () => {
    it("closes with a failure result", () => {
        const { handleClosed } = renderDialog();

        fireEvent.click(screen.getByText("Cancel"));

        expect(handleClosed).toHaveBeenCalledWith(false, false);
    });

    it("returns to the first step after cancelling", async () => {
        const { handleClosed, rerender } = renderDialog();

        fireEvent.click(screen.getByText("One-Time Password"));
        await screen.findByTestId("method-otp");

        fireEvent.click(screen.getByText("Cancel"));
        expect(handleClosed).toHaveBeenCalledWith(false, false);

        rerender(
            <SecondFactorDialog
                elevation={elevation}
                info={info}
                opening={true}
                handleClosed={handleClosed}
                handleOpened={vi.fn()}
            />,
        );

        await waitFor(() => expect(screen.getByText("One-Time Password")).toBeInTheDocument());
    });
});

describe("edge cases", () => {
    it("renders nothing for the auth step when no method was selected", () => {
        renderDialog({ elevation: { ...elevation, can_skip_second_factor: false } });

        fireEvent.click(screen.getByText("One-Time Password"));
        fireEvent.click(screen.getByText("Cancel"));

        expect(screen.queryByTestId("method-otp")).not.toBeInTheDocument();
    });

    it("keeps the dialog open across a re-render with the same elevation", async () => {
        const handleOpened = vi.fn();
        const { rerender } = renderDialog({ handleOpened });

        expect(handleOpened).toHaveBeenCalledTimes(1);

        rerender(
            <SecondFactorDialog
                elevation={elevation}
                info={info}
                opening={true}
                handleClosed={vi.fn()}
                handleOpened={handleOpened}
            />,
        );

        await waitFor(() => expect(handleOpened).toHaveBeenCalledTimes(1));
    });

    it("does not react while it is closing", async () => {
        vi.useFakeTimers();

        const handleClosed = vi.fn();
        renderDialog({ handleClosed });

        await act(async () => {
            fireEvent.click(screen.getByText("One-Time Password"));
        });

        expect(screen.getByTestId("method-otp")).toBeInTheDocument();

        await act(async () => {
            fireEvent.click(screen.getByTestId("otp-success"));
        });

        expect(screen.getByTestId("success-icon")).toBeInTheDocument();

        expect(handleClosed).not.toHaveBeenCalled();
    });

    it("renders success when reaching the completion step from the push method", async () => {
        vi.useFakeTimers();

        const handleClosed = vi.fn();
        renderDialog({ handleClosed, info: { has_duo: true, has_totp: false, has_webauthn: false } });

        await act(async () => {
            fireEvent.click(screen.getByText("Mobile Push"));
        });

        expect(screen.getByTestId("method-push")).toBeInTheDocument();

        fireEvent.click(screen.getByTestId("push-success"));

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1500);
        });

        expect(handleClosed).toHaveBeenCalledWith(true, true);
    });
});

describe("cleanup", () => {
    it("clears the success timer on unmount", async () => {
        vi.useFakeTimers();

        const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout");

        const { unmount } = renderDialog();

        await act(async () => {
            fireEvent.click(screen.getByText("One-Time Password"));
        });

        await act(async () => {
            fireEvent.click(screen.getByTestId("otp-success"));
        });

        expect(screen.getByTestId("success-icon")).toBeInTheDocument();

        // Identify the component's own success timer rather than any timer React happens to clear.
        const index = setTimeoutSpy.mock.calls.findIndex(([, timeout]) => timeout === 1500);

        expect(index).toBeGreaterThanOrEqual(0);

        const timerID = setTimeoutSpy.mock.results[index].value;

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalledWith(timerID);
    });
});
