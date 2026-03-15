import { ReactNode, createContext, useCallback, useContext, useMemo, useState } from "react";

import NotificationBar from "@components/NotificationBar";
import type { Notification, NotificationLevel } from "@models/Notifications";

const DEFAULT_TIMEOUT_SECONDS = 5;

export interface NotificationsContextValue {
    createErrorNotification: (message: string, timeout?: number) => void;
    createInfoNotification: (message: string, timeout?: number) => void;
    createSuccessNotification: (message: string, timeout?: number) => void;
    createWarnNotification: (message: string, timeout?: number) => void;
    isActive: boolean;
    notification: Notification | null;
    resetNotification: () => void;
    showNotification: (level: NotificationLevel, message: string, timeout?: number) => void;
}

export const NotificationsContext = createContext<NotificationsContextValue | null>(null);

interface Props {
    children: ReactNode;
}

export default function NotificationsContextProvider(props: Props) {
    const [notification, setNotification] = useState<Notification | null>(null);

    const showNotification = useCallback(
        (level: NotificationLevel, message: string, timeout = DEFAULT_TIMEOUT_SECONDS) => {
            setNotification({
                level,
                message,
                timeout,
            });
        },
        [],
    );

    const resetNotification = useCallback(() => {
        setNotification(null);
    }, []);

    const createInfoNotification = useCallback(
        (message: string, timeout?: number) => {
            showNotification("info", message, timeout);
        },
        [showNotification],
    );

    const createSuccessNotification = useCallback(
        (message: string, timeout?: number) => {
            showNotification("success", message, timeout);
        },
        [showNotification],
    );

    const createWarnNotification = useCallback(
        (message: string, timeout?: number) => {
            showNotification("warning", message, timeout);
        },
        [showNotification],
    );

    const createErrorNotification = useCallback(
        (message: string, timeout?: number) => {
            showNotification("error", message, timeout);
        },
        [showNotification],
    );

    const value = useMemo<NotificationsContextValue>(
        () => ({
            createErrorNotification,
            createInfoNotification,
            createSuccessNotification,
            createWarnNotification,
            isActive: notification !== null,
            notification,
            resetNotification,
            showNotification,
        }),
        [
            notification,
            showNotification,
            createErrorNotification,
            createInfoNotification,
            createSuccessNotification,
            createWarnNotification,
            resetNotification,
        ],
    );

    return (
        <NotificationsContext.Provider value={value}>
            <NotificationBar />
            {props.children}
        </NotificationsContext.Provider>
    );
}

export function useNotifications(): NotificationsContextValue {
    const context = useContext(NotificationsContext);

    if (context === null) {
        throw new Error("useNotifications must be used within a NotificationsProvider");
    }

    return context;
}
