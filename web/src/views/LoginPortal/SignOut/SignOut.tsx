import { useCallback, useEffect, useState } from "react";

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

    const handleRedirection = useCallback(() => {
        if (redirectionURL && safeRedirect) {
            console.log("Redirecting to safe target URL: " + redirectionURL);
            redirector(redirectionURL);
        } else {
            console.log("Redirecting to index route");

            if (query.has(RedirectionRestoreURL)) {
                const search = new URLSearchParams();

                for (const [key, value] of query) {
                    if (key === RedirectionRestoreURL) {
                        search.set(RedirectionURL, value);
                    } else {
                        search.set(key, value);
                    }
                }

                navigate(IndexRoute, false, false, false, search);
            } else {
                navigate(IndexRoute);
            }
        }
    }, [redirectionURL, safeRedirect, query, redirector, navigate]);

    useEffect(() => {
        const performSignOut = async () => {
            try {
                const res = await signOut(redirectionURL);
                if (!mounted.current) {
                    return;
                }
                if (res?.safeTargetURL) {
                    setSafeRedirect(true);
                }
                setTimeout(() => {
                    if (!mounted.current) {
                        return;
                    }
                    setTimedOut(true);
                }, 2000);
            } catch (err) {
                console.error(err);
                createErrorNotification(translate("There was an issue signing out"));
            }
        };

        performSignOut();
    }, [redirectionURL, mounted, createErrorNotification, translate]);

    useEffect(() => {
        if (timedOut) {
            handleRedirection();
        }
    }, [timedOut, handleRedirection]);

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
