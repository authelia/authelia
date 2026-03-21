import { act, fireEvent, render, screen } from "@testing-library/react";

import ResetPasswordStep1 from "@views/ResetPassword/ResetPasswordStep1";

const mockNavigate = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-router-dom", () => ({
    useNavigate: () => mockNavigate,
}));

vi.mock("@components/ComponentWithTooltip", () => ({
    default: (props: any) => <div>{props.children}</div>,
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

const mockCreateError = vi.fn();
const mockCreateInfo = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createInfoNotification: mockCreateInfo,
    }),
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => <div data-testid="layout">{props.children}</div>,
}));

vi.mock("@services/ResetPassword", () => ({
    initiateResetPasswordProcess: vi.fn(),
}));

beforeEach(() => {
    mockNavigate.mockReset();
    mockCreateError.mockReset();
    mockCreateInfo.mockReset();
});

it("renders the reset password form", () => {
    render(<ResetPasswordStep1 />);
    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByText("Reset")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
});

it("shows error when username is empty", async () => {
    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("Username is required");
});

it("navigates to index on successful reset", async () => {
    const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
    vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: false, retryAfter: 0 });

    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "testuser" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateInfo).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("shows rate limited error", async () => {
    const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
    vi.mocked(initiateResetPasswordProcess).mockResolvedValue({ limited: true, retryAfter: 30 });

    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "testuser" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("You have made too many requests");
});

it("navigates to index when cancel is clicked", () => {
    render(<ResetPasswordStep1 />);
    fireEvent.click(screen.getByText("Cancel"));
    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("shows error when service throws", async () => {
    const { initiateResetPasswordProcess } = await import("@services/ResetPassword");
    vi.mocked(initiateResetPasswordProcess).mockRejectedValue(new Error("network error"));

    render(<ResetPasswordStep1 />);

    await act(async () => {
        fireEvent.change(screen.getByLabelText("Username"), { target: { value: "testuser" } });
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Reset"));
    });

    expect(mockCreateError).toHaveBeenCalledWith("There was an issue initiating the password reset process");
});
