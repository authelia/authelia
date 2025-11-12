import { createContext, useCallback, useContext } from "react";

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
    const { notification, setNotification } = useContext(NotificationsContext);

    const createNotification = useCallback(
        (level: "error" | "info" | "success" | "warning", message: string, timeout?: number) => {
            setNotification({
                level,
                message,
                timeout: timeout ?? defaultOptions.timeout,
            });
        },
        [setNotification],
    );

    const resetNotification = useCallback(() => setNotification(null), [setNotification]);
    const createInfoNotification = useCallback(
        (message: string, timeout?: number) => createNotification("info", message, timeout),
        [createNotification],
    );
    const createSuccessNotification = useCallback(
        (message: string, timeout?: number) => createNotification("success", message, timeout),
        [createNotification],
    );
    const createWarnNotification = useCallback(
        (message: string, timeout?: number) => createNotification("warning", message, timeout),
        [createNotification],
    );
    const createErrorNotification = useCallback(
        (message: string, timeout?: number) => createNotification("error", message, timeout),
        [createNotification],
    );
    const isActive = notification !== null;

    return {
        createErrorNotification,
        createInfoNotification,
        createSuccessNotification,
        createWarnNotification,
        isActive,
        notification,
        resetNotification,
    };
}

export default NotificationsContext;
