// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen } from "@testing-library/react";

import { postFirstFactorReauthenticate } from "@services/Password";
import ReauthenticationPasswordForm from "@views/Settings/Common/ReauthenticationPasswordForm";

const mocks = vi.hoisted(() => ({ createErrorNotification: vi.fn() }));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({ createErrorNotification: mocks.createErrorNotification }),
}));

vi.mock("@services/Password", () => ({
    postFirstFactorReauthenticate: vi.fn(),
}));

const postMock = vi.mocked(postFirstFactorReauthenticate);

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
});

afterEach(() => {
    vi.restoreAllMocks();
});

function typePassword(value: string) {
    fireEvent.change(document.getElementById("reauthenticate-password-textfield") as HTMLInputElement, {
        target: { value },
    });
}

it("does not submit an empty password", async () => {
    const onAuthenticationSuccess = vi.fn();
    render(<ReauthenticationPasswordForm onAuthenticationSuccess={onAuthenticationSuccess} />);

    await act(async () => {
        fireEvent.click(document.getElementById("reauthenticate-button") as HTMLButtonElement);
    });

    expect(postMock).not.toHaveBeenCalled();
    expect(onAuthenticationSuccess).not.toHaveBeenCalled();
});

it("reauthenticates with the password", async () => {
    postMock.mockResolvedValue({} as any);
    const onAuthenticationSuccess = vi.fn();
    render(<ReauthenticationPasswordForm onAuthenticationSuccess={onAuthenticationSuccess} />);

    typePassword("secret");

    await act(async () => {
        fireEvent.click(document.getElementById("reauthenticate-button") as HTMLButtonElement);
    });

    expect(postMock).toHaveBeenCalledWith("secret");
    expect(onAuthenticationSuccess).toHaveBeenCalledOnce();
});

it("notifies and clears the password on failure", async () => {
    postMock.mockRejectedValue(new Error("bad"));
    const onAuthenticationSuccess = vi.fn();
    render(<ReauthenticationPasswordForm onAuthenticationSuccess={onAuthenticationSuccess} />);

    typePassword("wrong");

    await act(async () => {
        fireEvent.click(document.getElementById("reauthenticate-button") as HTMLButtonElement);
    });

    expect(mocks.createErrorNotification).toHaveBeenCalledWith("Incorrect password");
    expect(onAuthenticationSuccess).not.toHaveBeenCalled();
    expect(screen.getByDisplayValue("")).toBeInTheDocument();
});

it("submits on enter", async () => {
    postMock.mockResolvedValue({} as any);
    render(<ReauthenticationPasswordForm onAuthenticationSuccess={vi.fn()} />);

    typePassword("secret");

    await act(async () => {
        fireEvent.keyDown(document.getElementById("reauthenticate-password-textfield") as HTMLInputElement, {
            key: "Enter",
        });
    });

    expect(postMock).toHaveBeenCalledWith("secret");
});
