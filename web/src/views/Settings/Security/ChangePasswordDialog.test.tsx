// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { PasswordPolicyMode } from "@models/PasswordPolicy";
import { postPasswordChange } from "@services/ChangePassword";
import { getPasswordPolicyConfiguration } from "@services/PasswordPolicyConfiguration";
import ChangePasswordDialog from "@views/Settings/Security/ChangePasswordDialog";

const mocks = vi.hoisted(() => ({
    capsLockOn: false,
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

vi.mock("@hooks/CapsLock", () => ({
    default: (setter: (value: boolean) => void) => () => setter(mocks.capsLockOn),
}));

vi.mock("@components/PasswordMeter", () => ({
    default: (props: any) => <div data-testid="password-meter" data-value={props.value} />,
}));

vi.mock("@services/ChangePassword", () => ({
    postPasswordChange: vi.fn(),
}));

vi.mock("@services/PasswordPolicyConfiguration", () => ({
    getPasswordPolicyConfiguration: vi.fn(),
}));

const postPasswordChangeMock = vi.mocked(postPasswordChange);
const getPolicyMock = vi.mocked(getPasswordPolicyConfiguration);

const disabledPolicy = {
    max_length: 0,
    min_length: 8,
    min_score: 0,
    mode: PasswordPolicyMode.Disabled,
    require_lowercase: false,
    require_number: false,
    require_special: false,
    require_uppercase: false,
};

function renderDialog(
    props: Partial<{ disabled: boolean; open: boolean; setClosed: () => void; username: string }> = {},
) {
    const setClosed = props.setClosed ?? vi.fn();
    const merged = { disabled: false, open: true, username: "john", ...props, setClosed };

    return { ...render(<ChangePasswordDialog {...merged} />), setClosed };
}

function getOldPassword() {
    return document.getElementById("old-password") as HTMLInputElement;
}

function getNewPassword() {
    return document.getElementById("new-password") as HTMLInputElement;
}

function getRepeatNewPassword() {
    return document.getElementById("repeat-new-password") as HTMLInputElement;
}

function getSubmit() {
    return document.getElementById("password-change-dialog-submit") as HTMLButtonElement;
}

function fillForm(oldPw = "old-secret", newPw = "new-secret", repeatPw = "new-secret") {
    fireEvent.change(getOldPassword(), { target: { value: oldPw } });
    fireEvent.change(getNewPassword(), { target: { value: newPw } });
    fireEvent.change(getRepeatNewPassword(), { target: { value: repeatPw } });
}

beforeEach(() => {
    vi.clearAllMocks();
    mocks.capsLockOn = false;
    vi.spyOn(console, "error").mockImplementation(() => {});
    getPolicyMock.mockResolvedValue(disabledPolicy);
    postPasswordChangeMock.mockResolvedValue(undefined as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders dialog with change password title when open", async () => {
        renderDialog();
        expect(screen.getByText("Change Password")).toBeInTheDocument();
        expect(screen.getByText("Cancel")).toBeInTheDocument();
        expect(screen.getByText("Submit")).toBeInTheDocument();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());
    });

    it("does not render content when closed", () => {
        renderDialog({ open: false });
        expect(screen.queryByText("Submit")).not.toBeInTheDocument();
    });

    it("renders all three password fields", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());
        expect(getOldPassword()).toHaveAttribute("type", "password");
        expect(getNewPassword()).toHaveAttribute("type", "password");
        expect(getRepeatNewPassword()).toHaveAttribute("type", "password");
    });

    it("disables the fields when the dialog is disabled", async () => {
        renderDialog({ disabled: true });
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());
        expect(getOldPassword()).toBeDisabled();
        expect(getNewPassword()).toBeDisabled();
        expect(getRepeatNewPassword()).toBeDisabled();
    });

    it("keeps submit disabled until every field is filled", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        expect(getSubmit()).toBeDisabled();

        fireEvent.change(getOldPassword(), { target: { value: "old-secret" } });
        expect(getSubmit()).toBeDisabled();

        fireEvent.change(getNewPassword(), { target: { value: "new-secret" } });
        expect(getSubmit()).toBeDisabled();

        fireEvent.change(getRepeatNewPassword(), { target: { value: "new-secret" } });
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
    });
});

describe("password visibility toggle", () => {
    it("gives each toggle a distinct accessible name", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        expect(screen.getByLabelText("Toggle old password visibility")).toBeInTheDocument();
        expect(screen.getByLabelText("Toggle new password visibility")).toBeInTheDocument();
        expect(screen.getByLabelText("Toggle repeat new password visibility")).toBeInTheDocument();
    });

    it("reveals only the old password when its toggle is clicked", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.click(screen.getByLabelText("Toggle old password visibility"));

        expect(getOldPassword()).toHaveAttribute("type", "text");
        expect(getNewPassword()).toHaveAttribute("type", "password");
        expect(getRepeatNewPassword()).toHaveAttribute("type", "password");
    });

    it("reveals the new and repeat passwords together when the new password toggle is clicked", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.click(screen.getByLabelText("Toggle new password visibility"));

        expect(getOldPassword()).toHaveAttribute("type", "password");
        expect(getNewPassword()).toHaveAttribute("type", "text");
        expect(getRepeatNewPassword()).toHaveAttribute("type", "text");
    });

    it("reveals the new and repeat passwords together when the repeat password toggle is clicked", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.click(screen.getByLabelText("Toggle repeat new password visibility"));

        expect(getOldPassword()).toHaveAttribute("type", "password");
        expect(getNewPassword()).toHaveAttribute("type", "text");
        expect(getRepeatNewPassword()).toHaveAttribute("type", "text");
    });
});

describe("password policy", () => {
    it("hides the password meter when the policy is disabled", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());
        expect(screen.queryByTestId("password-meter")).not.toBeInTheDocument();
    });

    it("shows the password meter when a standard policy is configured", async () => {
        getPolicyMock.mockResolvedValue({ ...disabledPolicy, mode: PasswordPolicyMode.Standard });

        renderDialog();

        expect(await screen.findByTestId("password-meter")).toBeInTheDocument();
    });

    it("shows the password meter when a zxcvbn policy is configured", async () => {
        getPolicyMock.mockResolvedValue({ ...disabledPolicy, mode: PasswordPolicyMode.ZXCVBN });

        renderDialog();

        expect(await screen.findByTestId("password-meter")).toBeInTheDocument();
    });

    it("feeds the new password into the meter", async () => {
        getPolicyMock.mockResolvedValue({ ...disabledPolicy, mode: PasswordPolicyMode.Standard });

        renderDialog();
        await screen.findByTestId("password-meter");

        fireEvent.change(getNewPassword(), { target: { value: "new-secret" } });

        await waitFor(() => expect(screen.getByTestId("password-meter")).toHaveAttribute("data-value", "new-secret"));
    });

    it("notifies when the policy cannot be retrieved", async () => {
        getPolicyMock.mockRejectedValue(new Error("boom"));

        renderDialog();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "There was an issue completing the process the verification token might have expired",
            ),
        );
    });
});

describe("validation", () => {
    it("flags every empty field on submit", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm("old-secret", "new-secret", "new-secret");
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());

        fireEvent.change(getOldPassword(), { target: { value: "  " } });
        fireEvent.change(getNewPassword(), { target: { value: "  " } });
        fireEvent.change(getRepeatNewPassword(), { target: { value: "  " } });
        fireEvent.click(getSubmit());

        await waitFor(() => expect(getOldPassword()).toHaveAttribute("aria-invalid", "true"));
        expect(getNewPassword()).toHaveAttribute("aria-invalid", "true");
        expect(getRepeatNewPassword()).toHaveAttribute("aria-invalid", "true");
        expect(postPasswordChangeMock).not.toHaveBeenCalled();
    });

    it("flags only the whitespace-only fields", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm("old-secret", "  ", "new-secret");
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());

        fireEvent.click(getSubmit());

        await waitFor(() => expect(getNewPassword()).toHaveAttribute("aria-invalid", "true"));
        expect(getOldPassword()).not.toHaveAttribute("aria-invalid", "true");
        expect(getRepeatNewPassword()).not.toHaveAttribute("aria-invalid", "true");
        expect(postPasswordChangeMock).not.toHaveBeenCalled();
    });

    it("rejects mismatched new passwords", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm("old-secret", "new-secret", "other-secret");
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());

        fireEvent.click(getSubmit());

        await waitFor(() => expect(mocks.createErrorNotification).toHaveBeenCalledWith("Passwords do not match"));
        expect(getNewPassword()).toHaveAttribute("aria-invalid", "true");
        expect(getRepeatNewPassword()).toHaveAttribute("aria-invalid", "true");
        expect(postPasswordChangeMock).not.toHaveBeenCalled();
    });

    it("clears field errors on focus", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm("old-secret", "new-secret", "other-secret");
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());
        await waitFor(() => expect(getNewPassword()).toHaveAttribute("aria-invalid", "true"));

        fireEvent.focus(getNewPassword());
        await waitFor(() => expect(getNewPassword()).not.toHaveAttribute("aria-invalid", "true"));

        fireEvent.focus(getRepeatNewPassword());
        await waitFor(() => expect(getRepeatNewPassword()).not.toHaveAttribute("aria-invalid", "true"));

        fireEvent.focus(getOldPassword());
        await waitFor(() => expect(getOldPassword()).not.toHaveAttribute("aria-invalid", "true"));
    });
});

describe("submission", () => {
    it("posts the change and reports success", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() => expect(postPasswordChangeMock).toHaveBeenCalledWith("john", "old-secret", "new-secret"));
        expect(mocks.createSuccessNotification).toHaveBeenCalledWith("Password changed successfully");
        expect(setClosed).toHaveBeenCalled();
    });

    it("flags the new password when the server rejects it as too weak", async () => {
        postPasswordChangeMock.mockRejectedValue({ isAxiosError: true, response: { status: 400 } });

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Your supplied password does not meet the password policy requirements",
            ),
        );
        expect(getNewPassword()).toHaveAttribute("aria-invalid", "true");
        expect(getRepeatNewPassword()).toHaveAttribute("aria-invalid", "true");
    });

    it("flags the old password when the server reports it as incorrect", async () => {
        postPasswordChangeMock.mockRejectedValue({ isAxiosError: true, response: { status: 401 } });

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() => expect(mocks.createErrorNotification).toHaveBeenCalledWith("Incorrect password"));
        expect(getOldPassword()).toHaveAttribute("aria-invalid", "true");
    });

    it("reports a generic error on a server failure", async () => {
        postPasswordChangeMock.mockRejectedValue({ isAxiosError: true, response: { status: 500 } });

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue changing the password"),
        );
    });

    it("reports a generic error on an unexpected status", async () => {
        postPasswordChangeMock.mockRejectedValue({ isAxiosError: true, response: { status: 418 } });

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue changing the password"),
        );
    });

    it("reports a generic error on a non-axios failure", async () => {
        postPasswordChangeMock.mockRejectedValue(new Error("network down"));

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue changing the password"),
        );
    });

    it("reports a generic error on an axios error without a response", async () => {
        postPasswordChangeMock.mockRejectedValue({ isAxiosError: true });

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue changing the password"),
        );
    });

    it("re-enables submission after a failure", async () => {
        postPasswordChangeMock.mockRejectedValue({ isAxiosError: true, response: { status: 401 } });

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() => expect(mocks.createErrorNotification).toHaveBeenCalled());
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
    });
});

describe("keyboard navigation", () => {
    it("moves from the old password to the new password on Enter", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.change(getOldPassword(), { target: { value: "old-secret" } });
        fireEvent.keyDown(getOldPassword(), { key: "Enter" });

        await waitFor(() => expect(getNewPassword()).toHaveFocus());
    });

    it("flags the old password when Enter is pressed while it is empty", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyDown(getOldPassword(), { key: "Enter" });

        await waitFor(() => expect(getOldPassword()).toHaveAttribute("aria-invalid", "true"));
    });

    it("moves from the new password to the repeat field on Enter", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.change(getNewPassword(), { target: { value: "new-secret" } });
        fireEvent.keyDown(getNewPassword(), { key: "Enter" });

        await waitFor(() => expect(getRepeatNewPassword()).toHaveFocus());
    });

    it("flags the new password when Enter is pressed while it is empty", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyDown(getNewPassword(), { key: "Enter" });

        await waitFor(() => expect(getNewPassword()).toHaveAttribute("aria-invalid", "true"));
    });

    it("submits when Enter is pressed on the repeat field", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        fireEvent.keyDown(getRepeatNewPassword(), { key: "Enter" });

        await waitFor(() => expect(postPasswordChangeMock).toHaveBeenCalledWith("john", "old-secret", "new-secret"));
    });

    it("flags the repeat field when Enter is pressed while it is empty", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyDown(getRepeatNewPassword(), { key: "Enter" });

        await waitFor(() => expect(getRepeatNewPassword()).toHaveAttribute("aria-invalid", "true"));
        expect(postPasswordChangeMock).not.toHaveBeenCalled();
    });

    it("ignores keys other than Enter", async () => {
        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyDown(getOldPassword(), { key: "a" });
        fireEvent.keyDown(getNewPassword(), { key: "a" });
        fireEvent.keyDown(getRepeatNewPassword(), { key: "a" });

        expect(getOldPassword()).not.toHaveAttribute("aria-invalid", "true");
        expect(postPasswordChangeMock).not.toHaveBeenCalled();
    });
});

describe("caps lock", () => {
    it.each([
        ["old-password", () => getOldPassword()],
        ["new-password", () => getNewPassword()],
        ["repeat-new-password", () => getRepeatNewPassword()],
    ])("warns when caps lock is on in %s", async (_name, getField) => {
        mocks.capsLockOn = true;

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyUp(getField(), { key: "A" });

        expect(await screen.findByText("Caps Lock is on")).toBeInTheDocument();
    });

    it("clears the caps lock warning on blur", async () => {
        mocks.capsLockOn = true;

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyUp(getOldPassword(), { key: "A" });
        await screen.findByText("Caps Lock is on");

        fireEvent.blur(getOldPassword());

        await waitFor(() => expect(screen.queryByText("Caps Lock is on")).not.toBeInTheDocument());
    });

    it("does not warn when caps lock is off", async () => {
        mocks.capsLockOn = false;

        renderDialog();
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyUp(getOldPassword(), { key: "a" });

        expect(screen.queryByText("Caps Lock is on")).not.toBeInTheDocument();
    });
});

describe("closing", () => {
    it("closes and resets the form when cancelled", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        fireEvent.click(document.getElementById("password-change-dialog-cancel") as HTMLButtonElement);

        expect(setClosed).toHaveBeenCalled();
    });
});

describe("dismissal", () => {
    it("closes when Escape is pressed", async () => {
        const setClosed = vi.fn();

        renderDialog({ setClosed });
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        await waitFor(() => expect(setClosed).toHaveBeenCalled());
    });

    it("ignores Escape while a change is being submitted", async () => {
        let resolve: (value: unknown) => void = () => {};
        postPasswordChangeMock.mockReturnValue(new Promise((r) => (resolve = r)) as any);

        const setClosed = vi.fn();
        renderDialog({ setClosed });
        await waitFor(() => expect(getPolicyMock).toHaveBeenCalled());

        fillForm();
        await waitFor(() => expect(getSubmit()).not.toBeDisabled());
        fireEvent.click(getSubmit());

        await waitFor(() => expect(postPasswordChangeMock).toHaveBeenCalled());

        fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

        expect(setClosed).not.toHaveBeenCalled();

        resolve(undefined);
        await waitFor(() => expect(setClosed).toHaveBeenCalled());
    });
});
