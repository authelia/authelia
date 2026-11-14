// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from "react";

import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";

import ComponentWithTooltip from "@components/ComponentWithTooltip";
import RateLimitCountdown from "@components/RateLimitCountdown";
import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@contexts/NotificationsContext";
import { useRateLimit } from "@hooks/RateLimit";
import MinimalLayout from "@layouts/MinimalLayout";
import { RateLimitedError } from "@services/RateLimit";
import { initiateResetPasswordProcess } from "@services/ResetPassword";

const ResetPasswordStep1 = function () {
    const [username, setUsername] = useState("");
    const [error, setError] = useState(false);
    const [loading, setLoading] = useState(false);

    const { limited: rateLimited, secondsRemaining, trigger: triggerRateLimit } = useRateLimit();

    const { createErrorNotification, createInfoNotification } = useNotifications();
    const navigate = useNavigate();
    const { t: translate } = useTranslation();

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            triggerRateLimit(retryAfter);

            createErrorNotification(translate("You have made too many requests"));
        },
        [createErrorNotification, translate, triggerRateLimit],
    );

    const doInitiateResetPasswordProcess = async () => {
        setError(false);
        setLoading(true);

        if (username === "") {
            setError(true);
            setLoading(false);
            createErrorNotification(translate("Username is required"));
            return;
        }

        try {
            await initiateResetPasswordProcess(username);
            createInfoNotification(translate("An email has been sent to your address to complete the process"));
            navigate(IndexRoute);
        } catch (err) {
            if (err instanceof RateLimitedError) {
                handleRateLimited(err.retryAfter);
            } else {
                createErrorNotification(translate("There was an issue initiating the password reset process"));
            }
        }
        setLoading(false);
    };

    const handleResetClick = () => {
        doInitiateResetPasswordProcess();
    };

    const handleCancelClick = () => {
        navigate(IndexRoute);
    };

    return (
        <MinimalLayout title={translate("Reset password")} id="reset-password-step1-stage">
            <div id={"form-reset-password-username"}>
                <div className="my-4 grid grid-cols-1 gap-4">
                    <div className="w-full">
                        <Label htmlFor="username-textfield">{translate("Username")}</Label>
                        <Input
                            id="username-textfield"
                            disabled={loading}
                            error={error}
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            onKeyDown={(ev) => {
                                if (ev.key === "Enter") {
                                    ev.preventDefault();
                                    doInitiateResetPasswordProcess();
                                }
                            }}
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div className="w-full">
                            <ComponentWithTooltip
                                render={rateLimited}
                                title={translate("You have made too many requests")}
                            >
                                <Button
                                    id="reset-button"
                                    variant="default"
                                    disabled={loading || rateLimited}
                                    className="w-full"
                                    onClick={handleResetClick}
                                >
                                    {loading ? <Spinner size={20} className="mr-2 h-5 w-5" /> : null}
                                    {translate("Reset")}
                                </Button>
                            </ComponentWithTooltip>
                        </div>
                        <div className="w-full">
                            <Button
                                id="cancel-button"
                                variant="default"
                                disabled={loading}
                                className="w-full"
                                onClick={handleCancelClick}
                            >
                                {translate("Cancel")}
                            </Button>
                        </div>
                    </div>
                    <RateLimitCountdown secondsRemaining={secondsRemaining} />
                </div>
            </div>
        </MinimalLayout>
    );
};

export default ResetPasswordStep1;
