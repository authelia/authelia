import { useCallback, useEffect, useState } from "react";

import { Visibility, VisibilityOff } from "@mui/icons-material";
import { Button, FormControl, IconButton, InputAdornment, Theme } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { makeStyles } from "tss-react/mui";

import PasswordMeter from "@components/PasswordMeter";
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
    const { classes, cx } = useStyles();

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

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            createErrorNotification(translate("You have made too many requests")); // TODO: Do we want to add the amount of seconds a user should retry in the message?
        },
        [createErrorNotification, translate],
    );

    useEffect(() => {
        const submitReset = async () => {
            if (!processToken) {
                setFormDisabled(true);
                createErrorNotification(translate("No verification token provided"));
                return;
            }

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
            <FormControl id={"form-reset-password"}>
                <Grid container className={classes.root} spacing={2}>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            id="password1-textfield"
                            label={translate("New password")}
                            variant="outlined"
                            type={showPassword ? "text" : "password"}
                            value={password1}
                            disabled={formDisabled}
                            onChange={(e) => setPassword1(e.target.value)}
                            error={errorPassword1}
                            className={cx(classes.fullWidth)}
                            autoComplete="new-password"
                            slotProps={{
                                input: {
                                    endAdornment: (
                                        <InputAdornment position="end">
                                            <IconButton
                                                aria-label="toggle password visibility"
                                                edge="end"
                                                size="large"
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
                                                {showPassword ? <Visibility /> : <VisibilityOff />}
                                            </IconButton>
                                        </InputAdornment>
                                    ),
                                },
                            }}
                        />
                        {pPolicy.mode === PasswordPolicyMode.Disabled ? null : (
                            <PasswordMeter value={password1} policy={pPolicy} />
                        )}
                    </Grid>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            id="password2-textfield"
                            label={translate("Repeat new password")}
                            variant="outlined"
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
                            className={cx(classes.fullWidth)}
                            autoComplete="new-password"
                        />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                        <Button
                            id="reset-button"
                            variant="contained"
                            color="primary"
                            name="password1"
                            disabled={formDisabled}
                            onClick={handleResetClick}
                            className={classes.fullWidth}
                        >
                            {translate("Reset")}
                        </Button>
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                        <Button
                            id="cancel-button"
                            variant="contained"
                            color="primary"
                            name="password2"
                            onClick={handleCancelClick}
                            className={classes.fullWidth}
                        >
                            {translate("Cancel")}
                        </Button>
                    </Grid>
                </Grid>
            </FormControl>
        </MinimalLayout>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    fullWidth: {
        width: "100%",
    },
    root: {
        marginBottom: theme.spacing(2),
        marginTop: theme.spacing(2),
    },
}));

export default ResetPasswordStep2;
