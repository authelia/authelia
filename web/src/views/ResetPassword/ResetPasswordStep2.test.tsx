import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import ResetPasswordStep2 from "@views/ResetPassword/ResetPasswordStep2";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        createInfoNotification: vi.fn(),
        createSuccessNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => "test-token",
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
