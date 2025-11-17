import React, { useMemo } from "react";

import { Alert, AlertTitle, Box, Theme } from "@mui/material";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";
import zxcvbn from "zxcvbn";

import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy";

export interface Props {
    value: string;
    policy?: PasswordPolicyConfiguration;
}

const getStandardPasswordInfo = (
    password: string,
    policy: PasswordPolicyConfiguration,
    translate: (key: string, options?: any) => string,
) => {
    let feedback = "";
    let feedbackTitle = "";
    let passwordScore = 0;
    const maxScore = 3;

    if (password.length < policy.min_length) {
        feedback = translate("Must be at least {{len}} characters in length", {
            len: policy.min_length,
        });
        return { feedback, feedbackTitle, maxScore, passwordScore };
    }

    if (policy.max_length !== 0 && password.length > policy.max_length) {
        feedback = translate("Must not be more than {{len}} characters in length", {
            len: policy.max_length,
        });
        return { feedback, feedbackTitle, maxScore, passwordScore };
    }

    let required = 0;
    let hits = 0;
    let warning = "";

    const checks = [
        {
            message: "Must have at least one lowercase letter",
            regex: /[a-z]/,
            require: policy.require_lowercase,
        },
        {
            message: "Must have at least one UPPERCASE letter",
            regex: /[A-Z]/,
            require: policy.require_uppercase,
        },
        {
            message: "Must have at least one number",
            regex: /\d/,
            require: policy.require_number,
        },
        {
            message: "Must have at least one special character",
            regex: /[^a-z0-9]/i,
            require: policy.require_special,
        },
    ];

    for (const { message, regex, require } of checks) {
        if (require) {
            required++;
            if (regex.test(password)) {
                hits++;
            } else {
                warning += "* " + translate(message) + "\n";
            }
        }
    }

    let score = 1;
    score += hits > 0 ? 1 : 0;
    score += required === hits ? 1 : 0;

    if (warning !== "") {
        feedbackTitle = translate("The password does not meet the password policy");
    }

    feedback = warning;
    passwordScore = score;

    return { feedback, feedbackTitle, maxScore, passwordScore };
};

const getZXCVBNPasswordInfo = (password: string) => {
    const { feedback: zxcvbnFeedback, score } = zxcvbn(password);
    return {
        feedback: "",
        feedbackTitle: zxcvbnFeedback.warning,
        maxScore: 4,
        passwordScore: score,
    };
};

const PasswordMeter = function (props: Props) {
    const { t: translate } = useTranslation();

    const { feedback, feedbackTitle, isZXCVBN, maxScore, passwordScore } = useMemo(() => {
        const password = props.value;
        const policy = props.policy;

        if (!policy) {
            return {
                feedback: "",
                feedbackTitle: "",
                isZXCVBN: false,
                maxScore: 3,
                passwordScore: 0,
            };
        }

        if (policy.mode === PasswordPolicyMode.Standard) {
            return {
                ...getStandardPasswordInfo(password, policy, translate),
                isZXCVBN: false,
            };
        } else if (policy.mode === PasswordPolicyMode.ZXCVBN) {
            return {
                ...getZXCVBNPasswordInfo(password),
                isZXCVBN: true,
            };
        }

        return {
            feedback: "",
            feedbackTitle: "",
            isZXCVBN: false,
            maxScore: 3,
            passwordScore: 0,
        };
    }, [props, translate]);

    const progressColor = isZXCVBN
        ? ["#D32F2F", "#FF5722", "#FFEB3B", "#AFB42B", "#62D32F"]
        : ["#D32F2F", "#FF5722", "#FFEB3B", "#62D32F"];

    const { classes } = useStyles({ maxScore, passwordScore, progressColor });

    return (
        <Box className={classes.progressContainer}>
            <Box className={classes.progressBar} />
            {(feedbackTitle !== "" || feedback !== "") && (
                <Alert severity="warning">
                    {feedbackTitle !== "" && (
                        <AlertTitle className={classes.feedbackTitle}>
                            <p>{feedbackTitle}</p>
                        </AlertTitle>
                    )}
                    <Box className={classes.feedback}>{feedback}</Box>
                </Alert>
            )}
        </Box>
    );
};

PasswordMeter.defaultProps = {
    policy: { minLength: 0 },
};

const useStyles = makeStyles<{ progressColor: string[]; passwordScore: number; maxScore: number }>()(
    (theme: Theme, { maxScore, passwordScore, progressColor }) => ({
        feedback: {
            fontSize: "0.7rem",
            textAlign: "left",
            whiteSpace: "break-spaces",
        },
        feedbackTitle: {
            fontSize: "0.85rem",
            textAlign: "left",
            whiteSpace: "break-spaces",
        },
        progressBar: {
            backgroundColor: progressColor[passwordScore],
            height: "5px",
            marginTop: "2px",
            transition: "width .5s linear",
            width: `${passwordScore * (100 / maxScore)}%`,
        },
        progressContainer: {
            width: "100%",
        },
    }),
);

export default PasswordMeter;
