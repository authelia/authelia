import { useCallback, createContext, useContext } from "react";

import { Level } from "../components/ColoredSnackbarContent";
import { Notification } from "../models/Notifications";

const defaultOptions = {
    timeout: 5,
};

interface NotificationContextProps {
    notification: Notification | null;
    setNotification: (n: Notification | null) => void;
}

const NotificationsContext = createContext<NotificationContextProps>({ notification: null, setNotification: () => {} });

export default NotificationsContext;

export function useNotifications() {
    let useNotificationsProps = useContext(NotificationsContext);

    const notificationBuilder = (level: Level) => {
        return (message: string, timeout?: number) => {
            useNotificationsProps.setNotification({
                level,
                message,
                timeout: timeout ? timeout : defaultOptions.timeout,
            });
        };
    };

    const resetNotification = () => useNotificationsProps.setNotification(null);
    /* eslint-disable react-hooks/exhaustive-deps */
    const createInfoNotification = useCallback(notificationBuilder("info"), []);
    const createSuccessNotification = useCallback(notificationBuilder("success"), []);
    const createWarnNotification = useCallback(notificationBuilder("warning"), []);
    const createErrorNotification = useCallback(notificationBuilder("error"), []);
    /* eslint-enable react-hooks/exhaustive-deps */
    const isActive = useNotificationsProps.notification !== null;

    return {
        notification: useNotificationsProps.notification,
        resetNotification,
        createInfoNotification,
        createSuccessNotification,
        createWarnNotification,
        createErrorNotification,
        isActive,
    };
}
