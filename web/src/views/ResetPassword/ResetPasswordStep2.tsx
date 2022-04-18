import React, { useCallback, useEffect, useState } from "react";

import { Button, Grid, IconButton, InputAdornment, makeStyles } from "@material-ui/core";
import { Visibility, VisibilityOff } from "@material-ui/icons";
import classnames from "classnames";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate } from "react-router-dom";

import FixedTextField from "@components/FixedTextField";
import PasswordMeter from "@components/PasswordMeter";
import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import LoginLayout from "@layouts/LoginLayout";
import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy";
import { getPasswordPolicyConfiguration } from "@services/PasswordPolicyConfiguration";
import { completeResetPasswordProcess, resetPassword } from "@services/ResetPassword";
import { extractIdentityToken } from "@utils/IdentityToken";

const ResetPasswordStep2 = function () {
    const style = useStyles();
    const location = useLocation();
    const [formDisabled, setFormDisabled] = useState(true);
    const [password1, setPassword1] = useState("");
    const [password2, setPassword2] = useState("");
    const [errorPassword1, setErrorPassword1] = useState(false);
    const [errorPassword2, setErrorPassword2] = useState(false);
    const { createSuccessNotification, createErrorNotification } = useNotifications();
    const { t: translate } = useTranslation();
    const navigate = useNavigate();
    const [showPassword, setShowPassword] = useState(false);

    const [pPolicy, setPPolicy] = useState<PasswordPolicyConfiguration>({
        max_length: 0,
        min_length: 8,
        min_score: 0,
        require_lowercase: false,
        require_number: false,
        require_special: false,
        require_uppercase: false,
        mode: PasswordPolicyMode.Disabled,
    });

    // Get the token from the query param to give it back to the API when requesting
    // the secret for OTP.
    const processToken = extractIdentityToken(location.search);

    const completeProcess = useCallback(async () => {
        if (!processToken) {
            setFormDisabled(true);
            createErrorNotification(translate("No verification token provided"));
            return;
        }

        try {
            setFormDisabled(true);
            await completeResetPasswordProcess(processToken);
            const policy = await getPasswordPolicyConfiguration();
            setPPolicy(policy);
            setFormDisabled(false);
        } catch (err) {
            console.error(err);
            createErrorNotification(
                translate("There was an issue completing the process. The verification token might have expired"),
            );
            setFormDisabled(true);
        }
    }, [processToken, createErrorNotification, translate]);

    useEffect(() => {
        completeProcess();
    }, [completeProcess]);

    const doResetPassword = async () => {
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

        try {
            await resetPassword(password1);
            createSuccessNotification(translate("Password has been reset"));
            setTimeout(() => navigate(IndexRoute), 1500);
            setFormDisabled(true);
        } catch (err) {
            console.error(err);
            if ((err as Error).message.includes("0000052D.")) {
                createErrorNotification(
                    translate("Your supplied password does not meet the password policy requirements"),
                );
            } else if ((err as Error).message.includes("policy")) {
                createErrorNotification(
                    translate("Your supplied password does not meet the password policy requirements"),
                );
            } else {
                createErrorNotification(translate("There was an issue resetting the password"));
            }
        }
    };

    const handleResetClick = () => doResetPassword();

    const handleCancelClick = () => navigate(IndexRoute);

    return (
        <LoginLayout title={translate("Enter new password")} id="reset-password-step2-stage">
            <Grid container className={style.root} spacing={2}>
                <Grid item xs={12}>
                    <FixedTextField
                        id="password1-textfield"
                        label={translate("New password")}
                        variant="outlined"
                        type={showPassword ? "text" : "password"}
                        value={password1}
                        disabled={formDisabled}
                        onChange={(e) => setPassword1(e.target.value)}
                        error={errorPassword1}
                        className={classnames(style.fullWidth)}
                        autoComplete="new-password"
                        InputProps={{
                            endAdornment: (
                                <InputAdornment position="end">
                                    <IconButton
                                        aria-label="toggle password visibility"
                                        onClick={(e) => setShowPassword(!showPassword)}
                                        edge="end"
                                    >
                                        {showPassword ? <VisibilityOff></VisibilityOff> : <Visibility></Visibility>}
                                    </IconButton>
                                </InputAdornment>
                            ),
                        }}
                    />
                    {pPolicy.mode === PasswordPolicyMode.Disabled ? null : (
                        <PasswordMeter value={password1} policy={pPolicy} />
                    )}
                </Grid>
                <Grid item xs={12}>
                    <FixedTextField
                        id="password2-textfield"
                        label={translate("Repeat new password")}
                        variant="outlined"
                        type={showPassword ? "text" : "password"}
                        disabled={formDisabled}
                        value={password2}
                        onChange={(e) => setPassword2(e.target.value)}
                        error={errorPassword2}
                        onKeyPress={(ev) => {
                            if (ev.key === "Enter") {
                                doResetPassword();
                                ev.preventDefault();
                            }
                        }}
                        className={classnames(style.fullWidth)}
                        autoComplete="new-password"
                    />
                </Grid>
                <Grid item xs={6}>
                    <Button
                        id="reset-button"
                        variant="contained"
                        color="primary"
                        name="password1"
                        disabled={formDisabled}
                        onClick={handleResetClick}
                        className={style.fullWidth}
                    >
                        {translate("Reset")}
                    </Button>
                </Grid>
                <Grid item xs={6}>
                    <Button
                        id="cancel-button"
                        variant="contained"
                        color="primary"
                        name="password2"
                        onClick={handleCancelClick}
                        className={style.fullWidth}
                    >
                        {translate("Cancel")}
                    </Button>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default ResetPasswordStep2;

const useStyles = makeStyles((theme) => ({
    root: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
    fullWidth: {
        width: "100%",
    },
}));
