// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, render, screen } from "@testing-library/react";
import axios from "axios";

import App from "@root/App";
import "@i18n/index";

vi.mock("@views/ConsentPortal/ConsentPortal", () => ({
    default: () => <div data-testid="consent-portal" />,
}));

vi.mock("@views/LoginPortal/SignOut/SignOut", () => ({
    default: () => <div data-testid="sign-out" />,
}));

vi.mock("@views/ResetPassword/ResetPasswordStep1", () => ({
    default: () => <div data-testid="reset-password-step1" />,
}));

vi.mock("@views/ResetPassword/ResetPasswordStep2", () => ({
    default: () => <div data-testid="reset-password-step2" />,
}));

vi.mock("@views/Settings/SettingsRouter", () => ({
    default: () => <div data-testid="settings-router" />,
}));

vi.mock("@views/Revoke/RevokeOneTimeCodeView", () => ({
    default: () => <div data-testid="revoke-one-time-code" />,
}));

vi.mock("@views/Revoke/RevokeResetPasswordTokenView", () => ({
    default: () => <div data-testid="revoke-reset-password" />,
}));

vi.mock("@views/LoginPortal/LoginPortal", () => ({
    default: (props: any) => (
        <div
            data-testid="login-portal"
            data-duo-self-enrollment={String(props.duoSelfEnrollment)}
            data-passkey-login={String(props.passkeyLogin)}
            data-remember-me={String(props.rememberMe)}
            data-reset-password={String(props.resetPassword)}
            data-reset-password-custom-url={props.resetPasswordCustomURL}
            data-registration-url={props.registrationURL}
        />
    ),
}));

async function renderAt(path: string) {
    window.history.pushState({}, "", path);

    await act(async () => {
        render(<App />);
    });
}

beforeEach(() => {
    vi.spyOn(axios, "get").mockResolvedValue({ data: { data: {}, status: "OK" }, status: 200 });
});

afterEach(() => {
    vi.restoreAllMocks();
    window.history.pushState({}, "", "/");
});

it("renders without crashing", async () => {
    await renderAt("/");
});

it("renders the login portal at the index route", async () => {
    await renderAt("/");
    expect(screen.getByTestId("login-portal")).toBeInTheDocument();
});

it("passes the document configuration to the login portal", async () => {
    await renderAt("/");

    const portal = screen.getByTestId("login-portal");
    expect(portal).toHaveAttribute("data-duo-self-enrollment", "true");
    expect(portal).toHaveAttribute("data-passkey-login", "true");
    expect(portal).toHaveAttribute("data-remember-me", "true");
    expect(portal).toHaveAttribute("data-reset-password", "true");
    expect(portal).toHaveAttribute("data-reset-password-custom-url", "");
});

it.each([
    ["/reset-password/step1", "reset-password-step1"],
    ["/reset-password/step2", "reset-password-step2"],
    ["/logout", "sign-out"],
    ["/revoke/one-time-code", "revoke-one-time-code"],
    ["/revoke/reset-password", "revoke-reset-password"],
    ["/settings/two-factor-authentication", "settings-router"],
    ["/consent/openid", "consent-portal"],
])("renders %s", async (path, testId) => {
    await renderAt(path);
    expect(screen.getByTestId(testId)).toBeInTheDocument();
});

it("falls through to the login portal for unknown paths", async () => {
    await renderAt("/some/unknown/path");
    expect(screen.getByTestId("login-portal")).toBeInTheDocument();
});
