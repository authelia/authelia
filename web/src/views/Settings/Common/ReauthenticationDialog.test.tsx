// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen } from "@testing-library/react";

import ReauthenticationDialog from "@views/Settings/Common/ReauthenticationDialog";

const mocks = vi.hoisted(() => ({ supportsWebAuthn: true }));

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

vi.mock("@views/Settings/Common/ReauthenticationPasswordForm", () => ({
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

function elevation(methods: string[], required = true) {
    return { reauthentication_methods: methods, require_reauthentication: required } as any;
}

const info = { has_duo: false, has_totp: true, has_webauthn: true } as any;

function renderDialog(props: Partial<{ elevation: any; info: any; opening: boolean }> = {}) {
    const handleClosed = vi.fn();
    const handleOpened = vi.fn();
    const merged = { elevation: elevation(["password"]), info, opening: true, ...props, handleClosed, handleOpened };

    return { ...render(<ReauthenticationDialog {...merged} />), handleClosed, handleOpened };
}

beforeEach(() => {
    mocks.supportsWebAuthn = true;
});

afterEach(() => {
    vi.useRealTimers();
});

it("closes immediately when reauthentication is not required", () => {
    const { handleClosed, handleOpened } = renderDialog({ elevation: elevation([], false) });

    expect(handleClosed).toHaveBeenCalledWith(true, false);
    expect(handleOpened).not.toHaveBeenCalled();
});

it("waits for the elevation to load", () => {
    const { handleClosed, handleOpened } = renderDialog({ elevation: undefined });

    expect(handleClosed).not.toHaveBeenCalled();
    expect(handleOpened).not.toHaveBeenCalled();
});

it("does nothing while not opening", () => {
    const { handleClosed, handleOpened } = renderDialog({ opening: false });

    expect(handleClosed).not.toHaveBeenCalled();
    expect(handleOpened).not.toHaveBeenCalled();
});

it("skips the picker when only the password is allowed", () => {
    const { handleOpened } = renderDialog({ elevation: elevation(["password"]) });

    expect(handleOpened).toHaveBeenCalledOnce();
    expect(screen.getByTestId("password-form")).toBeInTheDocument();
});

it("offers the password and every enrolled second factor method for any", () => {
    renderDialog({ elevation: elevation(["password", "second_factor"]) });

    expect(screen.getByText("Password")).toBeInTheDocument();
    expect(screen.getByText("One-Time Password")).toBeInTheDocument();
    expect(screen.getByText("WebAuthn")).toBeInTheDocument();
    expect(screen.queryByText("Mobile Push")).not.toBeInTheDocument();
});

it("never offers the password when only second factor is allowed", () => {
    renderDialog({ elevation: elevation(["second_factor"]) });

    expect(screen.queryByText("Password")).not.toBeInTheDocument();
    expect(screen.getByText("One-Time Password")).toBeInTheDocument();
    expect(screen.getByText("WebAuthn")).toBeInTheDocument();
});

it("hides WebAuthn when the browser does not support it", async () => {
    mocks.supportsWebAuthn = false;

    renderDialog({ elevation: elevation(["second_factor"]) });

    expect(screen.queryByText("WebAuthn")).not.toBeInTheDocument();
    // The second factor method components are lazy loaded.
    expect(await screen.findByTestId("method-otp")).toBeInTheDocument();
});

it("shows the selected method", async () => {
    renderDialog({ elevation: elevation(["password", "second_factor"]) });

    fireEvent.click(screen.getByText("One-Time Password"));

    expect(await screen.findByTestId("method-otp")).toBeInTheDocument();
});

it("explains when no method is available", () => {
    renderDialog({
        elevation: elevation(["second_factor"]),
        info: { has_duo: false, has_totp: false, has_webauthn: false },
    });

    expect(screen.getByText("There are no methods available to authenticate again")).toBeInTheDocument();
});

it("reports a change after a successful reauthentication", () => {
    vi.useFakeTimers();

    const { handleClosed } = renderDialog({ elevation: elevation(["password"]) });

    fireEvent.click(screen.getByTestId("password-success"));

    expect(screen.getByTestId("success-icon")).toBeInTheDocument();

    act(() => {
        vi.advanceTimersByTime(1500);
    });

    expect(handleClosed).toHaveBeenCalledWith(true, true);
});

it("reports a cancellation", () => {
    const { handleClosed } = renderDialog({ elevation: elevation(["password", "second_factor"]) });

    fireEvent.click(screen.getByText("Cancel"));

    expect(handleClosed).toHaveBeenCalledWith(false, false);
});
