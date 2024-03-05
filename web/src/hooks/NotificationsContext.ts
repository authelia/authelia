import { createContext, useContext, useRef } from "react";

import { AlertColor } from "@mui/material";

import { Notification } from "@models/Notifications";

const defaultOptions = {
    timeout: 5,
};

interface NotificationContextProps {
    notification: Notification | null;
    setNotification: (n: Notification | null) => void;
}

const NotificationsContext = createContext<NotificationContextProps>({ notification: null, setNotification: () => {} });

export function useNotifications() {
    let useNotificationsProps = useContext(NotificationsContext);

    const notificationBuilder = (level: AlertColor) => {
        return (message: string, timeout?: number) => {
            useNotificationsProps.setNotification({
                level,
                message,
                timeout: timeout ? timeout : defaultOptions.timeout,
            });
        };
    };

    const resetNotification = () => useNotificationsProps.setNotification(null);
    const createInfoNotification = useRef(notificationBuilder("info")).current;
    const createSuccessNotification = useRef(notificationBuilder("success")).current;
    const createWarnNotification = useRef(notificationBuilder("warning")).current;
    const createErrorNotification = useRef(notificationBuilder("error")).current;
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

export default NotificationsContext;
