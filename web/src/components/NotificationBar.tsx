import { Alert, Slide, SlideProps, Snackbar } from "@mui/material";

import { useNotifications } from "@contexts/NotificationsContext";

type NotificationBarTransitionProps = Omit<SlideProps, "direction">;

function NotificationBarTransition(props: Readonly<NotificationBarTransitionProps>) {
    return <Slide {...props} direction={"left"} />;
}

const NotificationBar = function () {
    const { notification, resetNotification } = useNotifications();

    const shouldSnackbarBeOpen = notification !== undefined && notification !== null;

    return (
        <Snackbar
            open={shouldSnackbarBeOpen}
            anchorOrigin={{ horizontal: "right", vertical: "top" }}
            autoHideDuration={notification ? notification.timeout * 1000 : 10000}
            onClose={resetNotification}
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
