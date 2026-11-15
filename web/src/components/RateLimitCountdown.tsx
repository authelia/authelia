// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useState } from "react";

import { useTranslation } from "react-i18next";

export interface Props {
    secondsRemaining: number;
}

const RateLimitCountdown = function (props: Props) {
    const { t: translate } = useTranslation("portal");
    const [previous, setPrevious] = useState(props.secondsRemaining);
    const [announced, setAnnounced] = useState(props.secondsRemaining);

    if (props.secondsRemaining !== previous) {
        setPrevious(props.secondsRemaining);

        if (props.secondsRemaining > previous || props.secondsRemaining <= 0) {
            setAnnounced(props.secondsRemaining);
        }
    }

    return (
        <Fragment>
            <span role="status" className="sr-only">
                {announced > 0 ? translate("Try again in {{count}} seconds", { count: announced }) : null}
            </span>
            {props.secondsRemaining > 0 ? (
                <p id="rate-limit-countdown" className="w-full text-center text-sm text-muted-foreground">
                    {translate("Try again in {{count}} seconds", { count: props.secondsRemaining })}
                </p>
            ) : null}
        </Fragment>
    );
};

export default RateLimitCountdown;
