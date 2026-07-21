import { render, screen } from "@testing-library/react";

import NotificationBar from "@components/NotificationBar";
import { NotificationsContext, NotificationsContextValue } from "@contexts/NotificationsContext";
import { Notification } from "@models/Notifications";

vi.mock("@mui/material/Slide", () => ({
    default: ({ children, in: isIn }: { children: React.ReactElement; in?: boolean }) => (isIn ? children : children),
}));

const testNotification: Notification = {
    level: "success",
    message: "Test notification",
    timeout: 3,
};

const baseContextValue: NotificationsContextValue = {
    createErrorNotification: vi.fn(),
    createInfoNotification: vi.fn(),
    createSuccessNotification: vi.fn(),
    createWarnNotification: vi.fn(),
    isActive: false,
    notification: null,
    resetNotification: vi.fn(),
    showNotification: vi.fn(),
};

it("renders without crashing", () => {
    render(
        <NotificationsContext value={baseContextValue}>
            <NotificationBar />
        </NotificationsContext>,
    );
});

it("displays notification message and level correctly", async () => {
    render(
        <NotificationsContext value={{ ...baseContextValue, isActive: true, notification: testNotification }}>
            <NotificationBar />
        </NotificationsContext>,
    );

    const alert = screen.getByRole("alert");
    const message = await screen.findByText(testNotification.message);

    expect(alert).toHaveClass(
        `MuiAlert-filled${testNotification.level.charAt(0).toUpperCase() + testNotification.level.substring(1)}`,
        { exact: false },
    );
    expect(message).toHaveTextContent(testNotification.message);
});

it("retains notification styling during close transition", () => {
    const { rerender } = render(
        <NotificationsContext value={{ ...baseContextValue, isActive: true, notification: testNotification }}>
            <NotificationBar />
        </NotificationsContext>,
    );

    expect(screen.getByRole("alert")).toHaveClass("MuiAlert-filledSuccess", { exact: false });
    expect(screen.getByText(testNotification.message)).toBeInTheDocument();

    rerender(
        <NotificationsContext value={{ ...baseContextValue, isActive: false, notification: null }}>
            <NotificationBar />
        </NotificationsContext>,
    );

    const alert = screen.getByRole("alert");

    expect(alert).toHaveClass("MuiAlert-filledSuccess", { exact: false });
    expect(screen.getByText(testNotification.message)).toBeInTheDocument();
});
