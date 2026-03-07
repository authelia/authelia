import { act, render, screen } from "@testing-library/react";

import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { success: "" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@components/OneTimeCodeTextField", () => ({
    default: (props: any) => <input data-testid="one-time-code" value={props.value} readOnly />,
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@services/UserSessionElevation", () => ({
    deleteUserSessionElevation: vi.fn(),
    generateUserSessionElevation: vi.fn().mockResolvedValue({ delete_id: "del-123" }),
    verifyUserSessionElevation: vi.fn(),
}));

const elevation = {
    can_skip_second_factor: false,
    elevated: false,
    require_second_factor: true,
    skip_second_factor: false,
} as any;

it("renders dialog with title when opening and elevation resolve", async () => {
    await act(async () => {
        render(
            <IdentityVerificationDialog
                elevation={elevation}
                opening={true}
                handleClosed={vi.fn()}
                handleOpened={vi.fn()}
            />,
        );
    });
    expect(screen.getByText("Identity Verification")).toBeInTheDocument();
});

it("renders cancel and verify buttons after elevation generation resolves", async () => {
    await act(async () => {
        render(
            <IdentityVerificationDialog
                elevation={elevation}
                opening={true}
                handleClosed={vi.fn()}
                handleOpened={vi.fn()}
            />,
        );
    });
    expect(screen.getByText("Cancel")).toBeInTheDocument();
    expect(screen.getByText("Verify")).toBeInTheDocument();
});

it("does not render content when not opening", () => {
    render(
        <IdentityVerificationDialog
            elevation={elevation}
            opening={false}
            handleClosed={vi.fn()}
            handleOpened={vi.fn()}
        />,
    );
    expect(screen.queryByText("Verify")).not.toBeInTheDocument();
});
