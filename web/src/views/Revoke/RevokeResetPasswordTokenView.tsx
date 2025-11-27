import { useCallback, useEffect } from "react";

import { useTranslation } from "react-i18next";

import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { useToken } from "@hooks/Revoke";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { deleteResetPasswordToken } from "@services/ResetPassword";
import LoadingPage from "@views/LoadingPage/LoadingPage";

const RevokeResetPasswordTokenView = function () {
    const { t: translate } = useTranslation();
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const token = useToken();
    const navigate = useRouterNavigate();

    const handleRedirect = useCallback(() => {
        setTimeout(() => {
            navigate(IndexRoute, false);
        }, 1500);
    }, [navigate]);

    const handleRevoke = useCallback(async () => {
        if (!token) return;

        const { ok, status } = await deleteResetPasswordToken(token);

        if (ok) {
            createSuccessNotification(translate("Successfully revoked the Token"));
        } else if (status === 429) {
            createErrorNotification(translate("You have made too many requests"));
        } else {
            createErrorNotification(translate("Failed to revoke the Token"));
        }

        handleRedirect();
    }, [createErrorNotification, createSuccessNotification, handleRedirect, token, translate]);

    useEffect(() => {
        if (!token) {
            createErrorNotification(translate("The Token was not provided"));

            handleRedirect();

            return;
        }

        handleRevoke().catch(console.error);
    }, [createErrorNotification, handleRedirect, handleRevoke, token, translate]);

    return <LoadingPage />;
};

export default RevokeResetPasswordTokenView;
