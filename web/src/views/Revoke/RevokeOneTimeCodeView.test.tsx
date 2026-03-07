import { act, render } from "@testing-library/react";

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

vi.mock("@hooks/NotificationsContext", () => ({
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
