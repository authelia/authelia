import React, { CSSProperties } from "react";

import { Link } from "@mui/material";

import AppleStore from "@assets/images/applestore-badge.svg";
import GooglePlay from "@assets/images/googleplay-badge.svg";

export interface Props {
    iconSize: number;
    googlePlayLink: string;
    appleStoreLink: string;

    targetBlank?: boolean;
    style?: CSSProperties;
}

const AppStoreBadges = function (props: Props) {
    const target = props.targetBlank ? "_blank" : undefined;

    const width = props.iconSize;

    return (
        <div style={props.style}>
            <Link href={props.googlePlayLink} target={target} underline="hover">
                <img src={GooglePlay} alt="google play" style={{ width }} />
            </Link>
            <Link href={props.appleStoreLink} target={target} underline="hover">
                <img src={AppleStore} alt="apple store" style={{ width }} />
            </Link>
        </div>
    );
};

export default AppStoreBadges;
