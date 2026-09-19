// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useState } from "react";

import axios from "axios";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";

import { Button } from "@components/UI/Button";
import { IndexRoute } from "@constants/Routes";
import { FlowID, RedirectionRestoreURL, RedirectionURL } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import MinimalLayout from "@layouts/MinimalLayout";
import { cancelSignOut, getSignOutPending, signOut } from "@services/SignOut";

type Status = "checking" | "confirming" | "signing-out";

const SignOut = function () {
    const { t: translate } = useTranslation();

    const { createErrorNotification } = useNotifications();
    const redirectionURL = useQueryParam(RedirectionURL);
    const redirector = useRedirector();
    const navigate = useRouterNavigate();
    const [query] = useSearchParams();

    const flowID = query.get(FlowID);

    const [status, setStatus] = useState<Status>(flowID ? "checking" : "signing-out");
    const [confirmed, setConfirmed] = useState(false);
    const [client, setClient] = useState<string>();
    const [redirect, setRedirect] = useState(false);
    const [target, setTarget] = useState<string>();

    const handleRedirection = useCallback(() => {
        if (target) {
            console.log("Redirecting to target URL: " + target);
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
    }, [target, query, redirector, navigate]);

    useEffect(() => {
        if (status !== "checking" || !flowID) return;

        const controller = new AbortController();

        (async () => {
            try {
                const res = await getSignOutPending(flowID, controller.signal);

                if (res.pending) {
                    setClient(res.clientName || res.clientID);
                    setStatus("confirming");
                } else {
                    setStatus("signing-out");
                }
            } catch (err) {
                if (axios.isCancel(err)) return;
                console.error(err);

                // Without knowing, the user is asked rather than risk logging them out on behalf of a Relying Party.
                setStatus("confirming");
            }
        })();

        return () => {
            controller.abort();
        };
    }, [status, flowID]);

    useEffect(() => {
        if (status !== "signing-out") return;

        const controller = new AbortController();
        let timeoutId: ReturnType<typeof setTimeout> | undefined;

        (async () => {
            try {
                const res =
                    confirmed && flowID
                        ? await signOut(redirectionURL, controller.signal, flowID)
                        : await signOut(redirectionURL, controller.signal);

                if (res?.redirectURL) {
                    setTarget(res.redirectURL);
                } else if (res?.safeTargetURL && redirectionURL) {
                    setTarget(redirectionURL);
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
    }, [status, confirmed, flowID, redirectionURL, createErrorNotification, translate]);

    useEffect(() => {
        if (redirect) {
            handleRedirection();
        }
    }, [redirect, handleRedirection]);

    const handleConfirm = useCallback(() => {
        setConfirmed(true);
        setStatus("signing-out");
    }, []);

    const handleCancel = useCallback(async () => {
        try {
            if (flowID) {
                await cancelSignOut(flowID);
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("There was an issue cancelling the sign out"));
        }

        navigate(IndexRoute);
    }, [flowID, createErrorNotification, translate, navigate]);

    if (status === "checking") {
        return <MinimalLayout title={translate("Sign out")} />;
    }

    if (status === "confirming") {
        return (
            <MinimalLayout title={translate("Sign out")}>
                {client ? (
                    <p className="p-2">{translate("{{client}} has requested that you sign out", { client })}</p>
                ) : null}
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

export default SignOut;
