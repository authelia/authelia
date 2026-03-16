import { act, fireEvent, render, screen } from "@testing-library/react";

import PasswordForm from "@views/LoginPortal/SecondFactor/PasswordForm";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: null, id: null, subflow: null }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

const mockPostSecondFactor = vi.fn();

vi.mock("@services/Password", () => ({
    postSecondFactor: (...args: any[]) => mockPostSecondFactor(...args),
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: () => null,
}));

it("renders the password form", () => {
    render(<PasswordForm onAuthenticationSuccess={vi.fn()} />);
    expect(screen.getByText(/Password/)).toBeInTheDocument();
    expect(screen.getByText("Authenticate")).toBeInTheDocument();
});

beforeEach(() => {
    mockPostSecondFactor.mockReset();
});

it("shows error when submitting empty password", async () => {
    render(<PasswordForm onAuthenticationSuccess={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Authenticate"));
    });

    expect(mockPostSecondFactor).not.toHaveBeenCalled();
    expect(screen.getByLabelText(/Password/)).toHaveClass("border-destructive");
});
