import { useEffect, useRef } from "react";

import { toast } from "sonner";

import { Toaster } from "@components/UI/Sonner";
import { useNotifications } from "@contexts/NotificationsContext";

const NotificationBar = function () {
    const { notification, resetNotification } = useNotifications();
    const prevNotificationRef = useRef(notification);

    useEffect(() => {
        if (notification && notification !== prevNotificationRef.current) {
            const toastFn =
                notification.level === "success"
                    ? toast.success
                    : notification.level === "error"
                      ? toast.error
                      : notification.level === "warning"
                        ? toast.warning
                        : toast.info;

            toastFn(notification.message, {
                className: "notification",
                duration: notification.timeout * 1000,
                onAutoClose: () => resetNotification(),
                onDismiss: () => resetNotification(),
            });
        }

        prevNotificationRef.current = notification;
    }, [notification, resetNotification]);

    return <Toaster position="top-right" richColors />;
};

export default NotificationBar;
