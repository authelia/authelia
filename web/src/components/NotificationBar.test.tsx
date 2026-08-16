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
        <NotificationsContext value={baseContextValue}>
            <NotificationBar />
        </NotificationsContext>,
    );
});

it("displays notification message correctly", async () => {
    const { rerender } = render(
        <NotificationsContext value={baseContextValue}>
            <NotificationBar />
        </NotificationsContext>,
    );

    rerender(
        <NotificationsContext value={{ ...baseContextValue, isActive: true, notification: testNotification }}>
            <NotificationBar />
        </NotificationsContext>,
    );

    expect(await screen.findByText(testNotification.message)).toBeInTheDocument();
});
