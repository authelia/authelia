import { act, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import ConsentPortal from "@views/ConsentPortal/ConsentPortal";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@constants/Routes", () => ({
    ConsentCompletionSubRoute: "/completion",
    ConsentOpenIDSubRoute: "/openid",
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
        resetNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/State", () => ({
    useAutheliaState: () => [null, vi.fn(), false, null],
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoGET: () => [undefined, vi.fn(), false, null],
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

it("renders loading page when state is null", async () => {
    await act(async () => {
        render(
            <MemoryRouter>
                <ConsentPortal />
            </MemoryRouter>,
        );
    });

    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});
