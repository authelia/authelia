import { act, fireEvent, render, screen } from "@testing-library/react";

import PasswordForm from "@views/LoginPortal/SecondFactor/PasswordForm";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { actionRow: "", form: "" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
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

vi.mock("@services/Password", () => ({
    postSecondFactor: vi.fn(),
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: () => null,
}));

it("renders the password form", () => {
    render(<PasswordForm onAuthenticationSuccess={vi.fn()} />);
    expect(screen.getByText("Password")).toBeInTheDocument();
    expect(screen.getByText("Authenticate")).toBeInTheDocument();
});

it("shows error when submitting empty password", async () => {
    render(<PasswordForm onAuthenticationSuccess={vi.fn()} />);

    await act(async () => {
        fireEvent.click(screen.getByText("Authenticate"));
    });
});
