import { render, screen } from "@testing-library/react";

import SecurityView from "@views/Settings/Security/SecurityView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@mui/material", async () => {
    const actual = await vi.importActual("@mui/material");
    return {
        ...actual,
        useTheme: () => ({
            palette: { grey: { 600: "#999" } },
            spacing: (n: number) => `${(n || 1) * 8}px`,
        }),
    };
});

vi.mock("@hooks/Configuration", () => ({
    useConfiguration: () => [{ password_change_disabled: false }, vi.fn(), false, null],
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        createSuccessNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoGET: () => [
        { display_name: "John Doe", emails: ["john@example.com"], groups: [] },
        vi.fn(),
        false,
        null,
    ],
}));

vi.mock("@services/UserSessionElevation", () => ({
    getUserSessionElevation: vi.fn().mockResolvedValue({ elevated: false }),
}));

vi.mock("@views/Settings/Common/IdentityVerificationDialog", () => ({
    default: () => <div data-testid="identity-dialog" />,
}));

vi.mock("@views/Settings/Common/SecondFactorDialog", () => ({
    default: () => <div data-testid="second-factor-dialog" />,
}));

vi.mock("@views/Settings/Security/ChangePasswordDialog", () => ({
    default: () => <div data-testid="change-password-dialog" />,
}));

it("renders user info and change password button", () => {
    render(<SecurityView />);
    expect(screen.getByText(/John Doe/)).toBeInTheDocument();
    expect(screen.getByText("Change Password")).toBeInTheDocument();
});

it("renders dialogs", () => {
    render(<SecurityView />);
    expect(screen.getByTestId("identity-dialog")).toBeInTheDocument();
    expect(screen.getByTestId("second-factor-dialog")).toBeInTheDocument();
    expect(screen.getByTestId("change-password-dialog")).toBeInTheDocument();
});
