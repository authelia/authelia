// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
    completeDuoDeviceSelectionProcess,
    completePushNotificationSignIn,
    initiateDuoDeviceSelectionProcess,
} from "@services/PushNotification";
import { RateLimitedError } from "@services/RateLimit";
import SecondFactorMethodMobilePush from "@views/Settings/Common/SecondFactorMethodMobilePush";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@services/PushNotification", () => ({
    completeDuoDeviceSelectionProcess: vi.fn(),
    completePushNotificationSignIn: vi.fn(),
    initiateDuoDeviceSelectionProcess: vi.fn(),
}));

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@components/PushNotificationIcon", () => ({
    default: () => <div data-testid="push-notification-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/DeviceSelectionContainer", () => ({
    default: (props: any) => (
        <div data-testid="device-selection" data-devices={JSON.stringify(props.devices)}>
            <button data-testid="device-back" onClick={() => props.onBack()} />
            <button data-testid="device-select" onClick={() => props.onSelect({ id: "device-a", method: "push" })} />
        </div>
    ),
}));

const completeSignInMock = vi.mocked(completePushNotificationSignIn);
const initiateSelectionMock = vi.mocked(initiateDuoDeviceSelectionProcess);
const completeSelectionMock = vi.mocked(completeDuoDeviceSelectionProcess);

const authResponse = {
    devices: [
        { capabilities: ["push", "sms"], device: "device-a", display_name: "Phone" },
        { capabilities: ["push"], device: "device-b", display_name: "Tablet" },
    ],
    result: "auth",
};

function renderPush(onSecondFactorSuccess = vi.fn()) {
    return {
        ...render(<SecondFactorMethodMobilePush onSecondFactorSuccess={onSecondFactorSuccess} />),
        onSecondFactorSuccess,
    };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    completeSignInMock.mockRejectedValue(new Error("boom"));
    initiateSelectionMock.mockResolvedValue({} as any);
    completeSelectionMock.mockResolvedValue(undefined as any);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders push notification icon", async () => {
        renderPush();
        expect(screen.getByTestId("push-notification-icon")).toBeInTheDocument();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());
    });

    it("renders select a device link", async () => {
        renderPush();
        expect(screen.getByRole("button", { name: "Select a Device" })).toBeInTheDocument();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());
    });

    it("starts the push automatically", async () => {
        renderPush();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalledTimes(1));
    });
});

describe("push response handling", () => {
    it("reports success and hides the device link", async () => {
        completeSignInMock.mockResolvedValue({ result: "allow" } as any);

        const { onSecondFactorSuccess } = renderPush();

        await waitFor(() => expect(onSecondFactorSuccess).toHaveBeenCalled());
        expect(screen.queryByRole("button", { name: "Select a Device" })).not.toBeInTheDocument();
        expect(screen.getByTestId("push-notification-icon")).toBeInTheDocument();
    });

    it("shows the device selection when the response requires auth", async () => {
        completeSignInMock.mockResolvedValue(authResponse as any);

        renderPush();

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());
        expect(screen.getByTestId("device-selection")).toHaveAttribute(
            "data-devices",
            JSON.stringify([
                { id: "device-a", methods: ["push", "sms"], name: "Phone" },
                { id: "device-b", methods: ["push"], name: "Tablet" },
            ]),
        );
    });

    it("fails when the response requires enrollment", async () => {
        completeSignInMock.mockResolvedValue({ result: "enroll" } as any);

        renderPush();

        await waitFor(() => expect(mocks.createErrorNotification).toHaveBeenCalledWith("No compatible device found"));
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("fails when the Duo policy denies the request", async () => {
        completeSignInMock.mockResolvedValue({ result: "deny" } as any);

        renderPush();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("Device selection was denied by Duo policy"),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("reports success when the response carries no data", async () => {
        completeSignInMock.mockResolvedValue(undefined);

        const { onSecondFactorSuccess } = renderPush();

        await waitFor(() => expect(onSecondFactorSuccess).toHaveBeenCalled());
        expect(mocks.createErrorNotification).not.toHaveBeenCalled();
        expect(screen.queryByTestId("failure-icon")).not.toBeInTheDocument();
        expect(screen.queryByRole("button", { name: "Select a Device" })).not.toBeInTheDocument();
    });

    it("fails when the response carries an empty payload", async () => {
        completeSignInMock.mockResolvedValue({} as any);

        const { onSecondFactorSuccess } = renderPush();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue completing sign in process"),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
        expect(onSecondFactorSuccess).not.toHaveBeenCalled();
    });

    it("fails when the push request throws", async () => {
        completeSignInMock.mockRejectedValue(new Error("boom"));

        renderPush();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue completing sign in process"),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("reports rate limiting and recovers after the retry window", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockRejectedValue(new RateLimitedError(3));

        renderPush();

        await vi.waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You have made too many requests"),
        );
        expect(screen.queryByTestId("failure-icon")).not.toBeInTheDocument();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(3000);
        });

        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("retries the push from the failure state", async () => {
        completeSignInMock.mockRejectedValue(new Error("boom"));

        renderPush();

        await waitFor(() => expect(screen.getByTestId("failure-icon")).toBeInTheDocument());
        expect(completeSignInMock).toHaveBeenCalledTimes(1);

        fireEvent.click(screen.getByRole("button", { name: "Retry" }));

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalledTimes(2));
    });
});

describe("device selection", () => {
    it("opens the device selection from the link", async () => {
        initiateSelectionMock.mockResolvedValue({
            devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
            result: "auth",
        } as any);

        renderPush();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByRole("button", { name: "Select a Device" }));

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());
    });

    it("succeeds when the Duo policy bypasses selection", async () => {
        initiateSelectionMock.mockResolvedValue({ result: "allow" } as any);

        renderPush();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByRole("button", { name: "Select a Device" }));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("Device selection was bypassed by Duo policy"),
        );
        await waitFor(() => expect(screen.queryByRole("button", { name: "Select a Device" })).not.toBeInTheDocument());
    });

    it("fails when the Duo policy denies selection", async () => {
        initiateSelectionMock.mockResolvedValue({ result: "deny" } as any);

        renderPush();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByRole("button", { name: "Select a Device" }));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("Device selection was denied by Duo policy"),
        );
    });

    it("fails when the selection process reports enrollment", async () => {
        initiateSelectionMock.mockResolvedValue({ result: "enroll" } as any);

        renderPush();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByRole("button", { name: "Select a Device" }));

        await waitFor(() => expect(mocks.createErrorNotification).toHaveBeenCalledWith("No compatible device found"));
    });

    it("reports failures fetching Duo devices", async () => {
        initiateSelectionMock.mockRejectedValue(new Error("boom"));

        renderPush();
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByRole("button", { name: "Select a Device" }));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue fetching Duo device(s)"),
        );
    });

    it("restarts the push when the selection is dismissed", async () => {
        completeSignInMock.mockResolvedValue(authResponse as any);

        renderPush();
        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());
        expect(completeSignInMock).toHaveBeenCalledTimes(1);

        fireEvent.click(screen.getByTestId("device-back"));

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalledTimes(2));
    });

    it("persists the selected device and restarts the push", async () => {
        completeSignInMock.mockResolvedValue(authResponse as any);

        renderPush();
        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select"));

        await waitFor(() => expect(completeSelectionMock).toHaveBeenCalledWith({ device: "device-a", method: "push" }));
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalledTimes(2));
    });

    it("clears the rate limit when a device is selected during the retry window", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockRejectedValueOnce(new RateLimitedError(3));
        initiateSelectionMock.mockResolvedValue(authResponse as any);

        const { onSecondFactorSuccess } = renderPush();

        await vi.waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You have made too many requests"),
        );

        completeSignInMock.mockResolvedValue({ result: "allow" } as any);

        await act(async () => {
            fireEvent.click(screen.getByRole("button", { name: "Select a Device" }));
        });

        await vi.waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        await act(async () => {
            fireEvent.click(screen.getByTestId("device-select"));
        });

        await vi.waitFor(() => expect(onSecondFactorSuccess).toHaveBeenCalled());
        expect(document.getElementById("rate-limit-countdown")).not.toBeInTheDocument();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(3000);
        });

        expect(screen.queryByTestId("failure-icon")).not.toBeInTheDocument();
        expect(screen.queryByRole("button", { name: "Select a Device" })).not.toBeInTheDocument();
    });

    it("logs failures updating the preferred Duo device", async () => {
        completeSignInMock.mockResolvedValue(authResponse as any);
        completeSelectionMock.mockRejectedValue(new Error("boom"));

        renderPush();
        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select"));

        await waitFor(() =>
            expect(console.error).toHaveBeenCalledWith(
                expect.objectContaining({ message: "There was an issue updating preferred Duo device" }),
            ),
        );
        expect(screen.getByTestId("device-selection")).toBeInTheDocument();
    });
});

describe("cleanup", () => {
    it("clears the rate limit timer on unmount", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockRejectedValue(new RateLimitedError(3));

        const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout");

        const { unmount } = renderPush();

        await vi.waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You have made too many requests"),
        );

        const index = setTimeoutSpy.mock.calls.findIndex(([, timeout]) => timeout === 3000);

        expect(index).toBeGreaterThanOrEqual(0);

        const timerID = setTimeoutSpy.mock.results[index].value;

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalledWith(timerID);
    });
});
