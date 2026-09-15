// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useState } from "react";

import axios from "axios";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";

import { Button } from "@components/UI/Button";
import { IndexRoute } from "@constants/Routes";
import { Confirm, RedirectionRestoreURL, RedirectionURL, State } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import MinimalLayout from "@layouts/MinimalLayout";
import { checkSafePostLogoutRedirection } from "@services/SafeRedirection";
import { signOut } from "@services/SignOut";

type Status = "confirming" | "signing-out";

const SignOut = function () {
    const { t: translate } = useTranslation();

    const { createErrorNotification } = useNotifications();
    const redirectionURL = useQueryParam(RedirectionURL);
    const redirector = useRedirector();
    const navigate = useRouterNavigate();
    const [query] = useSearchParams();

    const state = query.get(State);

    const [status, setStatus] = useState<Status>(query.get(Confirm) === "true" ? "confirming" : "signing-out");
    const [redirect, setRedirect] = useState(false);
    const [safeRedirect, setSafeRedirect] = useState(false);

    const handleRedirection = useCallback(() => {
        if (redirectionURL && safeRedirect) {
            const target = withState(redirectionURL, state);

            console.log("Redirecting to safe target URL: " + target);
            redirector(target);
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
    }, [redirectionURL, safeRedirect, state, query, redirector, navigate]);

    useEffect(() => {
        if (status !== "signing-out") return;

        const controller = new AbortController();
        let timeoutId: ReturnType<typeof setTimeout> | undefined;

        (async () => {
            try {
                const res = await signOut(redirectionURL, controller.signal);
                if (res?.safeTargetURL) {
                    setSafeRedirect(true);
                }
                timeoutId = setTimeout(() => {
                    setRedirect(true);
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
    }, [status, redirectionURL, createErrorNotification, translate]);

    useEffect(() => {
        if (redirect) {
            handleRedirection();
        }
    }, [redirect, handleRedirection]);

    const handleConfirm = useCallback(() => setStatus("signing-out"), []);

    const handleCancel = useCallback(async () => {
        if (redirectionURL) {
            try {
                const res = await checkSafePostLogoutRedirection(redirectionURL);

                if (res?.ok) {
                    setSafeRedirect(true);
                }
            } catch (err) {
                if (!axios.isCancel(err)) {
                    console.error(err);
                }
            }
        }

        setRedirect(true);
    }, [redirectionURL]);

    if (status === "confirming") {
        return (
            <MinimalLayout title={translate("Sign out")}>
                <p className="p-2">{translate("Are you sure you want to sign out?")}</p>
                <div className="flex flex-row justify-center gap-4 p-2">
                    <Button id="sign-out-confirm" color="primary" onClick={handleConfirm}>
                        {translate("Sign out")}
                    </Button>
                    <Button id="sign-out-cancel" color="primary" variant="outline" onClick={handleCancel}>
                        {translate("Cancel")}
                    </Button>
                </div>
            </MinimalLayout>
        );
    }

    return (
        <MinimalLayout title={translate("Sign out")}>
            <p className="p-2">{translate("You're being signed out and redirected")}...</p>
        </MinimalLayout>
    );
};

function withState(redirectionURL: string, state: null | string): string {
    if (!state) {
        return redirectionURL;
    }

    try {
        const url = new URL(redirectionURL);

        url.searchParams.set(State, state);

        return url.toString();
    } catch {
        const separator = redirectionURL.includes("?") ? "&" : "?";

        return `${redirectionURL}${separator}${State}=${encodeURIComponent(state)}`;
    }
}

export default SignOut;
