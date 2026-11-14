// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import {
    cancelAppAuthenticatorRegistration,
    getAppAuthenticatorRegistration,
    startAppAuthenticatorRegistration,
} from "@services/AppAuthenticator";
import AppAuthenticatorRegisterDialog from "@views/Settings/TwoFactorAuthentication/AppAuthenticatorRegisterDialog";

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

vi.mock("@services/AppAuthenticator", () => ({
    cancelAppAuthenticatorRegistration: vi.fn(),
    getAppAuthenticatorRegistration: vi.fn(),
    startAppAuthenticatorRegistration: vi.fn(),
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

vi.mock("qrcode.react", () => ({
    QRCodeSVG: (props: any) => <div data-testid="qr-code" data-value={props.value} />,
}));

const startMock = vi.mocked(startAppAuthenticatorRegistration);
const getMock = vi.mocked(getAppAuthenticatorRegistration);
const cancelMock = vi.mocked(cancelAppAuthenticatorRegistration);

const qr = "authelia-app://login.example.com/api/register/app/abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ";

function expiry(seconds = 300) {
    return Math.floor(Date.now() / 1000) + seconds;
}

function renderDialog(open = true) {
    const setClosed = vi.fn();

    return { ...render(<AppAuthenticatorRegisterDialog open={open} setClosed={setClosed} />), setClosed };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    vi.spyOn(console, "error").mockImplementation(() => {});
    startMock.mockResolvedValue({ expires_at: expiry(), qr });
    getMock.mockResolvedValue({ expires_at: expiry(), state: "waiting" });
    cancelMock.mockResolvedValue(undefined);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

it("does nothing while closed", async () => {
    renderDialog(false);

    await vi.advanceTimersByTimeAsync(5000);

    expect(startMock).not.toHaveBeenCalled();
});

it("shows the qr code, its text and a countdown", async () => {
    renderDialog();

    await vi.waitFor(() => expect(screen.getByTestId("qr-code")).toHaveAttribute("data-value", qr));
    expect(screen.getByDisplayValue(qr)).toBeInTheDocument();
    expect(screen.getByTestId("copy-button")).toHaveAttribute("data-value", qr);
    expect(document.getElementById("app-authenticator-register-countdown")).toBeInTheDocument();
});

it("walks from waiting to the number to registered", async () => {
    getMock
        .mockResolvedValueOnce({ expires_at: expiry(), state: "waiting" })
        .mockResolvedValueOnce({
            device: { model: "iPhone15,2", name: "Jane's iPhone", platform: "iOS" },
            expires_at: expiry(),
            nonce_length: 2,
            number: "07",
            state: "confirm_number",
        })
        .mockResolvedValueOnce({ id: 12, state: "registered" });

    const { setClosed } = renderDialog();

    await vi.waitFor(() => expect(screen.getByTestId("qr-code")).toBeInTheDocument());
    expect(getMock).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(2000);
    expect(getMock).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId("qr-code")).toBeInTheDocument();

    await vi.advanceTimersByTimeAsync(2000);
    expect(screen.getByText("07")).toBeInTheDocument();
    expect(screen.getByText("Jane's iPhone, iOS")).toBeInTheDocument();
    expect(screen.getByText("Enter this number in the Authelia app.")).toBeInTheDocument();
    expect(screen.queryByTestId("qr-code")).not.toBeInTheDocument();

    await vi.advanceTimersByTimeAsync(2000);
    expect(screen.getByTestId("success-icon")).toBeInTheDocument();
    expect(mocks.createSuccessNotification).toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1500);
    expect(setClosed).toHaveBeenCalledTimes(1);
    expect(cancelMock).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(10000);
    expect(getMock).toHaveBeenCalledTimes(3);
});

it.each(["failed", "expired"] as const)("shows the failure for %s and starts again", async (state) => {
    getMock.mockResolvedValueOnce({ state });

    renderDialog();

    await vi.waitFor(() => expect(screen.getByTestId("qr-code")).toBeInTheDocument());
    await vi.advanceTimersByTimeAsync(2000);

    expect(screen.getByText("Registration failed. Start again.")).toBeInTheDocument();
    expect(screen.queryByTestId("qr-code")).not.toBeInTheDocument();

    await vi.advanceTimersByTimeAsync(10000);
    expect(getMock).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByText("Start again"));

    await vi.waitFor(() => expect(startMock).toHaveBeenCalledTimes(2));
    await vi.waitFor(() => expect(screen.getByTestId("qr-code")).toBeInTheDocument());
});

it("shows the failure when the registration can no longer be found", async () => {
    getMock.mockRejectedValueOnce(new Error("not found"));

    renderDialog();

    await vi.waitFor(() => expect(screen.getByTestId("qr-code")).toBeInTheDocument());
    await vi.advanceTimersByTimeAsync(2000);

    expect(screen.getByText("Registration failed. Start again.")).toBeInTheDocument();
});

it("shows the failure when the registration cannot be started", async () => {
    startMock.mockRejectedValueOnce(new Error("forbidden"));

    renderDialog();

    await vi.waitFor(() => expect(screen.getByText("Registration failed. Start again.")).toBeInTheDocument());
    expect(mocks.createErrorNotification).toHaveBeenCalled();
    expect(getMock).not.toHaveBeenCalled();
});

it("cancels the registration and stops polling when closed", async () => {
    const { setClosed } = renderDialog();

    await vi.waitFor(() => expect(screen.getByTestId("qr-code")).toBeInTheDocument());

    fireEvent.click(screen.getByText("Cancel"));

    await vi.waitFor(() => expect(cancelMock).toHaveBeenCalledTimes(1));
    expect(setClosed).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(10000);
    expect(getMock).not.toHaveBeenCalled();
});
