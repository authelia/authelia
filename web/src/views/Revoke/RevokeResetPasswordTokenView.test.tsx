import { act, render, screen } from "@testing-library/react";

import RevokeResetPasswordTokenView from "@views/Revoke/RevokeResetPasswordTokenView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

const mockNavigate = vi.fn();
const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

vi.mock("@hooks/Revoke", () => ({
    useToken: vi.fn(() => "test-token"),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

vi.mock("@services/ResetPassword", () => ({
    deleteResetPasswordToken: vi.fn(() => ({ ok: true, status: 200 })),
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading" />,
}));

beforeEach(() => {
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    mockNavigate.mockReset();
});

it("calls deleteResetPasswordToken with the token", async () => {
    const { deleteResetPasswordToken } = await import("@services/ResetPassword");

    await act(async () => {
        render(<RevokeResetPasswordTokenView />);
    });

    expect(deleteResetPasswordToken).toHaveBeenCalledWith("test-token");
});

it("shows error when token is not provided", async () => {
    const { useToken } = await import("@hooks/Revoke");
    vi.mocked(useToken).mockReturnValueOnce(null as any);

    await act(async () => {
        render(<RevokeResetPasswordTokenView />);
    });

    expect(mockCreateError).toHaveBeenCalledWith("The Token was not provided");
});

it("shows rate limited error on 429 response", async () => {
    const { deleteResetPasswordToken } = await import("@services/ResetPassword");
    vi.mocked(deleteResetPasswordToken).mockResolvedValueOnce({ ok: false, status: 429 });

    await act(async () => {
        render(<RevokeResetPasswordTokenView />);
    });

    expect(mockCreateError).toHaveBeenCalledWith("You have made too many requests");
});

it("reports a successful revocation", async () => {
    const { deleteResetPasswordToken } = await import("@services/ResetPassword");
    vi.mocked(deleteResetPasswordToken).mockResolvedValue({ ok: true, status: 200 } as any);

    await act(async () => {
        render(<RevokeResetPasswordTokenView />);
    });

    expect(mockCreateSuccess).toHaveBeenCalledWith("Successfully revoked the Token");
    expect(mockCreateError).not.toHaveBeenCalled();
});

it("reports a failed revocation", async () => {
    const { deleteResetPasswordToken } = await import("@services/ResetPassword");
    vi.mocked(deleteResetPasswordToken).mockResolvedValue({ ok: false, status: 400 } as any);

    await act(async () => {
        render(<RevokeResetPasswordTokenView />);
    });

    expect(mockCreateError).toHaveBeenCalledWith("Failed to revoke the Token");
});

it("navigates back to the index route after the delay", async () => {
    vi.useFakeTimers();

    const { deleteResetPasswordToken } = await import("@services/ResetPassword");
    vi.mocked(deleteResetPasswordToken).mockResolvedValue({ ok: true, status: 200 } as any);

    render(<RevokeResetPasswordTokenView />);

    await vi.waitFor(() => expect(mockCreateSuccess).toHaveBeenCalled());
    expect(mockNavigate).not.toHaveBeenCalled();

    await act(async () => {
        await vi.advanceTimersByTimeAsync(1500);
    });

    expect(mockNavigate).toHaveBeenCalledWith("/", false);

    vi.useRealTimers();
});

it("navigates back to the index route when the token is missing", async () => {
    vi.useFakeTimers();

    const { useToken } = await import("@hooks/Revoke");
    vi.mocked(useToken).mockReturnValue(null as any);

    render(<RevokeResetPasswordTokenView />);

    await act(async () => {
        await vi.advanceTimersByTimeAsync(1500);
    });

    expect(mockNavigate).toHaveBeenCalledWith("/", false);

    vi.mocked(useToken).mockReturnValue("test-token" as any);
    vi.useRealTimers();
});

it("renders the loading page while revoking", async () => {
    await act(async () => {
        render(<RevokeResetPasswordTokenView />);
    });

    expect(screen.getByTestId("loading")).toBeInTheDocument();
});
