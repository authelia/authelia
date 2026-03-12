import { act, fireEvent, render, screen } from "@testing-library/react";

import OneTimePasswordDeleteDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordDeleteDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

const mockCreateError = vi.fn();
const mockCreateSuccess = vi.fn();

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
        createSuccessNotification: mockCreateSuccess,
    }),
}));

vi.mock("@services/UserInfoTOTPConfiguration", () => ({
    deleteUserTOTPConfiguration: vi.fn(),
}));

beforeEach(() => {
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
});

it("calls handleClose on successful deletion", async () => {
    const { deleteUserTOTPConfiguration } = await import("@services/UserInfoTOTPConfiguration");
    vi.mocked(deleteUserTOTPConfiguration).mockResolvedValue({ data: { status: "OK" } } as any);

    const handleClose = vi.fn();
    render(<OneTimePasswordDeleteDialog open={true} handleClose={handleClose} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateSuccess).toHaveBeenCalledOnce();
    expect(handleClose).toHaveBeenCalledOnce();
});

it("shows error notification when deletion fails with elevation required", async () => {
    const { deleteUserTOTPConfiguration } = await import("@services/UserInfoTOTPConfiguration");
    vi.mocked(deleteUserTOTPConfiguration).mockResolvedValue({ data: { elevation: true, status: "KO" } } as any);

    render(<OneTimePasswordDeleteDialog open={true} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateError).toHaveBeenCalledOnce();
});

it("shows error notification when deletion fails with authentication required", async () => {
    const { deleteUserTOTPConfiguration } = await import("@services/UserInfoTOTPConfiguration");
    vi.mocked(deleteUserTOTPConfiguration).mockResolvedValue({
        data: { authentication: true, status: "KO" },
    } as any);

    render(<OneTimePasswordDeleteDialog open={true} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateError).toHaveBeenCalledOnce();
});

it("shows generic error notification on other KO responses", async () => {
    const { deleteUserTOTPConfiguration } = await import("@services/UserInfoTOTPConfiguration");
    vi.mocked(deleteUserTOTPConfiguration).mockResolvedValue({ data: { status: "KO" } } as any);

    render(<OneTimePasswordDeleteDialog open={true} handleClose={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Remove"));
    });

    expect(mockCreateError).toHaveBeenCalledOnce();
});
