import React, { useEffect, useCallback, useState } from "react";

import { Typography, makeStyles } from "@material-ui/core";
import { Redirect } from "react-router";

import { useIsMountedRef } from "../../../hooks/Mounted";
import { useNotifications } from "../../../hooks/NotificationsContext";
import { useRedirectionURL } from "../../../hooks/RedirectionURL";
import LoginLayout from "../../../layouts/LoginLayout";
import { FirstFactorRoute } from "../../../Routes";
import { signOut } from "../../../services/SignOut";

export interface Props {}

const SignOut = function (props: Props) {
    const mounted = useIsMountedRef();
    const style = useStyles();
    const { createErrorNotification } = useNotifications();
    const redirectionURL = useRedirectionURL();
    const [timedOut, setTimedOut] = useState(false);

    const doSignOut = useCallback(async () => {
        try {
            // TODO(c.michaud): pass redirection URL to backend for validation.
            await signOut();
            setTimeout(() => {
                if (!mounted) {
                    return;
                }
                setTimedOut(true);
            }, 2000);
        } catch (err) {
            console.error(err);
            createErrorNotification("There was an issue signing out");
        }
    }, [createErrorNotification, setTimedOut, mounted]);

    useEffect(() => {
        doSignOut();
    }, [doSignOut]);

    if (timedOut) {
        if (redirectionURL) {
            window.location.href = redirectionURL;
        } else {
            return <Redirect to={FirstFactorRoute} />;
        }
    }

    return (
        <LoginLayout title="Sign out">
            <Typography className={style.typo}>You're being signed out and redirected...</Typography>
        </LoginLayout>
    );
};

export default SignOut;

const useStyles = makeStyles((theme) => ({
    typo: {
        padding: theme.spacing(),
    },
}));
