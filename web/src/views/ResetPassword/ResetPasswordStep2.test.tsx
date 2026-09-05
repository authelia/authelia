import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { completeResetPasswordProcess, resetPassword } from "@services/ResetPassword";
import ResetPasswordStep2 from "@views/ResetPassword/ResetPasswordStep2";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        createInfoNotification: vi.fn(),
        createSuccessNotification: vi.fn(),
    }),
}));

const queryParam = vi.hoisted(() => ({ token: "test-token" as string | undefined }));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => queryParam.token,
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => <div data-testid="layout">{props.children}</div>,
}));

vi.mock("@services/ResetPassword", () => ({
    completeResetPasswordProcess: vi.fn().mockResolvedValue({}),
    resetPassword: vi.fn(),
}));

vi.mock("@services/PasswordPolicyConfiguration", () => ({
    getPasswordPolicyConfiguration: vi.fn().mockResolvedValue({
        max_length: 0,
        min_length: 8,
        min_score: 0,
        mode: "standard",
        require_lowercase: false,
        require_number: false,
        require_special: false,
        require_uppercase: false,
    }),
}));

vi.mock("@components/PasswordMeter", () => ({
    default: () => <div data-testid="password-meter" />,
}));

beforeEach(() => {
    vi.clearAllMocks();

    queryParam.token = "test-token";
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the reset password form", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(
        <MemoryRouter>
            <ResetPasswordStep2 />
        </MemoryRouter>,
    );
    expect(screen.getAllByText("New password").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("Reset")).toBeInTheDocument();
});

it("starts the process once a token becomes available", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});

    queryParam.token = undefined;

    const { rerender } = render(
        <MemoryRouter>
            <ResetPasswordStep2 />
        </MemoryRouter>,
    );

    expect(completeResetPasswordProcess).not.toHaveBeenCalled();

    queryParam.token = "test-token";

    rerender(
        <MemoryRouter>
            <ResetPasswordStep2 />
        </MemoryRouter>,
    );

    await waitFor(() => expect(completeResetPasswordProcess).toHaveBeenCalledWith("test-token"));
});

it("does not repeat the process when re-rendered with the same token", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});

    const { rerender } = render(
        <MemoryRouter>
            <ResetPasswordStep2 />
        </MemoryRouter>,
    );

    await waitFor(() => expect(completeResetPasswordProcess).toHaveBeenCalledTimes(1));

    rerender(
        <MemoryRouter>
            <ResetPasswordStep2 />
        </MemoryRouter>,
    );

    await waitFor(() => expect(completeResetPasswordProcess).toHaveBeenCalledTimes(1));
});

it("re-enables the form when the reset fails", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});

    vi.mocked(resetPassword).mockRejectedValue(new Error("There was an issue resetting the password"));

    render(
        <MemoryRouter>
            <ResetPasswordStep2 />
        </MemoryRouter>,
    );

    const password1 = screen.getByLabelText("New password");
    const password2 = screen.getByLabelText("Repeat new password");
    const reset = screen.getByRole("button", { name: "Reset" });

    await waitFor(() => expect(reset).not.toBeDisabled());

    fireEvent.change(password1, { target: { value: "password123" } });
    fireEvent.change(password2, { target: { value: "password123" } });
    fireEvent.click(reset);

    await waitFor(() => expect(resetPassword).toHaveBeenCalledWith("password123"));

    await waitFor(() => expect(reset).not.toBeDisabled());
    expect(password1).not.toBeDisabled();
    expect(password2).not.toBeDisabled();
});
