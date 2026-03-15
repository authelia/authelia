import { useEffect, useRef } from "react";

import { toast } from "sonner";

import { Toaster } from "@components/UI/Sonner";
import { useNotifications } from "@hooks/NotificationsContext";

export interface Props {
    onClose: () => void;
}

const NotificationBar = function (props: Props) {
    const { notification } = useNotifications();
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
                duration: notification.timeout * 1000,
                onAutoClose: () => props.onClose(),
                onDismiss: () => props.onClose(),
            });
        }

        prevNotificationRef.current = notification;
    }, [notification, props]);

    return <Toaster position="top-right" richColors />;
};

export default NotificationBar;
