// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect } from "react";

import { useTranslation } from "react-i18next";

import ComponentOrLoading from "@components/ComponentOrLoading";
import FailureIcon from "@components/FailureIcon";
import LockIcon from "@components/LockIcon";
import LogoutButton from "@components/LogoutButton";
import SwitchUserButton from "@components/SwitchUserButton";
import { ErrorCode, ErrorCodeForbidden, RedirectionURL } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoGET } from "@hooks/UserInfo";
import MinimalLayout from "@layouts/MinimalLayout";
import { AuthenticationLevel } from "@services/State";

function getHost(url?: string) {
    if (!url) {
        return undefined;
    }

    try {
        const parsed = new URL(url);

        return parsed.protocol === "https:" || parsed.protocol === "http:" ? parsed.host : undefined;
    } catch {
        return undefined;
    }
}

const ErrorView = function () {
    const { t: translate } = useTranslation();
    const { createErrorNotification } = useNotifications();

    const forbidden = useQueryParam(ErrorCode) === ErrorCodeForbidden;
    const host = getHost(useQueryParam(RedirectionURL));

    const [state, fetchState, , fetchStateError] = useAutheliaState();
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoGET();

    const authenticated = state !== undefined && state.authentication_level >= AuthenticationLevel.OneFactor;

    useEffect(() => {
        fetchState();
    }, [fetchState]);

    useEffect(() => {
        if (authenticated) {
            fetchUserInfo();
        }
    }, [authenticated, fetchUserInfo]);

    useEffect(() => {
        if (fetchStateError) {
            createErrorNotification(translate("There was an issue retrieving the current user state"));
        }
    }, [fetchStateError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(translate("There was an issue retrieving user preferences"));
        }
    }, [fetchUserInfoError, createErrorNotification, translate]);

    const ready =
        fetchStateError !== undefined ||
        fetchUserInfoError !== undefined ||
        (state !== undefined && (!authenticated || userInfo !== undefined));

    let message = translate("An unexpected error occurred");

    if (forbidden) {
        message = host
            ? translate("You do not have permission to access {{host}}", { host })
            : translate("You do not have permission to access this resource");
    }

    return (
        <ComponentOrLoading ready={ready}>
            <MinimalLayout
                id={forbidden ? "access-denied-stage" : "error-stage"}
                title={forbidden ? translate("Access Denied") : null}
                userInfo={userInfo}
            >
                <div className="flex flex-col items-center justify-center">
                    {authenticated ? (
                        <div className="flex w-full justify-center gap-2">
                            <SwitchUserButton />
                            <LogoutButton />
                        </div>
                    ) : null}
                    <div className="my-4 flex w-full flex-col items-center gap-4 rounded-[10px] border border-border p-8">
                        {forbidden ? <LockIcon /> : <FailureIcon />}
                        <p className="break-words">{message}</p>
                    </div>
                </div>
            </MinimalLayout>
        </ComponentOrLoading>
    );
};

export default ErrorView;
