import React, { useState, useEffect } from "react";
import { Snackbar } from "@material-ui/core";
import ColoredSnackbarContent from "./ColoredSnackbarContent";
import { useNotifications } from "../hooks/NotificationsContext";
import { Notification } from "../models/Notifications";

export interface Props {
    onClose: () => void;
}

export default function (props: Props) {
    const [tmpNotification, setTmpNotification] = useState(null as Notification | null);
    const { notification } = useNotifications();

    useEffect(() => {
        if (notification && notification !== null) {
            setTmpNotification(notification);
        }
    }, [notification]);

    const shouldSnackbarBeOpen = notification !== undefined && notification !== null;

    return (
        <Snackbar
            open={shouldSnackbarBeOpen}
            anchorOrigin={{ vertical: "top", horizontal: "right" }}
            autoHideDuration={tmpNotification ? tmpNotification.timeout * 1000 : 10000}
            onClose={props.onClose}
            onExited={() => setTmpNotification(null)}>
            <ColoredSnackbarContent
                className="notification"
                level={tmpNotification ? tmpNotification.level : "info"}
                message={tmpNotification ? tmpNotification.message : ""} />
        </Snackbar>
    )
}