import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { postSecondFactor } from "@services/Password";
import PasswordForm from "@views/LoginPortal/SecondFactor/PasswordForm";

const mocks = vi.hoisted(() => ({
    capsLockModified: null as boolean | null,
    createErrorNotification: vi.fn(),
    queryParams: {} as Record<string, null | string>,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: (name: string) => mocks.queryParams[name] ?? null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: "flow", id: "flow-id", subflow: "subflow" }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

vi.mock("@services/Password", () => ({
    postSecondFactor: vi.fn(),
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: () => mocks.capsLockModified,
}));

const postSecondFactorMock = vi.mocked(postSecondFactor);

function renderForm(onAuthenticationSuccess = vi.fn()) {
    return { ...render(<PasswordForm onAuthenticationSuccess={onAuthenticationSuccess} />), onAuthenticationSuccess };
}

function getPassword() {
    return screen.getByLabelText(/Password/) as HTMLInputElement;
}

function getSubmit() {
    return document.getElementById("sign-in-button") as HTMLButtonElement;
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.capsLockModified = null;
    mocks.queryParams = {};
    vi.spyOn(console, "error").mockImplementation(() => {});
    postSecondFactorMock.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders the password form", () => {
        renderForm();
        expect(screen.getByText(/Password/)).toBeInTheDocument();
        expect(screen.getByText("Authenticate")).toBeInTheDocument();
    });

    it("focuses the password field on mount", async () => {
        renderForm();
        await waitFor(() => expect(getPassword()).toHaveFocus());
    });

    it("starts with the password hidden", () => {
        renderForm();
        expect(getPassword()).toHaveAttribute("type", "password");
    });
});

describe("validation", () => {
    it("shows error when submitting empty password", async () => {
        renderForm();

        fireEvent.click(getSubmit());

        await waitFor(() => expect(getPassword()).toHaveAttribute("aria-invalid", "true"));
        expect(postSecondFactorMock).not.toHaveBeenCalled();
    });

    it("clears the error when the field is focused", async () => {
        renderForm();

        fireEvent.click(getSubmit());
        await waitFor(() => expect(getPassword()).toHaveAttribute("aria-invalid", "true"));

        fireEvent.focus(getPassword());

        await waitFor(() => expect(getPassword()).not.toHaveAttribute("aria-invalid", "true"));
    });
});

describe("sign in", () => {
    it("posts the password and reports success", async () => {
        postSecondFactorMock.mockResolvedValue({ redirect: "https://example.com" } as any);

        const { onAuthenticationSuccess } = renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.click(getSubmit());

        await waitFor(() =>
            expect(postSecondFactorMock).toHaveBeenCalledWith("secret", null, "flow-id", "flow", "subflow"),
        );
        expect(onAuthenticationSuccess).toHaveBeenCalledWith("https://example.com");
    });

    it("reports success with an undefined redirect when there is no response", async () => {
        postSecondFactorMock.mockResolvedValue(undefined as any);

        const { onAuthenticationSuccess } = renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.click(getSubmit());

        await waitFor(() => expect(onAuthenticationSuccess).toHaveBeenCalledWith(undefined));
    });

    it("forwards the redirection URL", async () => {
        mocks.queryParams = { rd: "https://app.example.com" };
        postSecondFactorMock.mockResolvedValue(undefined as any);

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.click(getSubmit());

        await waitFor(() =>
            expect(postSecondFactorMock).toHaveBeenCalledWith(
                "secret",
                "https://app.example.com",
                "flow-id",
                "flow",
                "subflow",
            ),
        );
    });

    it("disables the field and button while submitting", async () => {
        let resolve: (value: unknown) => void = () => {};
        postSecondFactorMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.click(getSubmit());

        await waitFor(() => expect(getSubmit()).toBeDisabled());
        expect(getPassword()).toBeDisabled();

        resolve(undefined);
        await waitFor(() => expect(postSecondFactorMock).toHaveBeenCalled());
    });

    it("notifies, clears the password and refocuses it on failure", async () => {
        postSecondFactorMock.mockRejectedValue(new Error("bad password"));

        const { onAuthenticationSuccess } = renderForm();

        await waitFor(() => expect(getPassword()).toHaveFocus());

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.click(getSubmit());

        await waitFor(() => expect(mocks.createErrorNotification).toHaveBeenCalledWith("Incorrect password"));
        await waitFor(() => expect(getPassword()).toHaveValue(""));
        expect(getPassword()).toHaveFocus();
        expect(getSubmit()).not.toBeDisabled();
        expect(onAuthenticationSuccess).not.toHaveBeenCalled();
    });
});

describe("keyboard handling", () => {
    it("signs in when Enter is pressed", async () => {
        postSecondFactorMock.mockResolvedValue(undefined as any);

        renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.keyDown(getPassword(), { key: "Enter" });

        await waitFor(() => expect(postSecondFactorMock).toHaveBeenCalled());
    });

    it("refocuses the field when Enter is pressed while empty", async () => {
        renderForm();

        fireEvent.keyDown(getPassword(), { key: "Enter" });

        await waitFor(() => expect(getPassword()).toHaveFocus());
        expect(postSecondFactorMock).not.toHaveBeenCalled();
        await waitFor(() => expect(getPassword()).toHaveAttribute("aria-invalid", "true"));
    });

    it("ignores keys other than Enter", () => {
        renderForm();

        fireEvent.change(getPassword(), { target: { value: "secret" } });
        fireEvent.keyDown(getPassword(), { key: "a" });

        expect(postSecondFactorMock).not.toHaveBeenCalled();
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

    it("does not warn when the caps lock state is unknown", () => {
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

        fireEvent.click(toggle);
        await waitFor(() => expect(getPassword()).toHaveAttribute("type", "text"));
        expect(toggle).toHaveAttribute("aria-pressed", "true");

        fireEvent.click(toggle);
        await waitFor(() => expect(getPassword()).toHaveAttribute("type", "password"));
        expect(toggle).toHaveAttribute("aria-pressed", "false");
    });
});
