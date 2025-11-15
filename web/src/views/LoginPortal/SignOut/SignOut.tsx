import React, { useCallback, useEffect, useRef, useState } from "react";

import { Theme, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { makeStyles } from "tss-react/mui";

import { IndexRoute } from "@constants/Routes";
import { RedirectionRestoreURL, RedirectionURL } from "@constants/SearchParams";
import { useIsMountedRef } from "@hooks/Mounted";
import { useNotifications } from "@hooks/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import MinimalLayout from "@layouts/MinimalLayout";
import { signOut } from "@services/SignOut";

const SignOut = function () {
    const { t: translate } = useTranslation();
    const { classes } = useStyles();

    const mounted = useIsMountedRef();
    const { createErrorNotification } = useNotifications();
    const redirectionURL = useQueryParam(RedirectionURL);
    const redirector = useRedirector();
    const navigate = useRouterNavigate();
    const [timedOut, setTimedOut] = useState(false);
    const [safeRedirect, setSafeRedirect] = useState(false);
    const [query] = useSearchParams();
    const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const doSignOut = useCallback(async () => {
        try {
            const res = await signOut(redirectionURL);
            if (res !== undefined && res.safeTargetURL) {
                setSafeRedirect(true);
            }
            timeoutRef.current = setTimeout(() => {
                if (!mounted.current) {
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

        return () => {
            if (timeoutRef.current !== null) {
                clearTimeout(timeoutRef.current);
                timeoutRef.current = null;
            }
        };
    }, [doSignOut]);

    if (timedOut) {
        if (redirectionURL && safeRedirect) {
            redirector(redirectionURL);
        } else {
            if (query.has(RedirectionRestoreURL)) {
                const search = new URLSearchParams();

                query.forEach((value, key) => {
                    if (key !== RedirectionRestoreURL) {
                        search.set(key, value);
                    } else {
                        search.set(RedirectionURL, value);
                    }
                });

                navigate(IndexRoute, false, false, false, search);
            } else {
                navigate(IndexRoute);
            }
        }
    }

    return (
        <MinimalLayout title={translate("Sign out")}>
            <Typography className={classes.typo}>{translate("You're being signed out and redirected")}...</Typography>
        </MinimalLayout>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    typo: {
        padding: theme.spacing(),
    },
}));

export default SignOut;
