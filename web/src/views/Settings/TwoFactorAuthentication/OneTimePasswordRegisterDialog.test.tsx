// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { AxiosError, type AxiosResponse } from "axios";

import { TOTPAlgorithm } from "@models/TOTPConfiguration";
import { completeTOTPRegister, stopTOTPRegister } from "@services/OneTimePassword";
import { getTOTPSecret } from "@services/RegisterDevice";
import { getTOTPOptions } from "@services/UserInfoTOTPConfiguration";
import OneTimePasswordRegisterDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordRegisterDialog";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    createSuccessNotification: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
        createSuccessNotification: mocks.createSuccessNotification,
    }),
}));

vi.mock("@constants/constants", () => ({
    GoogleAuthenticator: { appleStore: "https://apple.example.com", googlePlay: "https://play.example.com" },
}));

vi.mock("@services/OneTimePassword", () => ({
    completeTOTPRegister: vi.fn(),
    stopTOTPRegister: vi.fn(),
}));

vi.mock("@services/RegisterDevice", () => ({
    getTOTPSecret: vi.fn(),
}));

vi.mock("@services/UserInfoTOTPConfiguration", () => ({
    getTOTPOptions: vi.fn(),
}));

vi.mock("@components/AppStoreBadges", () => ({
    default: () => <div data-testid="app-store-badges" />,
}));

vi.mock("@components/CopyButton", () => ({
    default: (props: any) => (
        <button data-testid="copy-button" data-value={props.value ?? ""}>
            {props.children}
        </button>
    ),
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/OTPDial", () => ({
    default: (props: any) => (
        <div
            data-testid="otp-dial"
            data-state={props.state}
            data-digits={props.digits}
            data-period={props.period}
            data-passcode={props.passcode}
        >
            <button data-testid="otp-enter-full" onClick={() => props.onChange("123456")} />
            <button data-testid="otp-enter-full-8" onClick={() => props.onChange("12345678")} />
            <button data-testid="otp-enter-partial" onClick={() => props.onChange("123")} />
        </div>
    ),
    State: { Failure: 3, Idle: 0, InProgress: 1, RateLimited: 4, Success: 2 },
}));

vi.mock("qrcode.react", () => ({
    QRCodeSVG: (props: any) => <div data-testid="qr-code" data-value={props.value} />,
}));

const getTOTPOptionsMock = vi.mocked(getTOTPOptions);
const getTOTPSecretMock = vi.mocked(getTOTPSecret);
const completeTOTPRegisterMock = vi.mocked(completeTOTPRegister);
const stopTOTPRegisterMock = vi.mocked(stopTOTPRegister);

const singleOptions = {
    algorithm: TOTPAlgorithm.SHA1,
    algorithms: [TOTPAlgorithm.SHA1],
    length: 6,
    lengths: [6],
    period: 30,
    periods: [30],
} as any;

const multipleOptions = {
    algorithm: TOTPAlgorithm.SHA1,
    algorithms: [TOTPAlgorithm.SHA1, TOTPAlgorithm.SHA256, TOTPAlgorithm.SHA512],
    length: 6,
    lengths: [6, 8],
    period: 30,
    periods: [30, 60],
} as any;

const secret = { base32_secret: "BASE32SECRET", otpauth_url: "otpauth://totp/Authelia:john?secret=BASE32SECRET" };

function renderDialog(props: Partial<{ open: boolean; setClosed: () => void }> = {}) {
    const setClosed = props.setClosed ?? vi.fn();
    return { ...render(<OneTimePasswordRegisterDialog open={props.open ?? true} setClosed={setClosed} />), setClosed };
}

function next() {
    fireEvent.click(document.getElementById("dialog-next") as HTMLButtonElement);
}

function previous() {
    fireEvent.click(document.getElementById("dialog-previous") as HTMLButtonElement);
}

function cancel() {
    fireEvent.click(document.getElementById("dialog-cancel") as HTMLButtonElement);
}

async function advanceToRegisterStep() {
    await screen.findByText("To begin select next");
    next();
    await waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalled());
}

async function advanceToConfirmStep() {
    await advanceToRegisterStep();
    next();
    await screen.findByTestId("otp-dial");
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    getTOTPOptionsMock.mockResolvedValue(singleOptions);
    getTOTPSecretMock.mockResolvedValue(secret as any);
    completeTOTPRegisterMock.mockResolvedValue(undefined as any);
    stopTOTPRegisterMock.mockResolvedValue(undefined as any);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders dialog with title when open", async () => {
        renderDialog();
        expect(screen.getByText("Register {{item}}")).toBeInTheDocument();
        expect(screen.getByText("Start")).toBeInTheDocument();
        await waitFor(() => expect(getTOTPOptionsMock).toHaveBeenCalled());
    });

    it("does not render content when closed", () => {
        renderDialog({ open: false });
        expect(screen.queryByText("Register {{item}}")).not.toBeInTheDocument();
        expect(getTOTPOptionsMock).not.toHaveBeenCalled();
    });

    it("shows a loading message until the options arrive", async () => {
        let resolve: (value: unknown) => void = () => {};
        getTOTPOptionsMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderDialog();

        expect(screen.getByText("Loading")).toBeInTheDocument();

        await act(async () => {
            resolve(singleOptions);
        });

        expect(await screen.findByText("To begin select next")).toBeInTheDocument();
    });

    it("disables previous on the first step", async () => {
        renderDialog();
        await screen.findByText("To begin select next");
        expect(document.getElementById("dialog-previous")).toBeDisabled();
        expect(document.getElementById("dialog-next")).not.toBeDisabled();
    });
});

describe("advanced options", () => {
    it("keeps the advanced switch disabled when only one of each option exists", async () => {
        renderDialog();
        await screen.findByText("To begin select next");

        expect(document.getElementById("one-time-password-advanced")).toBeDisabled();
    });

    it("enables the advanced switch when multiple options exist", async () => {
        getTOTPOptionsMock.mockResolvedValue(multipleOptions);

        renderDialog();
        await screen.findByText("To begin select next");

        expect(document.getElementById("one-time-password-advanced")).not.toBeDisabled();
    });

    it("reveals the algorithm, length and period pickers", async () => {
        getTOTPOptionsMock.mockResolvedValue(multipleOptions);

        renderDialog();
        await screen.findByText("To begin select next");

        fireEvent.click(document.getElementById("one-time-password-advanced") as HTMLElement);

        await waitFor(() => expect(document.getElementById("one-time-password-algorithm-SHA256")).toBeTruthy());
        expect(document.getElementById("one-time-password-length-8")).toBeTruthy();
        expect(document.getElementById("one-time-password-period-60")).toBeTruthy();
    });

    it("hides pickers that only offer a single value", async () => {
        getTOTPOptionsMock.mockResolvedValue({
            ...multipleOptions,
            lengths: [6],
            periods: [30],
        });

        renderDialog();
        await screen.findByText("To begin select next");

        fireEvent.click(document.getElementById("one-time-password-advanced") as HTMLElement);

        await waitFor(() => expect(document.getElementById("one-time-password-algorithm-SHA256")).toBeTruthy());
        expect(screen.queryByText("Length")).not.toBeInTheDocument();
        expect(screen.queryByText("Seconds")).not.toBeInTheDocument();
    });

    it("requests the secret with the selected algorithm", async () => {
        getTOTPOptionsMock.mockResolvedValue(multipleOptions);

        renderDialog();
        await screen.findByText("To begin select next");
        fireEvent.click(document.getElementById("one-time-password-advanced") as HTMLElement);
        await waitFor(() => expect(document.getElementById("one-time-password-algorithm-SHA512")).toBeTruthy());

        fireEvent.click(document.getElementById("one-time-password-algorithm-SHA512") as HTMLElement);
        next();

        await waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalledWith("SHA512", 6, 30));
    });

    it("requests the secret with the selected length", async () => {
        getTOTPOptionsMock.mockResolvedValue(multipleOptions);

        renderDialog();
        await screen.findByText("To begin select next");
        fireEvent.click(document.getElementById("one-time-password-advanced") as HTMLElement);
        await waitFor(() => expect(document.getElementById("one-time-password-length-8")).toBeTruthy());

        fireEvent.click(document.getElementById("one-time-password-length-8") as HTMLElement);
        next();

        await waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalledWith("SHA1", 8, 30));
    });

    it("requests the secret with the selected period", async () => {
        getTOTPOptionsMock.mockResolvedValue(multipleOptions);

        renderDialog();
        await screen.findByText("To begin select next");
        fireEvent.click(document.getElementById("one-time-password-advanced") as HTMLElement);
        await waitFor(() => expect(document.getElementById("one-time-password-period-60")).toBeTruthy());

        fireEvent.click(document.getElementById("one-time-password-period-60") as HTMLElement);
        next();

        await waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalledWith("SHA1", 6, 60));
    });

    it("collapses the advanced panel when moving between steps", async () => {
        getTOTPOptionsMock.mockResolvedValue(multipleOptions);

        renderDialog();
        await screen.findByText("To begin select next");
        fireEvent.click(document.getElementById("one-time-password-advanced") as HTMLElement);
        await waitFor(() => expect(document.getElementById("one-time-password-advanced")).toBeChecked());

        next();
        await waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalled());
        previous();

        await waitFor(() => expect(document.getElementById("one-time-password-advanced")).not.toBeChecked());
    });
});

describe("register step", () => {
    it("shows the QR code for the returned secret", async () => {
        renderDialog();
        await advanceToRegisterStep();

        expect(await screen.findByTestId("qr-code")).toHaveAttribute("data-value", secret.otpauth_url);
    });

    it("offers the app store badges", async () => {
        renderDialog();
        await advanceToRegisterStep();

        expect(screen.getByTestId("app-store-badges")).toBeInTheDocument();
    });

    it("exposes the URI and secret through copy buttons", async () => {
        renderDialog();
        await advanceToRegisterStep();

        fireEvent.click(document.getElementById("qr-toggle") as HTMLElement);

        await waitFor(() => expect(screen.getAllByTestId("copy-button")).toHaveLength(2));
        const values = screen.getAllByTestId("copy-button").map((el) => el.getAttribute("data-value"));
        expect(values).toContain(secret.otpauth_url);
        expect(values).toContain(secret.base32_secret);
        expect(document.getElementById("secret-url")).toHaveValue(secret.otpauth_url);
    });

    it("notifies about a device mismatch on a 403", async () => {
        getTOTPSecretMock.mockRejectedValue(
            new AxiosError("Request failed with status code 403", "ERR_BAD_REQUEST", undefined, undefined, {
                status: 403,
            } as AxiosResponse),
        );

        renderDialog();
        await advanceToRegisterStep();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "You must use the code from the same device and browser that initiated the process",
            ),
        );
    });

    it("notifies about an expired code on other failures", async () => {
        getTOTPSecretMock.mockRejectedValue(new Error("boom"));

        renderDialog();
        await advanceToRegisterStep();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Failed to register device, the provided code is expired or has already been used",
            ),
        );
        expect(screen.queryByTestId("qr-code")).not.toBeInTheDocument();
    });
});

describe("confirm step", () => {
    it("shows the dial with the selected digits and period", async () => {
        renderDialog();
        await advanceToConfirmStep();

        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-digits", "6");
        expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-period", "30");
        expect(document.getElementById("dialog-next")).toBeDisabled();
    });

    it("does not submit a partial code", async () => {
        renderDialog();
        await advanceToConfirmStep();

        fireEvent.click(screen.getByTestId("otp-enter-partial"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123"));
        expect(completeTOTPRegisterMock).not.toHaveBeenCalled();
    });

    it("registers a complete code and reports success", async () => {
        vi.useFakeTimers();

        const setClosed = vi.fn();
        renderDialog({ setClosed });

        await vi.waitFor(() => expect(screen.getByText("To begin select next")).toBeInTheDocument());
        next();
        await vi.waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalled());
        next();
        await vi.waitFor(() => expect(screen.getByTestId("otp-dial")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("otp-enter-full"));

        await vi.waitFor(() => expect(completeTOTPRegisterMock).toHaveBeenCalledWith("123456"));
        await vi.waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());

        await act(async () => {
            await vi.advanceTimersByTimeAsync(750);
        });

        expect(mocks.createSuccessNotification).toHaveBeenCalledWith("Successfully {{action}} the {{item}}");
        expect(setClosed).toHaveBeenCalled();
    });

    it("marks the dial as failed when registration is rejected", async () => {
        completeTOTPRegisterMock.mockRejectedValue(new Error("bad code"));

        renderDialog();
        await advanceToConfirmStep();

        fireEvent.click(screen.getByTestId("otp-enter-full"));

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-state", "3"));
        expect(console.error).toHaveBeenCalled();
    });

    it("waits for the configured number of digits", async () => {
        getTOTPOptionsMock.mockResolvedValue(multipleOptions);

        renderDialog();
        await screen.findByText("To begin select next");
        fireEvent.click(document.getElementById("one-time-password-advanced") as HTMLElement);
        await waitFor(() => expect(document.getElementById("one-time-password-length-8")).toBeTruthy());
        fireEvent.click(document.getElementById("one-time-password-length-8") as HTMLElement);

        next();
        await waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalledWith("SHA1", 8, 30));
        next();
        await screen.findByTestId("otp-dial");

        fireEvent.click(screen.getByTestId("otp-enter-full"));
        await waitFor(() => expect(screen.getByTestId("otp-dial")).toHaveAttribute("data-passcode", "123456"));
        expect(completeTOTPRegisterMock).not.toHaveBeenCalled();

        fireEvent.click(screen.getByTestId("otp-enter-full-8"));
        await waitFor(() => expect(completeTOTPRegisterMock).toHaveBeenCalledWith("12345678"));
    });
});

describe("closing", () => {
    it("closes without stopping registration before a secret is issued", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });
        await screen.findByText("To begin select next");

        cancel();

        await waitFor(() => expect(setClosed).toHaveBeenCalled());
        expect(stopTOTPRegisterMock).not.toHaveBeenCalled();
    });

    it("stops the registration when a secret was already issued", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });
        await advanceToRegisterStep();

        cancel();

        await waitFor(() => expect(stopTOTPRegisterMock).toHaveBeenCalled());
        expect(setClosed).toHaveBeenCalled();
    });

    it("logs failures stopping the registration", async () => {
        stopTOTPRegisterMock.mockRejectedValue(new Error("boom"));

        renderDialog();
        await advanceToRegisterStep();

        cancel();

        await waitFor(() => expect(console.error).toHaveBeenCalled());
    });

    it("resets back to the first step after closing", async () => {
        renderDialog();
        await advanceToRegisterStep();

        cancel();

        await waitFor(() => expect(screen.getByText("To begin select next")).toBeInTheDocument());
        expect(document.getElementById("dialog-previous")).toBeDisabled();
    });
});

describe("dismissal", () => {
    it("closes when Escape is pressed", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });
        await screen.findByText("To begin select next");

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        await waitFor(() => expect(setClosed).toHaveBeenCalled());
    });

    it("stops the registration when Escape is pressed after a secret was issued", async () => {
        renderDialog();
        await advanceToRegisterStep();

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        await waitFor(() => expect(stopTOTPRegisterMock).toHaveBeenCalled());
    });
});

describe("step guards", () => {
    it("ignores previous on the first step", async () => {
        renderDialog();
        await screen.findByText("To begin select next");

        previous();

        expect(screen.getByText("To begin select next")).toBeInTheDocument();
        expect(getTOTPSecretMock).not.toHaveBeenCalled();
    });

    it("ignores next on the last step", async () => {
        renderDialog();
        await advanceToConfirmStep();

        next();

        await waitFor(() => expect(screen.getByTestId("otp-dial")).toBeInTheDocument());
    });

    it("does not re-request the options once they are loaded", async () => {
        renderDialog();
        await advanceToRegisterStep();

        previous();

        await waitFor(() => expect(screen.getByText("To begin select next")).toBeInTheDocument());
        expect(getTOTPOptionsMock).toHaveBeenCalledTimes(1);
    });
});

describe("cleanup", () => {
    it("clears the completion timer on unmount", async () => {
        vi.useFakeTimers();

        const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout");

        const { unmount } = renderDialog();

        await vi.waitFor(() => expect(screen.getByText("To begin select next")).toBeInTheDocument());
        next();
        await vi.waitFor(() => expect(getTOTPSecretMock).toHaveBeenCalled());
        next();
        await vi.waitFor(() => expect(screen.getByTestId("otp-dial")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("otp-enter-full"));

        await vi.waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());

        // Identify the component's own completion timer rather than any timer React happens to clear.
        const index = setTimeoutSpy.mock.calls.findIndex(([, timeout]) => timeout === 750);

        expect(index).toBeGreaterThanOrEqual(0);

        const timerID = setTimeoutSpy.mock.results[index].value;

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalledWith(timerID);
    });
});
