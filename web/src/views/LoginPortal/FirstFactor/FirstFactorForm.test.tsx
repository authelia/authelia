// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { postFirstFactor } from "@services/Password";
import FirstFactorForm from "@views/LoginPortal/FirstFactor/FirstFactorForm";

const mocks = vi.hoisted(() => ({
    capsLockModified: null as boolean | null,
    channelListeners: [] as ((authenticated: boolean) => void)[],
    createErrorNotification: vi.fn(),
    navigate: vi.fn(),
    postMessage: vi.fn(),
    queryParams: {} as Record<string, null | string>,
    userCode: null as null | string,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-router", async () => {
    const actual = await vi.importActual("react-router");
    return { ...actual, useNavigate: () => mocks.navigate };
});

vi.mock("broadcast-channel", () => {
    class MockBroadcastChannel {
        addEventListener = vi.fn((_name: string, handler: (authenticated: boolean) => void) => {
            mocks.channelListeners.push(handler);
        });
        removeEventListener = vi.fn((_name: string, handler: (authenticated: boolean) => void) => {
            mocks.channelListeners = mocks.channelListeners.filter((h) => h !== handler);
        });
        postMessage = mocks.postMessage;
    }
    return { BroadcastChannel: MockBroadcastChannel };
});

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: (name: string) => mocks.queryParams[name] ?? null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: "flow", id: "flow-id", subflow: "subflow" }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => mocks.userCode,
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@layouts/LoginLayout", () => ({
    default: (props: any) => <div data-testid="login-layout">{props.children}</div>,
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: () => mocks.capsLockModified,
}));

vi.mock("@services/Password", () => ({
    postFirstFactor: vi.fn(),
}));

vi.mock("@views/LoginPortal/FirstFactor/PasskeyForm", () => ({
    default: (props: any) => (
        <div
            data-testid="passkey-form"
            data-disabled={String(props.disabled)}
            data-remember-me={String(props.rememberMe)}
        >
            <button data-testid="passkey-start" onClick={() => props.onAuthenticationStart()} />
            <button data-testid="passkey-stop" onClick={() => props.onAuthenticationStop()} />
            <button
                data-testid="passkey-success"
                onClick={() => props.onAuthenticationSuccess("https://example.com")}
            />
            <button
                data-testid="passkey-error"
                onClick={() => props.onAuthenticationError(new Error("passkey failed"))}
            />
        </div>
    ),
}));

const postFirstFactorMock = vi.mocked(postFirstFactor);

const defaultProps = {
    disabled: false,
    onAuthenticationStart: vi.fn(),
    onAuthenticationStop: vi.fn(),
    onAuthenticationSuccess: vi.fn(),
    onChannelStateChange: vi.fn(),
    passkeyLogin: false,
    rememberMe: false,
    resetPassword: false,
    resetPasswordCustomURL: "",
};

function renderForm(props: Partial<typeof defaultProps> = {}) {
    const merged = {
        ...defaultProps,
        onAuthenticationStart: vi.fn(),
        onAuthenticationStop: vi.fn(),
        onAuthenticationSuccess: vi.fn(),
        onChannelStateChange: vi.fn(),
        ...props,
    };

    return { ...render(<FirstFactorForm {...merged} />), props: merged };
}

function getUsername() {
    return screen.getByLabelText(/Username/) as HTMLInputElement;
}

function getPassword() {
    return screen.getByLabelText(/Password/) as HTMLInputElement;
}

function fillCredentials(username = "john", password = "secret") {
    fireEvent.change(getUsername(), { target: { value: username } });
    fireEvent.change(getPassword(), { target: { value: password } });
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.capsLockModified = null;
    mocks.channelListeners = [];
    mocks.queryParams = {};
    mocks.userCode = null;
    postFirstFactorMock.mockReset();
});

describe("rendering", () => {
    it("renders login form with username, password, and sign in button", () => {
        renderForm();
        expect(screen.getByText(/Username/)).toBeInTheDocument();
        expect(screen.getByText(/Password/)).toBeInTheDocument();
        expect(screen.getByText("Sign in")).toBeInTheDocument();
    });

    it("renders remember me checkbox when enabled", () => {
        renderForm({ rememberMe: true });
        expect(screen.getByText("Remember me")).toBeInTheDocument();
    });

    it("does not render remember me checkbox when disabled", () => {
        renderForm({ rememberMe: false });
        expect(screen.queryByText("Remember me")).not.toBeInTheDocument();
    });

    it("renders passkey form when passkey login is enabled", () => {
        renderForm({ passkeyLogin: true });
        expect(screen.getByTestId("passkey-form")).toBeInTheDocument();
    });

    it("renders reset password link when enabled", () => {
        renderForm({ resetPassword: true });
        expect(screen.getByText("Reset password?")).toBeInTheDocument();
    });

    it("does not render reset password link when disabled", () => {
        renderForm({ resetPassword: false });
        expect(screen.queryByText("Reset password?")).not.toBeInTheDocument();
    });

    it("disables the inputs and the sign in button when disabled", () => {
        renderForm({ disabled: true });
        expect(getUsername()).toBeDisabled();
        expect(getPassword()).toBeDisabled();
        expect(screen.getByRole("button", { name: /Sign in/ })).toBeDisabled();
    });

    it("focuses the username field on mount", async () => {
        renderForm();
        await waitFor(() => expect(getUsername()).toHaveFocus());
    });
});

describe("validation", () => {
    it("flags both fields when signing in with an empty form", async () => {
        const { props } = renderForm();

        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() => expect(getUsername()).toHaveAttribute("aria-invalid", "true"));
        expect(getPassword()).toHaveAttribute("aria-invalid", "true");
        expect(postFirstFactorMock).not.toHaveBeenCalled();
        expect(props.onAuthenticationStart).not.toHaveBeenCalled();
    });

    it("flags only the password when the username is filled", async () => {
        renderForm();

        fireEvent.change(getUsername(), { target: { value: "john" } });
        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() => expect(getPassword()).toHaveAttribute("aria-invalid", "true"));
        expect(getUsername()).not.toHaveAttribute("aria-invalid", "true");
    });

    it("flags only the username when the password is filled", async () => {
        renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() => expect(getUsername()).toHaveAttribute("aria-invalid", "true"));
        expect(getPassword()).not.toHaveAttribute("aria-invalid", "true");
    });

    it("clears the username error when the field is focused", async () => {
        renderForm();

        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));
        await waitFor(() => expect(getUsername()).toHaveAttribute("aria-invalid", "true"));

        fireEvent.focus(getUsername());
        await waitFor(() => expect(getUsername()).not.toHaveAttribute("aria-invalid", "true"));
    });

    it("clears the password error when the field is focused", async () => {
        renderForm();

        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));
        await waitFor(() => expect(getPassword()).toHaveAttribute("aria-invalid", "true"));

        fireEvent.focus(getPassword());
        await waitFor(() => expect(getPassword()).not.toHaveAttribute("aria-invalid", "true"));
    });
});

describe("sign in", () => {
    it("posts the credentials and reports success", async () => {
        postFirstFactorMock.mockResolvedValue({ redirect: "https://example.com/after" } as any);

        const { props } = renderForm();

        fillCredentials();
        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com/after"));
        expect(props.onAuthenticationStart).toHaveBeenCalled();
        expect(mocks.postMessage).toHaveBeenCalledWith(true);
        expect(postFirstFactorMock).toHaveBeenCalledWith(
            "john",
            "secret",
            false,
            null,
            null,
            "flow-id",
            "flow",
            "subflow",
            null,
        );
    });

    it("reports success with an undefined redirect when the response is empty", async () => {
        postFirstFactorMock.mockResolvedValue(undefined as any);

        const { props } = renderForm();

        fillCredentials();
        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith(undefined));
    });

    it("forwards the redirection URL, request method and user code", async () => {
        mocks.queryParams = { rd: "https://app.example.com", rm: "GET" };
        mocks.userCode = "ABCD-EFGH";
        postFirstFactorMock.mockResolvedValue(undefined as any);

        renderForm();

        fillCredentials();
        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() =>
            expect(postFirstFactorMock).toHaveBeenCalledWith(
                "john",
                "secret",
                false,
                "https://app.example.com",
                "GET",
                "flow-id",
                "flow",
                "subflow",
                "ABCD-EFGH",
            ),
        );
    });

    it("passes the remember me selection through", async () => {
        postFirstFactorMock.mockResolvedValue(undefined as any);

        renderForm({ rememberMe: true });

        fireEvent.click(screen.getByRole("checkbox"));
        fillCredentials();
        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() => expect(postFirstFactorMock).toHaveBeenCalled());
        expect(postFirstFactorMock.mock.calls[0][2]).toBe(true);
    });

    it("notifies, clears the password and refocuses it on failure", async () => {
        const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
        postFirstFactorMock.mockRejectedValue(new Error("bad credentials"));

        const { props } = renderForm();

        await waitFor(() => expect(getUsername()).toHaveFocus());

        fillCredentials();
        fireEvent.click(screen.getByRole("button", { name: /Sign in/ }));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("Incorrect username or password"),
        );
        expect(props.onAuthenticationStop).toHaveBeenCalled();
        expect(props.onAuthenticationSuccess).not.toHaveBeenCalled();
        await waitFor(() => expect(getPassword()).toHaveValue(""));
        expect(getPassword()).toHaveFocus();

        consoleError.mockRestore();
    });

    it("ignores repeated submissions while a sign in is in flight", async () => {
        let resolve: (value: unknown) => void = () => {};
        postFirstFactorMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderForm();

        fillCredentials();
        const button = screen.getByRole("button", { name: /Sign in/ });
        fireEvent.click(button);

        await waitFor(() => expect(button).toBeDisabled());

        fireEvent.keyDown(getPassword(), { key: "Enter" });

        expect(postFirstFactorMock).toHaveBeenCalledTimes(1);

        resolve(undefined);
        await waitFor(() => expect(button).not.toBeDisabled());
    });
});

describe("keyboard handling", () => {
    it("flags the username when Enter is pressed on an empty username", async () => {
        renderForm();

        fireEvent.keyDown(getUsername(), { key: "Enter" });

        await waitFor(() => expect(getUsername()).toHaveAttribute("aria-invalid", "true"));
        expect(postFirstFactorMock).not.toHaveBeenCalled();
    });

    it("moves focus to the password when Enter is pressed with only a username", async () => {
        renderForm();

        fireEvent.change(getUsername(), { target: { value: "john" } });
        fireEvent.keyDown(getUsername(), { key: "Enter" });

        await waitFor(() => expect(getPassword()).toHaveFocus());
        expect(postFirstFactorMock).not.toHaveBeenCalled();
    });

    it("signs in when Enter is pressed on the username with both fields filled", async () => {
        postFirstFactorMock.mockResolvedValue(undefined as any);

        renderForm();

        fillCredentials();
        fireEvent.keyDown(getUsername(), { key: "Enter" });

        await waitFor(() => expect(postFirstFactorMock).toHaveBeenCalled());
    });

    it("ignores non-Enter keys on the username", () => {
        renderForm();

        fireEvent.change(getUsername(), { target: { value: "john" } });
        fireEvent.keyDown(getUsername(), { key: "a" });

        expect(postFirstFactorMock).not.toHaveBeenCalled();
    });

    it("signs in when Enter is pressed on the password", async () => {
        postFirstFactorMock.mockResolvedValue(undefined as any);

        renderForm();

        fillCredentials();
        fireEvent.keyDown(getPassword(), { key: "Enter" });

        await waitFor(() => expect(postFirstFactorMock).toHaveBeenCalled());
    });

    it("refocuses the username when Enter is pressed on the password without a username", async () => {
        renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.keyDown(getPassword(), { key: "Enter" });

        await waitFor(() => expect(getUsername()).toHaveFocus());
        expect(postFirstFactorMock).not.toHaveBeenCalled();
    });

    it("keeps focus on the password when Enter is pressed with an empty password", async () => {
        renderForm();

        fireEvent.change(getUsername(), { target: { value: "john" } });
        fireEvent.keyDown(getPassword(), { key: "Enter" });

        await waitFor(() => expect(getPassword()).toHaveFocus());
        expect(postFirstFactorMock).not.toHaveBeenCalled();
    });

    it("ignores non-Enter keys on the password", () => {
        renderForm();

        fillCredentials();
        fireEvent.keyDown(getPassword(), { key: "Shift" });

        expect(postFirstFactorMock).not.toHaveBeenCalled();
    });

    it("signs in when Enter is pressed on the remember me checkbox", async () => {
        postFirstFactorMock.mockResolvedValue(undefined as any);

        renderForm({ rememberMe: true });

        fillCredentials();
        fireEvent.keyDown(screen.getByRole("checkbox"), { key: "Enter" });

        await waitFor(() => expect(postFirstFactorMock).toHaveBeenCalled());
    });

    it("refocuses the username when Enter is pressed on remember me without a username", async () => {
        renderForm({ rememberMe: true });

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.keyDown(screen.getByRole("checkbox"), { key: "Enter" });

        await waitFor(() => expect(getUsername()).toHaveFocus());
    });

    it("focuses the password when Enter is pressed on remember me without a password", async () => {
        renderForm({ rememberMe: true });

        fireEvent.change(getUsername(), { target: { value: "john" } });
        fireEvent.keyDown(screen.getByRole("checkbox"), { key: "Enter" });

        await waitFor(() => expect(getPassword()).toHaveFocus());
    });

    it("ignores non-Enter keys on the remember me checkbox", () => {
        renderForm({ rememberMe: true });

        fillCredentials();
        fireEvent.keyDown(screen.getByRole("checkbox"), { key: "a" });

        expect(postFirstFactorMock).not.toHaveBeenCalled();
    });
});

describe("caps lock detection", () => {
    it("warns when the password was entered with caps lock", async () => {
        mocks.capsLockModified = true;

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "SECRET" } });
        fireEvent.keyUp(getPassword(), { key: "T" });

        expect(await screen.findByText("The password was entered with Caps Lock")).toBeInTheDocument();
    });

    it("warns when the password was only partially entered with caps lock", async () => {
        mocks.capsLockModified = true;

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "SECRET" } });
        fireEvent.keyUp(getPassword(), { key: "T" });
        await screen.findByText("The password was entered with Caps Lock");

        mocks.capsLockModified = false;
        fireEvent.keyUp(getPassword(), { key: "t" });

        expect(await screen.findByText("The password was partially entered with Caps Lock")).toBeInTheDocument();
    });

    it("does not warn when caps lock state cannot be determined", () => {
        mocks.capsLockModified = null;

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "SECRET" } });
        fireEvent.keyUp(getPassword(), { key: "T" });

        expect(screen.queryByText("The password was entered with Caps Lock")).not.toBeInTheDocument();
    });

    it("resets the warning once the password is cleared", async () => {
        mocks.capsLockModified = true;

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "SECRET" } });
        fireEvent.keyUp(getPassword(), { key: "T" });
        await screen.findByText("The password was entered with Caps Lock");

        fireEvent.change(getPassword(), { target: { value: "" } });
        fireEvent.keyUp(getPassword(), { key: "Backspace" });

        await waitFor(() =>
            expect(screen.queryByText("The password was entered with Caps Lock")).not.toBeInTheDocument(),
        );
    });

    it("re-evaluates caps lock on a single character password", async () => {
        mocks.capsLockModified = true;

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "S" } });
        fireEvent.keyUp(getPassword(), { key: "S" });

        expect(await screen.findByText("The password was entered with Caps Lock")).toBeInTheDocument();
    });
});

describe("password visibility toggle", () => {
    it("reveals the password on click and hides it again on a second click", async () => {
        renderForm();

        const toggle = screen.getByLabelText("Toggle password visibility");
        expect(getPassword()).toHaveAttribute("type", "password");

        fireEvent.click(toggle);
        await waitFor(() => expect(getPassword()).toHaveAttribute("type", "text"));
        expect(toggle).toHaveAttribute("aria-pressed", "true");

        fireEvent.click(toggle);
        await waitFor(() => expect(getPassword()).toHaveAttribute("type", "password"));
        expect(toggle).toHaveAttribute("aria-pressed", "false");
    });
});

describe("reset password", () => {
    it("navigates to the reset password route", () => {
        renderForm({ resetPassword: true });

        fireEvent.click(screen.getByText("Reset password?"));

        expect(mocks.navigate).toHaveBeenCalledWith("/reset-password/step1");
    });

    it("opens the custom reset password URL when configured", () => {
        const open = vi.spyOn(window, "open").mockImplementation(() => null);

        renderForm({ resetPassword: true, resetPasswordCustomURL: "https://reset.example.com" });

        fireEvent.click(screen.getByText("Reset password?"));

        expect(open).toHaveBeenCalledWith("https://reset.example.com");
        expect(mocks.navigate).not.toHaveBeenCalled();

        open.mockRestore();
    });
});

describe("remember me", () => {
    it("toggles the checkbox on and off", async () => {
        renderForm({ rememberMe: true });

        const checkbox = screen.getByRole("checkbox");
        expect(checkbox).toHaveAttribute("aria-checked", "false");

        fireEvent.click(checkbox);
        await waitFor(() => expect(checkbox).toHaveAttribute("aria-checked", "true"));

        fireEvent.click(checkbox);
        await waitFor(() => expect(checkbox).toHaveAttribute("aria-checked", "false"));
    });
});

describe("login broadcast channel", () => {
    it("notifies the parent when another tab authenticates", () => {
        const { props } = renderForm();

        mocks.channelListeners.forEach((handler) => handler(true));

        expect(props.onChannelStateChange).toHaveBeenCalled();
    });

    it("ignores unauthenticated channel messages", () => {
        const { props } = renderForm();

        mocks.channelListeners.forEach((handler) => handler(false));

        expect(props.onChannelStateChange).not.toHaveBeenCalled();
    });

    it("removes the listener on unmount", () => {
        const { unmount } = renderForm();

        expect(mocks.channelListeners).toHaveLength(1);

        unmount();

        expect(mocks.channelListeners).toHaveLength(0);
    });
});

describe("passkey integration", () => {
    it("clears the credentials and starts loading when passkey authentication starts", async () => {
        const { props } = renderForm({ passkeyLogin: true });

        fillCredentials();
        fireEvent.click(screen.getByTestId("passkey-start"));

        await waitFor(() => expect(getUsername()).toHaveValue(""));
        expect(getPassword()).toHaveValue("");
        expect(props.onAuthenticationStart).toHaveBeenCalled();
        expect(screen.getByTestId("passkey-form")).toHaveAttribute("data-disabled", "true");
    });

    it("stops loading when passkey authentication stops", async () => {
        const { props } = renderForm({ passkeyLogin: true });

        fireEvent.click(screen.getByTestId("passkey-start"));
        await waitFor(() => expect(screen.getByTestId("passkey-form")).toHaveAttribute("data-disabled", "true"));

        fireEvent.click(screen.getByTestId("passkey-stop"));

        await waitFor(() => expect(screen.getByTestId("passkey-form")).toHaveAttribute("data-disabled", "false"));
        expect(props.onAuthenticationStop).toHaveBeenCalled();
    });

    it("forwards passkey authentication success", async () => {
        const { props } = renderForm({ passkeyLogin: true });

        fireEvent.click(screen.getByTestId("passkey-success"));

        await waitFor(() => expect(props.onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com"));
    });

    it("surfaces passkey authentication errors as notifications", () => {
        renderForm({ passkeyLogin: true });

        fireEvent.click(screen.getByTestId("passkey-error"));

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("passkey failed");
    });

    it("passes remember me through to the passkey form", () => {
        renderForm({ passkeyLogin: true, rememberMe: true });

        expect(screen.getByTestId("passkey-form")).toHaveAttribute("data-remember-me", "true");
    });
});
