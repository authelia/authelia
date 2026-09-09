import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { PasswordPolicyMode } from "@models/PasswordPolicy";
import { getPasswordPolicyConfiguration } from "@services/PasswordPolicyConfiguration";
import { completeResetPasswordProcess, resetPassword } from "@services/ResetPassword";
import ResetPasswordStep2 from "@views/ResetPassword/ResetPasswordStep2";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    createInfoNotification: vi.fn(),
    createSuccessNotification: vi.fn(),
    navigate: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
        createInfoNotification: mocks.createInfoNotification,
        createSuccessNotification: mocks.createSuccessNotification,
    }),
}));

const queryParam = vi.hoisted(() => ({ token: "test-token" as string | undefined }));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => queryParam.token,
}));

vi.mock("react-router", async () => {
    const actual = await vi.importActual("react-router");
    return { ...actual, useNavigate: () => mocks.navigate };
});

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => <div data-testid="layout">{props.children}</div>,
}));

vi.mock("@services/ResetPassword", () => ({
    completeResetPasswordProcess: vi.fn(),
    resetPassword: vi.fn(),
}));

vi.mock("@services/PasswordPolicyConfiguration", () => ({
    getPasswordPolicyConfiguration: vi.fn(),
}));

vi.mock("@components/PasswordMeter", () => ({
    default: (props: any) => <div data-testid="password-meter" data-value={props.value} />,
}));

const completeProcessMock = vi.mocked(completeResetPasswordProcess);
const resetPasswordMock = vi.mocked(resetPassword);
const getPolicyMock = vi.mocked(getPasswordPolicyConfiguration);

const standardPolicy = {
    max_length: 0,
    min_length: 8,
    min_score: 0,
    mode: PasswordPolicyMode.Standard,
    require_lowercase: false,
    require_number: false,
    require_special: false,
    require_uppercase: false,
};

function renderStep() {
    return render(
        <MemoryRouter>
            <ResetPasswordStep2 />
        </MemoryRouter>,
    );
}

function getPassword1() {
    return screen.getByLabelText("New password") as HTMLInputElement;
}

function getPassword2() {
    return screen.getByLabelText("Repeat new password") as HTMLInputElement;
}

function getReset() {
    return screen.getByRole("button", { name: "Reset" });
}

async function renderReady() {
    const result = renderStep();
    await waitFor(() => expect(getReset()).not.toBeDisabled());
    return result;
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    queryParam.token = "test-token";
    completeProcessMock.mockResolvedValue({} as any);
    getPolicyMock.mockResolvedValue(standardPolicy);
    resetPasswordMock.mockResolvedValue(undefined as any);
});

afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
});

describe("rendering", () => {
    it("renders the reset password form", () => {
        renderStep();
        expect(screen.getAllByText("New password").length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText("Reset")).toBeInTheDocument();
        expect(screen.getByText("Cancel")).toBeInTheDocument();
    });

    it("starts disabled until the process completes", () => {
        renderStep();
        expect(getReset()).toBeDisabled();
        expect(getPassword1()).toBeDisabled();
        expect(getPassword2()).toBeDisabled();
    });

    it("enables the form once the process and policy resolve", async () => {
        await renderReady();
        expect(getPassword1()).not.toBeDisabled();
        expect(getPassword2()).not.toBeDisabled();
    });

    it("shows the password meter for an enabled policy", async () => {
        await renderReady();
        expect(screen.getByTestId("password-meter")).toBeInTheDocument();
    });

    it("hides the password meter when the policy is disabled", async () => {
        getPolicyMock.mockResolvedValue({ ...standardPolicy, mode: PasswordPolicyMode.Disabled });

        await renderReady();

        expect(screen.queryByTestId("password-meter")).not.toBeInTheDocument();
    });

    it("feeds the new password into the meter", async () => {
        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });

        await waitFor(() => expect(screen.getByTestId("password-meter")).toHaveAttribute("data-value", "password123"));
    });
});

describe("process initiation", () => {
    it("starts the process once a token becomes available", async () => {
        queryParam.token = undefined;

        const { rerender } = renderStep();

        expect(completeProcessMock).not.toHaveBeenCalled();
        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("No verification token provided"),
        );

        queryParam.token = "test-token";

        rerender(
            <MemoryRouter>
                <ResetPasswordStep2 />
            </MemoryRouter>,
        );

        await waitFor(() => expect(completeProcessMock).toHaveBeenCalledWith("test-token"));
    });

    it("does not repeat the process when re-rendered with the same token", async () => {
        const { rerender } = renderStep();

        await waitFor(() => expect(completeProcessMock).toHaveBeenCalledTimes(1));

        rerender(
            <MemoryRouter>
                <ResetPasswordStep2 />
            </MemoryRouter>,
        );

        await waitFor(() => expect(completeProcessMock).toHaveBeenCalledTimes(1));
    });

    it("notifies and keeps the form disabled when rate limited", async () => {
        completeProcessMock.mockResolvedValue({ limited: true, retryAfter: 30 } as any);

        renderStep();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith("You have made too many requests"),
        );
        expect(getReset()).toBeDisabled();
        expect(getPolicyMock).not.toHaveBeenCalled();
    });

    it("notifies and keeps the form disabled when the token expired", async () => {
        completeProcessMock.mockRejectedValue(new Error("expired"));

        renderStep();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "There was an issue completing the process the verification token might have expired",
            ),
        );
        expect(getReset()).toBeDisabled();
    });

    it("notifies when the policy cannot be retrieved", async () => {
        getPolicyMock.mockRejectedValue(new Error("boom"));

        renderStep();

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "There was an issue completing the process the verification token might have expired",
            ),
        );
    });
});

describe("validation", () => {
    it("flags both fields when both are empty", async () => {
        await renderReady();

        fireEvent.click(getReset());

        await waitFor(() => expect(getPassword1()).toHaveAttribute("aria-invalid", "true"));
        expect(getPassword2()).toHaveAttribute("aria-invalid", "true");
        expect(resetPasswordMock).not.toHaveBeenCalled();
    });

    it("flags only the repeat field when the first is filled", async () => {
        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.click(getReset());

        await waitFor(() => expect(getPassword2()).toHaveAttribute("aria-invalid", "true"));
        expect(getPassword1()).not.toHaveAttribute("aria-invalid", "true");
        expect(resetPasswordMock).not.toHaveBeenCalled();
    });

    it("flags only the first field when the repeat is filled", async () => {
        await renderReady();

        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.click(getReset());

        await waitFor(() => expect(getPassword1()).toHaveAttribute("aria-invalid", "true"));
        expect(getPassword2()).not.toHaveAttribute("aria-invalid", "true");
    });

    it("rejects mismatched passwords", async () => {
        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "different" } });
        fireEvent.click(getReset());

        await waitFor(() => expect(mocks.createErrorNotification).toHaveBeenCalledWith("Passwords do not match"));
        expect(getPassword1()).toHaveAttribute("aria-invalid", "true");
        expect(getPassword2()).toHaveAttribute("aria-invalid", "true");
        expect(resetPasswordMock).not.toHaveBeenCalled();
    });
});

describe("resetting", () => {
    it("resets the password and navigates home after the delay", async () => {
        vi.useFakeTimers();

        renderStep();
        await vi.waitFor(() => expect(getReset()).not.toBeDisabled());

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.click(getReset());

        await vi.waitFor(() => expect(resetPasswordMock).toHaveBeenCalledWith("password123"));
        expect(mocks.createSuccessNotification).toHaveBeenCalledWith("Password has been reset");
        expect(mocks.navigate).not.toHaveBeenCalled();

        await act(async () => {
            await vi.advanceTimersByTimeAsync(1500);
        });

        expect(mocks.navigate).toHaveBeenCalledWith("/");
    });

    it("resets when Enter is pressed on the repeat field", async () => {
        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.keyDown(getPassword2(), { key: "Enter" });

        await waitFor(() => expect(resetPasswordMock).toHaveBeenCalledWith("password123"));
    });

    it("ignores keys other than Enter on the repeat field", async () => {
        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.keyDown(getPassword2(), { key: "a" });

        expect(resetPasswordMock).not.toHaveBeenCalled();
    });

    it("re-enables the form when the reset fails", async () => {
        resetPasswordMock.mockRejectedValue(new Error("There was an issue resetting the password"));

        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.click(getReset());

        await waitFor(() => expect(resetPasswordMock).toHaveBeenCalledWith("password123"));

        await waitFor(() => expect(getReset()).not.toBeDisabled());
        expect(getPassword1()).not.toBeDisabled();
        expect(getPassword2()).not.toBeDisabled();
        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue resetting the password");
    });

    it("reports a policy violation reported as 0000052D.", async () => {
        resetPasswordMock.mockRejectedValue(new Error("LDAP error 0000052D. rejected"));

        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.click(getReset());

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Your supplied password does not meet the password policy requirements",
            ),
        );
    });

    it("reports a policy violation reported as policy", async () => {
        resetPasswordMock.mockRejectedValue(new Error("password does not meet policy"));

        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.click(getReset());

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "Your supplied password does not meet the password policy requirements",
            ),
        );
    });

    it("clears both fields on submission", async () => {
        await renderReady();

        fireEvent.change(getPassword1(), { target: { value: "password123" } });
        fireEvent.change(getPassword2(), { target: { value: "password123" } });
        fireEvent.click(getReset());

        await waitFor(() => expect(getPassword1()).toHaveValue(""));
        expect(getPassword2()).toHaveValue("");
    });
});

describe("cancelling", () => {
    it("navigates home", async () => {
        await renderReady();

        fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

        expect(mocks.navigate).toHaveBeenCalledWith("/");
    });
});

describe("password visibility toggle", () => {
    it("reveals both fields on click and hides them again on a second click", async () => {
        await renderReady();

        const toggle = screen.getByLabelText("Toggle password visibility");

        fireEvent.click(toggle);
        await waitFor(() => expect(getPassword1()).toHaveAttribute("type", "text"));
        expect(getPassword2()).toHaveAttribute("type", "text");
        expect(toggle).toHaveAttribute("aria-pressed", "true");

        fireEvent.click(toggle);
        await waitFor(() => expect(getPassword1()).toHaveAttribute("type", "password"));
        expect(getPassword2()).toHaveAttribute("type", "password");
        expect(toggle).toHaveAttribute("aria-pressed", "false");
    });
});
