// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";

import PasswordChangeRequiredForm from "@views/LoginPortal/PasswordChangeRequired/PasswordChangeRequiredForm";

const createErrorNotification = vi.fn();
const postPasswordChange = vi.fn();

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

beforeEach(() => {
    createErrorNotification.mockReset();
    postPasswordChange.mockReset();
    postPasswordChange.mockResolvedValue(undefined);
});

function renderForm(onPasswordChanged = vi.fn()) {
    render(<PasswordChangeRequiredForm username="john" onPasswordChanged={onPasswordChanged} />);

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
