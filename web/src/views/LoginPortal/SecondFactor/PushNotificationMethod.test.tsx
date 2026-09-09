// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
    completeDuoDeviceSelectionProcess,
    completePushNotificationSignIn,
    getPreferredDuoDevice,
    initiateDuoDeviceSelectionProcess,
} from "@services/PushNotification";
import { AuthenticationLevel } from "@services/State";
import PushNotificationMethod from "@views/LoginPortal/SecondFactor/PushNotificationMethod";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: null, id: null, subflow: null }),
    useFlowPresent: () => false,
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

vi.mock("@services/PushNotification", () => ({
    completeDuoDeviceSelectionProcess: vi.fn(),
    completePushNotificationSignIn: vi.fn(),
    getPreferredDuoDevice: vi.fn(),
    initiateDuoDeviceSelectionProcess: vi.fn(),
}));

vi.mock("@components/FailureIcon", () => ({
    default: () => <div data-testid="failure-icon" />,
}));

vi.mock("@components/PushNotificationIcon", () => ({
    default: () => <div data-testid="push-notification-icon" />,
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
            data-self-enrollment={String(props.duoSelfEnrollment)}
            data-registered={String(props.registered)}
        >
            <button data-testid="method-select" onClick={() => props.onSelectClick()} />
            <button data-testid="method-register" onClick={() => props.onRegisterClick()} />
            {props.children}
        </div>
    ),
    State: { ALREADY_AUTHENTICATED: "ALREADY_AUTHENTICATED", METHOD: "METHOD", NOT_REGISTERED: "NOT_REGISTERED" },
}));

vi.mock("@views/LoginPortal/SecondFactor/DeviceSelectionContainer", () => ({
    default: (props: any) => (
        <div data-testid="device-selection" data-devices={JSON.stringify(props.devices)}>
            <button data-testid="device-back" onClick={() => props.onBack()} />
            <button data-testid="device-select-a" onClick={() => props.onSelect({ id: "device-a", method: "push" })} />
            <button data-testid="device-select-b" onClick={() => props.onSelect({ id: "device-b", method: "sms" })} />
        </div>
    ),
}));

const completeSignInMock = vi.mocked(completePushNotificationSignIn);
const getPreferredMock = vi.mocked(getPreferredDuoDevice);
const initiateSelectionMock = vi.mocked(initiateDuoDeviceSelectionProcess);
const completeSelectionMock = vi.mocked(completeDuoDeviceSelectionProcess);

const defaultProps = {
    authenticationLevel: AuthenticationLevel.OneFactor,
    duoSelfEnrollment: false,
    id: "push-method",
    onSelectionClick: vi.fn(),
    onSignInError: vi.fn(),
    onSignInSuccess: vi.fn(),
    registered: true,
};

function renderMethod(props: Partial<typeof defaultProps> = {}) {
    const merged = {
        ...defaultProps,
        onSelectionClick: vi.fn(),
        onSignInError: vi.fn(),
        onSignInSuccess: vi.fn(),
        ...props,
    };

    return { ...render(<PushNotificationMethod {...merged} />), props: merged };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.spyOn(console, "debug").mockImplementation(() => {});
    getPreferredMock.mockResolvedValue({} as any);
    completeSignInMock.mockResolvedValue(null as any);
    initiateSelectionMock.mockResolvedValue({} as any);
    completeSelectionMock.mockResolvedValue(undefined as any);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders method container with push notification title", async () => {
        renderMethod();
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-title", "Push Notification");
        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());
    });

    it("renders already authenticated state at two factor level", () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "ALREADY_AUTHENTICATED");
    });

    it("shows the success icon when already authenticated", () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });
        expect(screen.getByTestId("success-icon")).toBeInTheDocument();
    });

    it("shows the animated push icon while signing in", () => {
        renderMethod();
        expect(screen.getByTestId("push-notification-icon")).toBeInTheDocument();
    });

    it("does not sign in when already at two factor level", async () => {
        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });
        await waitFor(() => expect(getPreferredMock).not.toHaveBeenCalled());
        expect(completeSignInMock).not.toHaveBeenCalled();
    });
});

describe("preferred device lookup", () => {
    it("fetches the preferred device below two factor level", async () => {
        getPreferredMock.mockResolvedValue({ preferred_device: "device-a", preferred_method: "push" } as any);

        renderMethod();

        await waitFor(() => expect(getPreferredMock).toHaveBeenCalled());
    });

    it("ignores a response without a preferred device", async () => {
        getPreferredMock.mockResolvedValue({ preferred_device: "device-a" } as any);

        renderMethod();

        await waitFor(() => expect(getPreferredMock).toHaveBeenCalled());
        expect(screen.getByTestId("method-container")).toBeInTheDocument();
    });

    it("logs but does not surface preferred device lookup failures", async () => {
        getPreferredMock.mockRejectedValue(new Error("nope"));

        const { props } = renderMethod();

        await waitFor(() => expect(console.debug).toHaveBeenCalled());
        expect(props.onSignInError).not.toHaveBeenCalledWith(
            expect.objectContaining({ message: "There was an issue fetching Duo device(s)" }),
        );
    });
});

describe("sign in", () => {
    it("reports success after the success delay", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ data: { redirect: "https://example.com" } } as any);

        const { props } = renderMethod();

        await vi.waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());
        expect(props.onSignInSuccess).not.toHaveBeenCalled();

        await vi.advanceTimersByTimeAsync(1500);

        expect(props.onSignInSuccess).toHaveBeenCalledWith("https://example.com");
    });

    it("reports success with no redirect when the response carries no data", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ data: undefined } as any);

        const { props } = renderMethod();

        await vi.waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());

        await vi.advanceTimersByTimeAsync(1500);

        expect(props.onSignInSuccess).toHaveBeenCalledWith(undefined);
    });

    it("fails when the sign in response is missing", async () => {
        completeSignInMock.mockResolvedValue(null as any);

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "There was an issue completing sign in process" }),
            ),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("fails when the sign in request throws", async () => {
        completeSignInMock.mockRejectedValue(new Error("boom"));

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "There was an issue completing sign in process" }),
            ),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("swallows cancelled sign in requests", async () => {
        completeSignInMock.mockRejectedValue({ __CANCEL__: true });

        const { props } = renderMethod();

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());
        expect(props.onSignInError).not.toHaveBeenCalled();
        expect(screen.queryByTestId("failure-icon")).not.toBeInTheDocument();
    });

    it("shows the device selection when the result requires auth", async () => {
        completeSignInMock.mockResolvedValue({
            data: {
                devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
                result: "auth",
            },
        } as any);

        renderMethod();

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());
        expect(screen.getByTestId("device-selection")).toHaveAttribute(
            "data-devices",
            JSON.stringify([{ id: "device-a", methods: ["push"], name: "Phone" }]),
        );
    });

    it("moves to the enroll state when no compatible device exists", async () => {
        completeSignInMock.mockResolvedValue({
            data: { enroll_url: "https://enroll.example.com", result: "enroll" },
        } as any);

        const { props } = renderMethod({ duoSelfEnrollment: true });

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "No compatible device found" }),
            ),
        );
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "NOT_REGISTERED");
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-self-enrollment", "true");
    });

    it("does not offer self enrollment when it is disabled", async () => {
        completeSignInMock.mockResolvedValue({
            data: { enroll_url: "https://enroll.example.com", result: "enroll" },
        } as any);

        renderMethod({ duoSelfEnrollment: false });

        await waitFor(() =>
            expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "NOT_REGISTERED"),
        );
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-self-enrollment", "false");
    });

    it("does not offer self enrollment when no enroll url is returned", async () => {
        completeSignInMock.mockResolvedValue({ data: { result: "enroll" } } as any);

        renderMethod({ duoSelfEnrollment: true });

        await waitFor(() =>
            expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "NOT_REGISTERED"),
        );
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-self-enrollment", "false");
    });

    it("fails when the Duo policy denies the device", async () => {
        completeSignInMock.mockResolvedValue({ data: { result: "deny" } } as any);

        const { props } = renderMethod();

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Device selection was denied by Duo policy" }),
            ),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("reports rate limiting and recovers after the retry window", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 5 } as any);

        const { props } = renderMethod();

        await vi.waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "You have made too many requests" }),
            ),
        );
        expect(screen.queryByTestId("failure-icon")).not.toBeInTheDocument();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("retries the sign in from the failure state", async () => {
        completeSignInMock.mockResolvedValue(null as any);

        renderMethod();

        await waitFor(() => expect(screen.getByTestId("failure-icon")).toBeInTheDocument());
        expect(completeSignInMock).toHaveBeenCalledTimes(1);

        fireEvent.click(screen.getByRole("button", { name: "Retry" }));

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalledTimes(2));
    });
});

describe("device selection", () => {
    it("opens device selection from the method container", async () => {
        initiateSelectionMock.mockResolvedValue({
            devices: [{ capabilities: ["push", "sms"], device: "device-a", display_name: "Phone" }],
            result: "auth",
        } as any);

        renderMethod();

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("method-select"));

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());
    });

    it("records the preferred device returned by the selection process", async () => {
        initiateSelectionMock.mockResolvedValue({
            devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
            preferred_device: "device-a",
            preferred_method: "push",
            result: "auth",
        } as any);

        renderMethod();

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("method-select"));
        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select-a"));

        await waitFor(() => expect(screen.queryByTestId("device-selection")).not.toBeInTheDocument());
        expect(completeSelectionMock).not.toHaveBeenCalled();
    });

    it("succeeds immediately when the Duo policy allows the request", async () => {
        initiateSelectionMock.mockResolvedValue({ result: "allow" } as any);

        const { props } = renderMethod();

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("method-select"));

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Device selection was bypassed by Duo policy" }),
            ),
        );
        expect(screen.getByTestId("success-icon")).toBeInTheDocument();
    });

    it("fails when the Duo policy denies the selection", async () => {
        initiateSelectionMock.mockResolvedValue({ result: "deny" } as any);

        const { props } = renderMethod();

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("method-select"));

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "Device selection was denied by Duo policy" }),
            ),
        );
        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("moves to enroll when the selection process reports enrollment", async () => {
        initiateSelectionMock.mockResolvedValue({
            enroll_url: "https://enroll.example.com",
            result: "enroll",
        } as any);

        const { props } = renderMethod({ duoSelfEnrollment: true });

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("method-select"));

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "No compatible device found" }),
            ),
        );
        expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "NOT_REGISTERED");
    });

    it("reports failures fetching Duo devices", async () => {
        initiateSelectionMock.mockRejectedValue(new Error("boom"));

        const { props } = renderMethod();

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        fireEvent.click(screen.getByTestId("method-select"));

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "There was an issue fetching Duo device(s)" }),
            ),
        );
    });

    it("swallows cancelled Duo device fetches", async () => {
        initiateSelectionMock.mockRejectedValue({ __CANCEL__: true });

        const { props } = renderMethod();

        await waitFor(() => expect(completeSignInMock).toHaveBeenCalled());
        props.onSignInError.mockClear();

        fireEvent.click(screen.getByTestId("method-select"));

        await waitFor(() => expect(initiateSelectionMock).toHaveBeenCalled());
        expect(props.onSignInError).not.toHaveBeenCalled();
    });

    it("returns to the previous state when selection is dismissed", async () => {
        completeSignInMock.mockResolvedValue({
            data: {
                devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
                result: "auth",
            },
        } as any);

        renderMethod();

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-back"));

        await waitFor(() => expect(screen.getByTestId("method-container")).toBeInTheDocument());
    });

    it("persists a newly selected device and signs in again", async () => {
        completeSignInMock.mockResolvedValue({
            data: {
                devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
                result: "auth",
            },
        } as any);

        renderMethod({ registered: false });

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select-b"));

        await waitFor(() =>
            expect(completeSelectionMock).toHaveBeenCalledWith(
                { device: "device-b", method: "sms" },
                expect.anything(),
            ),
        );
    });

    it("notifies the parent when an unregistered user picks a device", async () => {
        completeSignInMock.mockResolvedValue({
            data: {
                devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
                result: "auth",
            },
        } as any);

        const { props } = renderMethod({ registered: false });

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select-b"));

        await waitFor(() => expect(props.onSelectionClick).toHaveBeenCalled());
    });

    it("does not notify the parent when the user is already registered", async () => {
        completeSignInMock.mockResolvedValue({
            data: {
                devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
                result: "auth",
            },
        } as any);

        const { props } = renderMethod({ registered: true });

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select-b"));

        await waitFor(() => expect(completeSelectionMock).toHaveBeenCalled());
        expect(props.onSelectionClick).not.toHaveBeenCalled();
    });

    it("reports failures updating the preferred Duo device", async () => {
        completeSignInMock.mockResolvedValue({
            data: {
                devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
                result: "auth",
            },
        } as any);
        completeSelectionMock.mockRejectedValue(new Error("boom"));

        const { props } = renderMethod();

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select-b"));

        await waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "There was an issue updating preferred Duo device" }),
            ),
        );
    });

    it("swallows cancelled preferred device updates", async () => {
        completeSignInMock.mockResolvedValue({
            data: {
                devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
                result: "auth",
            },
        } as any);
        completeSelectionMock.mockRejectedValue({ __CANCEL__: true });

        const { props } = renderMethod();

        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select-b"));

        await waitFor(() => expect(completeSelectionMock).toHaveBeenCalled());
        expect(props.onSignInError).not.toHaveBeenCalledWith(
            expect.objectContaining({ message: "There was an issue updating preferred Duo device" }),
        );
    });

    it("shows success after selecting a device when already at two factor level", async () => {
        initiateSelectionMock.mockResolvedValue({
            devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
            result: "auth",
        } as any);

        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });

        fireEvent.click(screen.getByTestId("method-select"));
        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-select-b"));

        await waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());
    });
});

describe("self enrollment", () => {
    it("opens the enrollment url in a new tab", async () => {
        const open = vi.spyOn(window, "open").mockImplementation(() => null);
        completeSignInMock.mockResolvedValue({
            data: { enroll_url: "https://enroll.example.com", result: "enroll" },
        } as any);

        renderMethod({ duoSelfEnrollment: true });

        await waitFor(() =>
            expect(screen.getByTestId("method-container")).toHaveAttribute("data-self-enrollment", "true"),
        );

        fireEvent.click(screen.getByTestId("method-register"));

        expect(open).toHaveBeenCalledWith("https://enroll.example.com", "_blank", "noopener,noreferrer");
    });
});

describe("cleanup and edge cases", () => {
    it("clears both pending timers on unmount", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 5 } as any);

        const { unmount } = renderMethod();

        await vi.waitFor(() => expect(completeSignInMock).toHaveBeenCalled());

        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout");

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalled();
    });

    it("clears the success timer on unmount", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ data: { redirect: "https://example.com" } } as any);

        const { props, unmount } = renderMethod();

        await vi.waitFor(() => expect(screen.getByTestId("success-icon")).toBeInTheDocument());

        unmount();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1500);
        });

        expect(props.onSignInSuccess).not.toHaveBeenCalled();
    });

    it("replaces an in flight rate limit timer", async () => {
        vi.useFakeTimers();
        completeSignInMock.mockResolvedValue({ limited: true, retryAfter: 5 } as any);

        const { props } = renderMethod();

        await vi.waitFor(() =>
            expect(props.onSignInError).toHaveBeenCalledWith(
                expect.objectContaining({ message: "You have made too many requests" }),
            ),
        );

        fireEvent.click(screen.getByTestId("method-select"));
        await vi.waitFor(() => expect(initiateSelectionMock).toHaveBeenCalled());

        fireEvent.click(screen.getByRole("button", { name: "Retry" }));

        await vi.waitFor(() => expect(completeSignInMock).toHaveBeenCalledTimes(2));

        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        expect(screen.getByTestId("failure-icon")).toBeInTheDocument();
    });

    it("swallows a cancelled preferred device lookup", async () => {
        getPreferredMock.mockRejectedValue({ __CANCEL__: true });

        renderMethod();

        await waitFor(() => expect(getPreferredMock).toHaveBeenCalled());
        expect(console.debug).not.toHaveBeenCalled();
    });

    it("does nothing when signing in at two factor level", async () => {
        initiateSelectionMock.mockResolvedValue({
            devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
            result: "auth",
        } as any);

        renderMethod({ authenticationLevel: AuthenticationLevel.TwoFactor });

        fireEvent.click(screen.getByTestId("method-select"));
        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-back"));

        await waitFor(() => expect(screen.getByTestId("method-container")).toBeInTheDocument());
        expect(completeSignInMock).not.toHaveBeenCalled();
    });

    it("returns to the pre-selection state after browsing devices from the method container", async () => {
        completeSignInMock.mockResolvedValue(null as any);
        initiateSelectionMock.mockResolvedValue({
            devices: [{ capabilities: ["push"], device: "device-a", display_name: "Phone" }],
            result: "auth",
        } as any);

        renderMethod();

        await waitFor(() => expect(screen.getByTestId("failure-icon")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("method-select"));
        await waitFor(() => expect(screen.getByTestId("device-selection")).toBeInTheDocument());

        fireEvent.click(screen.getByTestId("device-back"));

        await waitFor(() => expect(screen.getByTestId("failure-icon")).toBeInTheDocument());
    });
});
