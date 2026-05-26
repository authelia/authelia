import { useCallback, useEffect, useState } from "react";

import { Button, Stack, Typography } from "@mui/material";
import axios from "axios";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { Confirm, RedirectionRestoreURL, RedirectionURL, State } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import MinimalLayout from "@layouts/MinimalLayout";
import { checkSafeRedirection } from "@services/SafeRedirection";
import { signOut } from "@services/SignOut";

type Status = "checking" | "confirming" | "signing-out";

const SignOut = function () {
    const { t: translate } = useTranslation();

    const { createErrorNotification } = useNotifications();
    const redirectionURL = useQueryParam(RedirectionURL);
    const redirector = useRedirector();
    const navigate = useRouterNavigate();
    const [query] = useSearchParams();

    const confirmRequired = query.get(Confirm) === "true";
    const state = query.get(State);

    const [status, setStatus] = useState<Status>("checking");
    const [safeRedirect, setSafeRedirect] = useState(false);

    const handleRedirection = useCallback(() => {
        if (redirectionURL && safeRedirect) {
            let target = redirectionURL;

            if (state) {
                try {
                    const url = new URL(redirectionURL);
                    url.searchParams.set(State, state);
                    target = url.toString();
                } catch {
                    const sep = redirectionURL.includes("?") ? "&" : "?";
                    target = `${redirectionURL}${sep}${State}=${encodeURIComponent(state)}`;
                }
            }

            console.log("Redirecting to safe target URL: " + target);
            redirector(target);

            return;
        }

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
    }, [redirectionURL, safeRedirect, state, query, redirector, navigate]);

    useEffect(() => {
        const controller = new AbortController();

        (async () => {
            let safe = false;

            if (redirectionURL) {
                try {
                    const res = await checkSafeRedirection(redirectionURL);
                    safe = !!res?.ok;
                } catch (err) {
                    if (axios.isCancel(err)) return;
                    console.error(err);
                }
            }

            if (controller.signal.aborted) return;

            setSafeRedirect(safe);
            setStatus(confirmRequired ? "confirming" : "signing-out");
        })();

        return () => {
            controller.abort();
        };
    }, [redirectionURL, confirmRequired]);

    useEffect(() => {
        if (status !== "signing-out") return;

        const controller = new AbortController();
        let timeoutId: ReturnType<typeof setTimeout> | undefined;

        (async () => {
            try {
                await signOut(redirectionURL, controller.signal);
                timeoutId = setTimeout(() => handleRedirection(), 2000);
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
    }, [status, redirectionURL, handleRedirection, createErrorNotification, translate]);

    const handleConfirm = () => setStatus("signing-out");
    const handleCancel = () => handleRedirection();

    if (status === "confirming") {
        return (
            <MinimalLayout title={translate("Sign out")}>
                <Typography sx={{ padding: (theme) => theme.spacing() }}>
                    {translate("Are you sure you want to sign out?")}
                </Typography>
                <Stack
                    direction="row"
                    spacing={2}
                    sx={{ justifyContent: "center", padding: (theme) => theme.spacing() }}
                >
                    <Button id="sign-out-confirm" variant="contained" color="primary" onClick={handleConfirm}>
                        {translate("Yes")}
                    </Button>
                    <Button id="sign-out-cancel" variant="outlined" color="primary" onClick={handleCancel}>
                        {translate("No")}
                    </Button>
                </Stack>
            </MinimalLayout>
        );
    }

    return (
        <MinimalLayout title={translate("Sign out")}>
            <Typography sx={{ padding: (theme) => theme.spacing() }}>
                {translate("You're being signed out and redirected")}...
            </Typography>
        </MinimalLayout>
    );
};

export default SignOut;
