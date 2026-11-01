// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, render, screen } from "@testing-library/react";

import RevokeOneTimeCodeView from "@views/Revoke/RevokeOneTimeCodeView";

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
    useID: vi.fn(() => "test-id"),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

vi.mock("@services/UserSessionElevation", () => ({
    deleteUserSessionElevation: vi.fn(() => true),
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading" />,
}));

beforeEach(() => {
    mockCreateError.mockReset();
    mockCreateSuccess.mockReset();
    mockNavigate.mockReset();
});

it("renders loading page", async () => {
    await act(async () => {
        render(<RevokeOneTimeCodeView />);
    });
});

it("calls deleteUserSessionElevation with the id", async () => {
    const { deleteUserSessionElevation } = await import("@services/UserSessionElevation");

    await act(async () => {
        render(<RevokeOneTimeCodeView />);
    });

    expect(deleteUserSessionElevation).toHaveBeenCalledWith("test-id");
});

it("shows error when id is not provided", async () => {
    const { useID } = await import("@hooks/Revoke");
    vi.mocked(useID).mockReturnValueOnce(null as any);

    await act(async () => {
        render(<RevokeOneTimeCodeView />);
    });

    expect(mockCreateError).toHaveBeenCalledWith("The One-Time Code identifier was not provided");
});

it("reports a successful revocation", async () => {
    const { deleteUserSessionElevation } = await import("@services/UserSessionElevation");
    vi.mocked(deleteUserSessionElevation).mockResolvedValue(true as any);

    await act(async () => {
        render(<RevokeOneTimeCodeView />);
    });

    expect(mockCreateSuccess).toHaveBeenCalledWith("Successfully revoked the One-Time Code");
    expect(mockCreateError).not.toHaveBeenCalled();
});

it("reports a failed revocation", async () => {
    const { deleteUserSessionElevation } = await import("@services/UserSessionElevation");
    vi.mocked(deleteUserSessionElevation).mockResolvedValue(false as any);

    await act(async () => {
        render(<RevokeOneTimeCodeView />);
    });

    expect(mockCreateError).toHaveBeenCalledWith("Failed to revoke the One-Time Code");
    expect(mockCreateSuccess).not.toHaveBeenCalled();
});

it("navigates back to the index route after the delay", async () => {
    vi.useFakeTimers();

    const { deleteUserSessionElevation } = await import("@services/UserSessionElevation");
    vi.mocked(deleteUserSessionElevation).mockResolvedValue(true as any);

    render(<RevokeOneTimeCodeView />);

    await vi.waitFor(() => expect(mockCreateSuccess).toHaveBeenCalled());
    expect(mockNavigate).not.toHaveBeenCalled();

    await act(async () => {
        await vi.advanceTimersByTimeAsync(1500);
    });

    expect(mockNavigate).toHaveBeenCalledWith("/", false);

    vi.useRealTimers();
});

it("navigates back to the index route when the identifier is missing", async () => {
    vi.useFakeTimers();

    const { useID } = await import("@hooks/Revoke");
    vi.mocked(useID).mockReturnValue(null as any);

    render(<RevokeOneTimeCodeView />);

    await act(async () => {
        await vi.advanceTimersByTimeAsync(1500);
    });

    expect(mockNavigate).toHaveBeenCalledWith("/", false);

    vi.mocked(useID).mockReturnValue("test-id" as any);
    vi.useRealTimers();
});

it("renders the loading page while revoking", async () => {
    await act(async () => {
        render(<RevokeOneTimeCodeView />);
    });

    expect(screen.getByTestId("loading")).toBeInTheDocument();
});
