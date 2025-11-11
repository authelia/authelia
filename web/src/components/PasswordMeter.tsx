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
        return { passwordScore, maxScore, feedback, feedbackTitle };
    }

    if (policy.max_length !== 0 && password.length > policy.max_length) {
        feedback = translate("Must not be more than {{len}} characters in length", {
            len: policy.max_length,
        });
        return { passwordScore, maxScore, feedback, feedbackTitle };
    }

    let required = 0;
    let hits = 0;
    let warning = "";

    const checks = [
        {
            require: policy.require_lowercase,
            regex: /[a-z]/,
            message: "Must have at least one lowercase letter",
        },
        {
            require: policy.require_uppercase,
            regex: /[A-Z]/,
            message: "Must have at least one UPPERCASE letter",
        },
        {
            require: policy.require_number,
            regex: /\d/,
            message: "Must have at least one number",
        },
        {
            require: policy.require_special,
            regex: /[^a-z0-9]/i,
            message: "Must have at least one special character",
        },
    ];

    for (const { require, regex, message } of checks) {
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

    return { passwordScore, maxScore, feedback, feedbackTitle };
};

const getZXCVBNPasswordInfo = (password: string) => {
    const { score, feedback: zxcvbnFeedback } = zxcvbn(password);
    return {
        passwordScore: score,
        maxScore: 4,
        feedback: "",
        feedbackTitle: zxcvbnFeedback.warning,
    };
};

const PasswordMeter = function (props: Props) {
    const { t: translate } = useTranslation();

    const { passwordScore, maxScore, feedback, feedbackTitle, isZXCVBN } = useMemo(() => {
        const password = props.value;
        const policy = props.policy;

        if (!policy) {
            return {
                passwordScore: 0,
                maxScore: 3,
                feedback: "",
                feedbackTitle: "",
                isZXCVBN: false,
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
            passwordScore: 0,
            maxScore: 3,
            feedback: "",
            feedbackTitle: "",
            isZXCVBN: false,
        };
    }, [props, translate]);

    const progressColor = isZXCVBN
        ? ["#D32F2F", "#FF5722", "#FFEB3B", "#AFB42B", "#62D32F"]
        : ["#D32F2F", "#FF5722", "#FFEB3B", "#62D32F"];

    const { classes } = useStyles({ progressColor, passwordScore, maxScore });

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
    (theme: Theme, { progressColor, passwordScore, maxScore }) => ({
        progressBar: {
            height: "5px",
            marginTop: "2px",
            backgroundColor: progressColor[passwordScore],
            width: `${passwordScore * (100 / maxScore)}%`,
            transition: "width .5s linear",
        },
        progressContainer: {
            width: "100%",
        },
        feedbackTitle: {
            whiteSpace: "break-spaces",
            textAlign: "left",
            fontSize: "0.85rem",
        },
        feedback: {
            whiteSpace: "break-spaces",
            textAlign: "left",
            fontSize: "0.7rem",
        },
    }),
);

export default PasswordMeter;
