import { useEffect, useRef } from "react";

import { Toast as ToastPrimitive } from "@base-ui/react/toast";

import { ToastProvider, Toaster } from "@components/UI/Toast";
import { useNotifications } from "@contexts/NotificationsContext";

const NotificationToasts = function () {
    const { notification, resetNotification } = useNotifications();
    const manager = ToastPrimitive.useToastManager();
    const prevNotificationRef = useRef(notification);

    useEffect(() => {
        if (notification && notification !== prevNotificationRef.current) {
            const current = notification;

            manager.add({
                onClose: () => {
                    if (prevNotificationRef.current === current) resetNotification();
                },
                timeout: notification.timeout * 1000,
                title: notification.message,
                type: notification.level,
            });
        }

        prevNotificationRef.current = notification;
    }, [notification, resetNotification, manager]);

    return <Toaster />;
};

const NotificationBar = function () {
    return (
        <ToastProvider>
            <NotificationToasts />
        </ToastProvider>
    );
};

export default NotificationBar;
