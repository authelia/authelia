import { useCallback, useEffect, useRef, useState } from "react";

import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import ComponentWithTooltip from "@components/ComponentWithTooltip";
import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import MinimalLayout from "@layouts/MinimalLayout";
import { initiateResetPasswordProcess } from "@services/ResetPassword";
import { cn } from "@utils/Styles";

const ResetPasswordStep1 = function () {
    const [username, setUsername] = useState("");
    const [error, setError] = useState(false);
    const [loading, setLoading] = useState(false);

    const [rateLimited, setRateLimited] = useState(false);
    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    const { createErrorNotification, createInfoNotification } = useNotifications();
    const navigate = useNavigate();
    const { t: translate } = useTranslation();

    useEffect(() => {
        if (timeoutRateLimit.current === null) return;

        return clearTimeout(timeoutRateLimit.current);
    }, []);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setRateLimited(true);

            createErrorNotification(translate("You have made too many requests"));

            timeoutRateLimit.current = setTimeout(() => {
                setRateLimited(false);
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [createErrorNotification, translate],
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
            const response = await initiateResetPasswordProcess(username);
            if (response?.limited === false) {
                createInfoNotification(translate("An email has been sent to your address to complete the process"));
                navigate(IndexRoute);
            } else if (response?.limited) {
                handleRateLimited(response.retryAfter);
            } else {
                createErrorNotification(translate("There was an issue initiating the password reset process"));
            }
        } catch {
            createErrorNotification(translate("There was an issue initiating the password reset process"));
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
                            className={cn(error && "border-destructive")}
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
                                    {loading ? <Spinner className="mr-2 h-5 w-5" /> : null}
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
                </div>
            </div>
        </MinimalLayout>
    );
};

export default ResetPasswordStep1;
