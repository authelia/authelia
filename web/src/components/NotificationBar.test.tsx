import React from "react";

import { render, screen } from "@testing-library/react";

import NotificationBar from "@components/NotificationBar";
import NotificationsContext from "@hooks/NotificationsContext";
import { Notification } from "@models/Notifications";

const testNotification: Notification = {
    level: "success",
    message: "Test notification",
    timeout: 3,
};

it("renders without crashing", () => {
    render(<NotificationBar onClose={() => {}} />);
});

it("displays notification message and level correctly", async () => {
    render(
        <NotificationsContext.Provider value={{ notification: testNotification, setNotification: () => {} }}>
            <NotificationBar onClose={() => {}} />
        </NotificationsContext.Provider>,
    );

    const alert = await screen.getByRole("alert");
    const message = await screen.findByText(testNotification.message);

    expect(alert).toHaveClass(
        `MuiAlert-filled${testNotification.level.charAt(0).toUpperCase() + testNotification.level.substring(1)}`,
        { exact: false },
    );
    expect(message).toHaveTextContent(testNotification.message);
});
