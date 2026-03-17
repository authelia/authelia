import { render, screen } from "@testing-library/react";

import ChangePasswordDialog from "@views/Settings/Security/ChangePasswordDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        createSuccessNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/CapsLock", () => ({
    default: () => vi.fn(),
}));

vi.mock("@components/PasswordMeter", () => ({
    default: () => <div data-testid="password-meter" />,
}));

vi.mock("@services/ChangePassword", () => ({
    postPasswordChange: vi.fn(),
}));

vi.mock("@services/PasswordPolicyConfiguration", () => ({
    getPasswordPolicyConfiguration: vi.fn().mockResolvedValue({
        max_length: 0,
        min_length: 8,
        min_score: 0,
        mode: "disabled",
        require_lowercase: false,
        require_number: false,
        require_special: false,
        require_uppercase: false,
    }),
}));

it("renders dialog with change password title when open", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(<ChangePasswordDialog username="john" open={true} setClosed={vi.fn()} />);
    expect(screen.getByText("Change Password")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
    expect(screen.getByText("Submit")).toBeInTheDocument();
});

it("does not render content when closed", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(<ChangePasswordDialog username="john" open={false} setClosed={vi.fn()} />);
    expect(screen.queryByText("Submit")).not.toBeInTheDocument();
});
