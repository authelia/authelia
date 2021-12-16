import React, { useState, useEffect } from "react";

import { makeStyles } from "@material-ui/core";
import classnames from "classnames";
import zxcvbn from "zxcvbn";

export interface Props {
    value: string;
    /**
     * legacy mode requires at least one uppercase, lowercase, number, and special letter to be entered
     **/
    legacy?: boolean;
    minLength?: number;
}

const PasswordMeter = function (props: Props) {
    const [progressColor] = useState(["#D32F2F", "#FF5722", "#FFEB3B", "#AFB42B", "#62D32F"]);
    const [passwordScore, setPasswordScore] = useState(0);

    const [maxScores] = useState(5);
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

        if (password.length < props.minLength) {
            setPasswordScore(0);
            return;
        }
        let score = 0;

        if (props.legacy) {
            //use legacy mode
            const hasLowercase = /[a-z]/.test(password);
            score += hasLowercase ? 1 : 0;

            const hasUppercase = /[A-Z]/.test(password);
            score += hasUppercase ? 1 : 0;

            const hasNumber = /[0-9]/.test(password);
            score += hasNumber ? 1 : 0;

            const hasSpecial = /[^0-9\w]/i.test(password);
            score += hasSpecial ? 1 : 0;
        } else {
            //use zxcvbn mode
            const evaluation = zxcvbn(password);
            score = evaluation.score;
        }

        setPasswordScore(score);
    }, [props]);

    return (
        <div
            style={{
                width: "100%",
            }}
        >
            <div
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
