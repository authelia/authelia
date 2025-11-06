import React, { useCallback, useEffect, useRef } from "react";

import { useTranslation } from "react-i18next";

import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { useID } from "@hooks/Revoke";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { deleteUserSessionElevation } from "@services/UserSessionElevation";
import LoadingPage from "@views/LoadingPage/LoadingPage";

const RevokeOneTimeCodeView = function () {
    const { t: translate } = useTranslation();
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const id = useID();
    const navigate = useRouterNavigate();

    const timeoutRef = useRef<NodeJS.Timeout | null>(null);

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
        if (!id) return;

        const ok = await deleteUserSessionElevation(id);

        if (ok) {
            createSuccessNotification(translate("Successfully revoked the One-Time Code"));
        } else {
            createErrorNotification(translate("Failed to revoke the One-Time Code"));
        }

        handleRedirect();
    }, [createErrorNotification, createSuccessNotification, handleRedirect, id, translate]);

    useEffect(() => {
        if (!id) {
            createErrorNotification(translate("The One-Time Code identifier was not provided"));

            handleRedirect();

            return;
        }

        handleRevoke().catch(console.error);
    }, [createErrorNotification, handleRedirect, handleRevoke, id, translate]);

    return <LoadingPage />;
};

export default RevokeOneTimeCodeView;
