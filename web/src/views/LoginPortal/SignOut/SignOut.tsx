import { useCallback, useEffect, useState } from "react";

import { Typography } from "@mui/material";
import axios from "axios";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { RedirectionRestoreURL, RedirectionURL } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import MinimalLayout from "@layouts/MinimalLayout";
import { signOut } from "@services/SignOut";

const SignOut = function () {
    const { t: translate } = useTranslation();

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
        const controller = new AbortController();
        let timeoutId: ReturnType<typeof setTimeout> | undefined;

        (async () => {
            try {
                const res = await signOut(redirectionURL, controller.signal);
                if (res?.safeTargetURL) {
                    setSafeRedirect(true);
                }
                timeoutId = setTimeout(() => {
                    setTimedOut(true);
                }, 2000);
            } catch (err) {
                if (axios.isCancel(err)) return;
                console.error(err);
                createErrorNotification(translate("There was an issue signing out"));
            }
        })();

        return () => {
            controller.abort();
            if (timeoutId !== undefined) {
                clearTimeout(timeoutId);
            }
        };
    }, [redirectionURL, createErrorNotification, translate]);

    useEffect(() => {
        if (timedOut) {
            handleRedirection();
        }
    }, [timedOut, handleRedirection]);

    return (
        <MinimalLayout title={translate("Sign out")}>
            <Typography sx={{ padding: (theme) => theme.spacing() }}>
                {translate("You're being signed out and redirected")}...
            </Typography>
        </MinimalLayout>
    );
};

export default SignOut;
