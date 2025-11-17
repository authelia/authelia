import React from "react";

import { Alert, Slide, SlideProps, Snackbar } from "@mui/material";

import { useNotifications } from "@hooks/NotificationsContext";

export interface Props {
    onClose: () => void;
}

type NotificationBarTransitionProps = Omit<SlideProps, "direction">;

function NotificationBarTransition(props: Readonly<NotificationBarTransitionProps>) {
    return <Slide {...props} direction={"left"} />;
}

const NotificationBar = function (props: Props) {
    const { notification } = useNotifications();

    const shouldSnackbarBeOpen = notification !== undefined && notification !== null;

    return (
        <Snackbar
            open={shouldSnackbarBeOpen}
            anchorOrigin={{ horizontal: "right", vertical: "top" }}
            autoHideDuration={notification ? notification.timeout * 1000 : 10000}
            onClose={props.onClose}
            slots={{ transition: NotificationBarTransition }}
        >
            {notification ? (
                <Alert severity={notification.level} variant={"filled"} elevation={6} className={"notification"}>
                    {notification.message}
                </Alert>
            ) : undefined}
        </Snackbar>
    );
};

export default NotificationBar;
