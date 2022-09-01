import React, { useEffect, useState } from "react";

import { Alert, Slide, SlideProps, Snackbar } from "@mui/material";

import { useNotifications } from "@hooks/NotificationsContext";
import { Notification } from "@models/Notifications";

export interface Props {
    onClose: () => void;
}

type NotificationBarTransitionProps = Omit<SlideProps, "direction">;

function NotificationBarTransition(props: NotificationBarTransitionProps) {
    return <Slide {...props} direction={"left"} />;
}

const NotificationBar = function (props: Props) {
    const [tmpNotification, setTmpNotification] = useState(null as Notification | null);
    const { notification } = useNotifications();

    useEffect(() => {
        if (notification) {
            setTmpNotification(notification);
        }
    }, [notification, setTmpNotification]);

    const shouldSnackbarBeOpen = notification !== undefined && notification !== null;

    return (
        <Snackbar
            open={shouldSnackbarBeOpen}
            anchorOrigin={{ vertical: "top", horizontal: "right" }}
            autoHideDuration={tmpNotification ? tmpNotification.timeout * 1000 : 10000}
            onClose={props.onClose}
            TransitionComponent={NotificationBarTransition}
            TransitionProps={{
                onExited: () => setTmpNotification(null),
            }}
        >
            {tmpNotification ? (
                <Alert severity={tmpNotification.level} variant={"filled"} elevation={6} className={"notification"}>
                    {tmpNotification.message}
                </Alert>
            ) : (
                <Alert severity={"success"} elevation={6} variant={"filled"} className={"notification"} />
            )}
        </Snackbar>
    );
};

export default NotificationBar;
