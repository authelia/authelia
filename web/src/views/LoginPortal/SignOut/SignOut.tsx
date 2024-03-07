import React, { useCallback, useEffect, useState } from "react";

import { Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { Navigate } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { RedirectionURL } from "@constants/SearchParams";
import { useIsMountedRef } from "@hooks/Mounted";
import { useNotifications } from "@hooks/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import MinimalLayout from "@layouts/MinimalLayout";
import { signOut } from "@services/SignOut";

export interface Props {}

const SignOut = function (props: Props) {
    const mounted = useIsMountedRef();
    const styles = useStyles();
    const { createErrorNotification } = useNotifications();
    const redirectionURL = useQueryParam(RedirectionURL);
    const redirector = useRedirector();
    const [timedOut, setTimedOut] = useState(false);
    const [safeRedirect, setSafeRedirect] = useState(false);
    const { t: translate } = useTranslation();

    const doSignOut = useCallback(async () => {
        try {
            const res = await signOut(redirectionURL);
            if (res !== undefined && res.safeTargetURL) {
                setSafeRedirect(true);
            }
            setTimeout(() => {
                if (!mounted) {
                    return;
                }
                setTimedOut(true);
            }, 2000);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("There was an issue signing out"));
        }
    }, [createErrorNotification, redirectionURL, setSafeRedirect, setTimedOut, mounted, translate]);

    useEffect(() => {
        doSignOut();
    }, [doSignOut]);

    if (timedOut) {
        if (redirectionURL && safeRedirect) {
            redirector(redirectionURL);
        } else {
            return <Navigate to={IndexRoute} />;
        }
    }

    return (
        <MinimalLayout title={translate("Sign out")}>
            <Typography className={styles.typo}>{translate("You're being signed out and redirected")}...</Typography>
        </MinimalLayout>
    );
};

export default SignOut;

const useStyles = makeStyles((theme: Theme) => ({
    typo: {
        padding: theme.spacing(),
    },
}));
