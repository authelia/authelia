import { useCallback, useEffect, useRef, useState } from "react";

import { Eye, EyeOff } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import PasswordMeter from "@components/PasswordMeter";
import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { IndexRoute } from "@constants/Routes";
import { IdentityToken } from "@constants/SearchParams";
import { useNotifications } from "@hooks/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import MinimalLayout from "@layouts/MinimalLayout";
import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy";
import { getPasswordPolicyConfiguration } from "@services/PasswordPolicyConfiguration";
import { completeResetPasswordProcess, resetPassword } from "@services/ResetPassword";

const ResetPasswordStep2 = function () {
    const { t: translate } = useTranslation();

    const [formDisabled, setFormDisabled] = useState(true);
    const [password1, setPassword1] = useState("");
    const [password2, setPassword2] = useState("");
    const [errorPassword1, setErrorPassword1] = useState(false);
    const [errorPassword2, setErrorPassword2] = useState(false);
    const { createErrorNotification, createSuccessNotification } = useNotifications();
    const navigate = useNavigate();
    const [showPassword, setShowPassword] = useState(false);

    const [pPolicy, setPPolicy] = useState<PasswordPolicyConfiguration>({
        max_length: 0,
        min_length: 8,
        min_score: 0,
        mode: PasswordPolicyMode.Disabled,
        require_lowercase: false,
        require_number: false,
        require_special: false,
        require_uppercase: false,
    });

    // Get the token from the query param to give it back to the API when requesting
    // the secret for OTP.
    const processToken = useQueryParam(IdentityToken);
    const tokenSubmittedRef = useRef(false);

    const handleRateLimited = useCallback(
        (_retryAfter: number) => {
            createErrorNotification(translate("You have made too many requests")); // TODO: Do we want to add the amount of seconds a user should retry in the message?
        },
        [createErrorNotification, translate],
    );

    useEffect(() => {
        if (tokenSubmittedRef.current) return;

        const submitReset = async () => {
            if (!processToken) {
                setFormDisabled(true);
                createErrorNotification(translate("No verification token provided"));
                return;
            }

            tokenSubmittedRef.current = true;

            try {
                const response = await completeResetPasswordProcess(processToken);

                if (response?.limited) {
                    handleRateLimited(response.retryAfter);
                    return;
                }

                const policy = await getPasswordPolicyConfiguration();
                setPPolicy(policy);
                setFormDisabled(false);
            } catch (err) {
                console.error(err);
                createErrorNotification(
                    translate("There was an issue completing the process the verification token might have expired"),
                );
                setFormDisabled(true);
            }
        };

        submitReset().catch(console.error);
    }, [processToken, createErrorNotification, translate, handleRateLimited]);

    const doResetPassword = async () => {
        setPassword1("");
        setPassword2("");

        if (password1 === "" || password2 === "") {
            if (password1 === "") {
                setErrorPassword1(true);
            }
            if (password2 === "") {
                setErrorPassword2(true);
            }
            return;
        }

        if (password1 !== password2) {
            setErrorPassword1(true);
            setErrorPassword2(true);
            createErrorNotification(translate("Passwords do not match"));
            return;
        }

        setFormDisabled(true);

        try {
            await resetPassword(password1);

            createSuccessNotification(translate("Password has been reset"));
            setTimeout(() => navigate(IndexRoute), 1500);
        } catch (err) {
            console.error(err);
            if ((err as Error).message.includes("0000052D.") || (err as Error).message.includes("policy")) {
                createErrorNotification(
                    translate("Your supplied password does not meet the password policy requirements"),
                );
            } else {
                createErrorNotification(translate("There was an issue resetting the password"));
            }
        }
    };

    const handleResetClick = () => {
        doResetPassword().catch(console.error);
    };

    const handleCancelClick = () => navigate(IndexRoute);

    return (
        <MinimalLayout title={translate("Enter new password")} id="reset-password-step2-stage">
            <div id={"form-reset-password"}>
                <div className="my-4 grid grid-cols-1 gap-4">
                    <div className="w-full">
                        <Label htmlFor="password1-textfield">{translate("New password")}</Label>
                        <div className="relative">
                            <Input
                                id="password1-textfield"
                                type={showPassword ? "text" : "password"}
                                value={password1}
                                disabled={formDisabled}
                                onChange={(e) => setPassword1(e.target.value)}
                                error={errorPassword1}
                                className="pr-10"
                                autoComplete="new-password"
                            />
                            <button
                                type="button"
                                className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
                                aria-label="toggle password visibility"
                                onMouseDown={() => setShowPassword(true)}
                                onMouseUp={() => setShowPassword(false)}
                                onMouseLeave={() => setShowPassword(false)}
                                onTouchStart={() => setShowPassword(true)}
                                onTouchEnd={() => setShowPassword(false)}
                                onTouchCancel={() => setShowPassword(false)}
                                onKeyDown={(e) => {
                                    if (e.key === " ") {
                                        setShowPassword(true);
                                        e.preventDefault();
                                    }
                                }}
                                onKeyUp={(e) => {
                                    if (e.key === " ") {
                                        setShowPassword(false);
                                        e.preventDefault();
                                    }
                                }}
                            >
                                {showPassword ? <Eye className="h-5 w-5" /> : <EyeOff className="h-5 w-5" />}
                            </button>
                        </div>
                        {pPolicy.mode === PasswordPolicyMode.Disabled ? null : (
                            <PasswordMeter value={password1} policy={pPolicy} />
                        )}
                    </div>
                    <div className="w-full">
                        <Label htmlFor="password2-textfield">{translate("Repeat new password")}</Label>
                        <Input
                            id="password2-textfield"
                            type={showPassword ? "text" : "password"}
                            disabled={formDisabled}
                            value={password2}
                            onChange={(e) => setPassword2(e.target.value)}
                            error={errorPassword2}
                            onKeyDown={(ev) => {
                                if (ev.key === "Enter") {
                                    doResetPassword().catch(console.error);
                                    ev.preventDefault();
                                }
                            }}
                            autoComplete="new-password"
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div className="w-full">
                            <Button
                                id="reset-button"
                                variant="default"
                                name="password1"
                                disabled={formDisabled}
                                onClick={handleResetClick}
                                className="w-full"
                            >
                                {translate("Reset")}
                            </Button>
                        </div>
                        <div className="w-full">
                            <Button
                                id="cancel-button"
                                variant="default"
                                name="password2"
                                onClick={handleCancelClick}
                                className="w-full"
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

export default ResetPasswordStep2;
