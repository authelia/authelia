import React, { useState, useEffect } from "react";

import { makeStyles } from "@material-ui/core";
import classnames from "classnames";
import { useTranslation } from "react-i18next";
import zxcvbn from "zxcvbn";

export interface Props {
    value: string;
    /**
     * mode password meter mode
     *   classic: classic mode (checks lowercase, uppercase, specials and numbers)
     *   zxcvbn: uses zxcvbn package to get the password strength
     **/
    mode: string;
    minLength: number;
    maxLength: number;
    requireLowerCase: boolean;
    requireUpperCase: boolean;
    requireNumber: boolean;
    requireSpecial: boolean;
}

const PasswordMeter = function (props: Props) {
    const [progressColor] = useState(["#D32F2F", "#FF5722", "#FFEB3B", "#AFB42B", "#62D32F"]);
    const [passwordScore, setPasswordScore] = useState(0);
    const [maxScores, setMaxScores] = useState(0);
    const [feedback, setFeedback] = useState("");
    const { t: translate } = useTranslation("Portal");
    const style = makeStyles((theme) => ({
        progressBar: {
            height: "5px",
            marginTop: "2px",
            backgroundColor: "red",
            width: "50%",
            transition: "width .5s linear",
        },
    }))();

    useEffect(() => {
        const password = props.value;
        if (props.mode === "classic") {
            //use mode mode
            setMaxScores(4);
            if (password.length < props.minLength) {
                setPasswordScore(0);
                setFeedback(translate("Must be at least {{len}} characters in length", { len: props.minLength }));
                return;
            }
            if (password.length > props.maxLength) {
                setPasswordScore(0);
                setFeedback(translate("Must not be more than {{len}} characters in length", { len: props.maxLength }));
                return;
            }
            setFeedback("");
            let score = 1;
            let required = 0;
            let hits = 0;
            let warning = "";
            if (props.requireLowerCase) {
                required++;
                const hasLowercase = /[a-z]/.test(password);
                if (hasLowercase) {
                    hits++;
                } else {
                    warning += "* " + translate("Must have at least one lowercase letter") + "\n";
                }
            }

            if (props.requireUpperCase) {
                required++;
                const hasUppercase = /[A-Z]/.test(password);
                if (hasUppercase) {
                    hits++;
                } else {
                    warning += "* " + translate("Must have at least one UPPERCASE letter") + "\n";
                }
            }

            if (props.requireNumber) {
                required++;
                const hasNumber = /[0-9]/.test(password);
                if (hasNumber) {
                    hits++;
                } else {
                    warning += "* " + translate("Must have at least one number") + "\n";
                }
            }

            if (props.requireSpecial) {
                required++;
                const hasSpecial = /[^0-9\w]/i.test(password);
                if (hasSpecial) {
                    hits++;
                } else {
                    warning += "* " + translate("Must have at least one special character") + "\n";
                }
            }
            score += hits > 0 ? 1 : 0;
            score += required === hits ? 1 : 0;
            setFeedback(translate("The password does not meet the password policy") + ":\n" + warning);
            setPasswordScore(score);
        } else if (props.mode === "zxcvbn") {
            //use zxcvbn mode
            setMaxScores(5);
            const { score, feedback } = zxcvbn(password);
            setFeedback(feedback.warning);
            setPasswordScore(score);
        }
    }, [props, translate]);

    if (props.mode === "" || props.mode === "none") return <span></span>;

    return (
        <div
            style={{
                width: "100%",
            }}
        >
            <div
                title={feedback}
                className={classnames(style.progressBar)}
                style={{
                    width: `${(passwordScore + 1) * (100 / maxScores)}%`,
                    backgroundColor: progressColor[passwordScore],
                }}
            ></div>
        </div>
    );
};

PasswordMeter.defaultProps = {
    minLength: 0,
};

export default PasswordMeter;
