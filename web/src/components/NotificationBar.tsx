import { useCallback, useState } from "react";

import { Alert, Slide, SlideProps, Snackbar } from "@mui/material";

import { useNotifications } from "@contexts/NotificationsContext";
import { Notification } from "@models/Notifications";

type NotificationBarTransitionProps = Omit<SlideProps, "direction">;

function NotificationBarTransition(props: Readonly<NotificationBarTransitionProps>) {
    return <Slide {...props} direction={"left"} />;
}

const NotificationBar = function () {
    const { notification, resetNotification } = useNotifications();
    const [lastNotification, setLastNotification] = useState<Notification | null>(null);

    if (notification !== null && notification !== lastNotification) {
        setLastNotification(notification);
    }

    const handleExited = useCallback(() => {
        setLastNotification(null);
    }, []);

    const open = notification !== null;
    const displayed = notification ?? lastNotification;

    return (
        <Snackbar
            open={open}
            anchorOrigin={{ horizontal: "right", vertical: "top" }}
            autoHideDuration={displayed ? displayed.timeout * 1000 : 10000}
            onClose={resetNotification}
            slots={{ transition: NotificationBarTransition }}
            slotProps={{ transition: { onExited: handleExited } }}
        >
            {displayed ? (
                <Alert severity={displayed.level} variant={"filled"} elevation={6} className={"notification"}>
                    {displayed.message}
                </Alert>
            ) : undefined}
        </Snackbar>
    );
};

export default NotificationBar;
