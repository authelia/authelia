import React from "react";

import { Link } from "@material-ui/core";

import AppleStore from "../assets/images/applestore-badge.svg";
import GooglePlay from "../assets/images/googleplay-badge.svg";

export interface Props {
    iconSize: number;
    googlePlayLink: string;
    appleStoreLink: string;

    targetBlank?: boolean;
    className?: string;
}

const AppStoreBadges = function (props: Props) {
    const target = props.targetBlank ? "_blank" : undefined;

    const width = props.iconSize;

    return (
        <div className={props.className}>
            <Link href={props.googlePlayLink} target={target}>
                <img src={GooglePlay} alt="google play" style={{ width }} />
            </Link>
            <Link href={props.appleStoreLink} target={target}>
                <img src={AppleStore} alt="apple store" style={{ width }} />
            </Link>
        </div>
    );
};

export default AppStoreBadges;
