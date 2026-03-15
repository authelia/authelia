import { render } from "@testing-library/react";

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

it("displays notification message correctly", async () => {
    render(
        <NotificationsContext.Provider value={{ notification: testNotification, setNotification: () => {} }}>
            <NotificationBar onClose={() => {}} />
        </NotificationsContext.Provider>,
    );
});
