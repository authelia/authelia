import { act, render, screen } from "@testing-library/react";

import SignOut from "@views/LoginPortal/SignOut/SignOut";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-router-dom", () => ({
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

vi.mock("@constants/SearchParams", () => ({
    RedirectionRestoreURL: "rd_restore",
    RedirectionURL: "rd",
}));

const mockNavigate = vi.fn();
const mockCreateError = vi.fn();

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mockCreateError,
    }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => vi.fn(),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => <div data-testid="layout">{props.children}</div>,
}));

vi.mock("@services/SignOut", () => ({
    signOut: vi.fn().mockResolvedValue({ safeTargetURL: false }),
}));

beforeEach(() => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    mockNavigate.mockReset();
    mockCreateError.mockReset();
});

it("renders sign out message", async () => {
    await act(async () => {
        render(<SignOut />);
    });

    expect(screen.getByText(/You're being signed out and redirected/)).toBeInTheDocument();
});
