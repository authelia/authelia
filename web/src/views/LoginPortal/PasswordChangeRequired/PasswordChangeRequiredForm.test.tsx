// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import PasswordChangeRequiredForm from "@views/LoginPortal/PasswordChangeRequired/PasswordChangeRequiredForm";

const createErrorNotification = vi.fn();
const postPasswordChange = vi.fn();
const signOut = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({ createErrorNotification, createSuccessNotification: vi.fn() }),
}));

vi.mock("@layouts/LoginLayout", () => ({
    default: (props: any) => (
        <div data-testid="login-layout" data-title={props.title}>
            {props.children}
        </div>
    ),
}));

vi.mock("@services/ChangePassword", () => ({
    postPasswordChange: (...args: unknown[]) => postPasswordChange(...args),
}));

vi.mock("@services/SignOut", () => ({
    signOut: (...args: unknown[]) => signOut(...args),
}));

beforeEach(() => {
    createErrorNotification.mockReset();
    postPasswordChange.mockReset();
    postPasswordChange.mockResolvedValue(undefined);
    signOut.mockReset();
    signOut.mockResolvedValue(undefined);
});

function renderForm(onPasswordChanged = vi.fn(), onSignedOut = vi.fn()) {
    render(
        <PasswordChangeRequiredForm username="john" onPasswordChanged={onPasswordChanged} onSignedOut={onSignedOut} />,
    );

    return onPasswordChanged;
}

function type(id: string, value: string) {
    fireEvent.change(document.getElementById(id) as HTMLElement, { target: { value } });
}

it("tells the user why they are here", () => {
    renderForm();

    expect(screen.getByText("Your password must be changed before you can continue")).toBeInTheDocument();
});

it("changes the password and reports back", async () => {
    const onPasswordChanged = renderForm();

    type("old-password", "temporary");
    type("new-password", "not-a-secret");
    type("repeat-new-password", "not-a-secret");

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() => expect(postPasswordChange).toHaveBeenCalledWith("john", "temporary", "not-a-secret"));
    await waitFor(() => expect(onPasswordChanged).toHaveBeenCalled());
});

it("refuses to submit when the new passwords differ", async () => {
    const onPasswordChanged = renderForm();

    type("old-password", "temporary");
    type("new-password", "not-a-secret");
    type("repeat-new-password", "a-different-one");

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() => expect(createErrorNotification).toHaveBeenCalled());
    expect(postPasswordChange).not.toHaveBeenCalled();
    expect(onPasswordChanged).not.toHaveBeenCalled();
});

it("refuses to submit without the current password", async () => {
    renderForm();

    type("new-password", "not-a-secret");
    type("repeat-new-password", "not-a-secret");

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() => expect(postPasswordChange).not.toHaveBeenCalled());
});

it("keeps the user on the form when the change fails", async () => {
    postPasswordChange.mockRejectedValue(new Error("incorrect password"));

    const onPasswordChanged = renderForm();

    type("old-password", "wrong");
    type("new-password", "not-a-secret");
    type("repeat-new-password", "not-a-secret");

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() => expect(createErrorNotification).toHaveBeenCalled());
    expect(onPasswordChanged).not.toHaveBeenCalled();
});

function axiosError(status: number) {
    return Object.assign(new Error("request failed"), { isAxiosError: true, response: { status } });
}

it("tells the user when the current password is incorrect", async () => {
    postPasswordChange.mockRejectedValue(axiosError(401));

    renderForm();

    type("old-password", "wrong");
    type("new-password", "not-a-secret");
    type("repeat-new-password", "not-a-secret");

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() => expect(createErrorNotification).toHaveBeenCalledWith("Incorrect password"));
});

it("tells the user when the new password does not meet the policy", async () => {
    postPasswordChange.mockRejectedValue(axiosError(400));

    renderForm();

    type("old-password", "temporary");
    type("new-password", "weak");
    type("repeat-new-password", "weak");

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() =>
        expect(createErrorNotification).toHaveBeenCalledWith(
            "Your supplied password does not meet the password policy requirements",
        ),
    );
});

it("signs the user out", async () => {
    const onSignedOut = vi.fn();

    renderForm(vi.fn(), onSignedOut);

    fireEvent.click(document.getElementById("password-change-sign-out-button") as HTMLElement);

    await waitFor(() => expect(signOut).toHaveBeenCalledWith(undefined));
    await waitFor(() => expect(onSignedOut).toHaveBeenCalled());
});

it("reports a failed sign out", async () => {
    signOut.mockRejectedValue(new Error("failed"));

    const onSignedOut = vi.fn();

    renderForm(vi.fn(), onSignedOut);

    fireEvent.click(document.getElementById("password-change-sign-out-button") as HTMLElement);

    await waitFor(() => expect(createErrorNotification).toHaveBeenCalledWith("There was an issue signing out"));
    expect(onSignedOut).toHaveBeenCalled();
});

it("offers to retry only the state refresh after the password has changed", async () => {
    const onPasswordChanged = renderForm();

    type("old-password", "temporary");
    type("new-password", "not-a-secret");
    type("repeat-new-password", "not-a-secret");

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() => expect(onPasswordChanged).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(screen.getByText("Retry")).toBeInTheDocument());

    expect(document.getElementById("password-change-button")).not.toBeDisabled();
    expect(document.getElementById("password-change-sign-out-button")).not.toBeDisabled();

    fireEvent.click(document.getElementById("password-change-button") as HTMLElement);

    await waitFor(() => expect(onPasswordChanged).toHaveBeenCalledTimes(2));
    expect(postPasswordChange).toHaveBeenCalledTimes(1);
});
