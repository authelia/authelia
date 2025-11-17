import React, { useCallback, useEffect, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { useToken } from "@hooks/Revoke";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { deleteResetPasswordToken } from "@services/ResetPassword";
import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";

const RevokeResetPasswordTokenView = function () {
    const { t: translate } = useTranslation();
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const token = useToken();
    const navigate = useRouterNavigate();

    const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const [statusMessage, setStatusMessage] = useState(translate("Revoking reset password token..."));

    useEffect(() => {
        return () => {
            if (timeoutRef.current !== null) {
                clearTimeout(timeoutRef.current);
                timeoutRef.current = null;
            }
        };
    }, []);

    const handleRedirect = useCallback(() => {
        if (timeoutRef.current !== null) {
            clearTimeout(timeoutRef.current);
        }
        timeoutRef.current = setTimeout(() => {
            navigate(IndexRoute, false);
            timeoutRef.current = null;
        }, 1500);
    }, [navigate]);

    const handleRevoke = useCallback(async () => {
        if (!token) {
            return;
        }

        try {
            const { ok, status } = await deleteResetPasswordToken(token);

            if (ok) {
                createSuccessNotification(translate("Successfully revoked the Token"));
            } else if (status === 429) {
                createErrorNotification(translate("You have made too many requests"));
            } else {
                createErrorNotification(translate("Failed to revoke the Token"));
            }
        } catch (error) {
            console.error(error);
            createErrorNotification(translate("An unexpected error occurred while revoking the token"));
        } finally {
            setStatusMessage(translate("Redirecting..."));
            handleRedirect();
        }
    }, [createErrorNotification, createSuccessNotification, handleRedirect, token, translate]);

    useEffect(() => {
        if (!token) {
            setStatusMessage(translate("The Token was not provided"));
            createErrorNotification(translate("The Token was not provided"));

            handleRedirect();

            return;
        }

        handleRevoke();
    }, [createErrorNotification, handleRedirect, handleRevoke, token, translate]);

    return <BaseLoadingPage message={statusMessage} />;
};

export default RevokeResetPasswordTokenView;
