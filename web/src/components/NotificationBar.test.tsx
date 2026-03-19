import { render, screen } from "@testing-library/react";

import NotificationBar from "@components/NotificationBar";
import { NotificationsContext, NotificationsContextValue } from "@contexts/NotificationsContext";
import { Notification } from "@models/Notifications";

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
        <NotificationsContext.Provider value={baseContextValue}>
            <NotificationBar />
        </NotificationsContext.Provider>,
    );
});

it("displays notification message and level correctly", async () => {
    render(
        <NotificationsContext.Provider value={{ ...baseContextValue, isActive: true, notification: testNotification }}>
            <NotificationBar />
        </NotificationsContext.Provider>,
    );

    const alert = screen.getByRole("alert");
    const message = await screen.findByText(testNotification.message);

    expect(alert).toHaveClass(
        `MuiAlert-filled${testNotification.level.charAt(0).toUpperCase() + testNotification.level.substring(1)}`,
        { exact: false },
    );
    expect(message).toHaveTextContent(testNotification.message);
});
