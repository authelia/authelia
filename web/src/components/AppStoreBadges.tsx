import React from "react";
import GooglePlay from "../assets/images/googleplay-badge.svg";
import AppleStore from "../assets/images/applestore-badge.svg";
import { Link } from "@material-ui/core";

export interface Props {
    iconSize: number;
    googlePlayLink: string;
    appleStoreLink: string;

    targetBlank?: boolean;
    className?: string;
}

export default function (props: Props) {
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
        </div >
    )
}