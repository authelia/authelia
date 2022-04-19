import React, { useState, useEffect } from "react";

import { Alert, Snackbar } from "@mui/material";

import { useNotifications } from "@hooks/NotificationsContext";
import { Notification } from "@models/Notifications";

export interface Props {
    onClose: () => void;
}

const NotificationBar = function (props: Props) {
    const [tmpNotification, setTmpNotification] = useState(null as Notification | null);
    const { notification } = useNotifications();

    useEffect(() => {
        if (notification && notification !== null) {
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
            TransitionProps={{
                onExited: () => setTmpNotification(null),
            }}
        >
            <Alert severity={tmpNotification ? tmpNotification.level : "success"}>
                {tmpNotification ? tmpNotification.message : ""}
            </Alert>
        </Snackbar>
    );
};

export default NotificationBar;
