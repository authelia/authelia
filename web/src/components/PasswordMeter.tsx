import { useMemo } from "react";

import { useTranslation } from "react-i18next";
import zxcvbn from "zxcvbn";

import { Alert, AlertTitle } from "@components/UI/Alert";
import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy";

export interface Props {
    value: string;
    policy?: PasswordPolicyConfiguration;
}

const getStandardPasswordInfo = (
    password: string,
    policy: PasswordPolicyConfiguration,
    translate: (_key: string, _options?: any) => string,
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

    return (
        <div className="w-full">
            <div
                className="mt-0.5 transition-[width] duration-500 linear"
                style={{
                    backgroundColor: progressColor[passwordScore],
                    height: "5px",
                    width: `${passwordScore * (100 / maxScore)}%`,
                }}
            />
            {(feedbackTitle !== "" || feedback !== "") && (
                <Alert variant="default">
                    {feedbackTitle !== "" && (
                        <AlertTitle className="text-[0.85rem] text-left whitespace-break-spaces">
                            <p>{feedbackTitle}</p>
                        </AlertTitle>
                    )}
                    <div className="text-[0.7rem] text-left whitespace-break-spaces">{feedback}</div>
                </Alert>
            )}
        </div>
    );
};

PasswordMeter.defaultProps = {
    policy: { minLength: 0 },
};

export default PasswordMeter;
