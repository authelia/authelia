import React from "react";

import { Box, Link, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";

import AppleStore from "@assets/images/applestore-badge.svg";
import GooglePlay from "@assets/images/googleplay-badge.svg";

export interface Props {
    iconSize: number;
    googlePlayLink: string;
    appleStoreLink: string;

    targetBlank?: boolean;
    className?: string;
}

const AppStoreBadges = function (props: Props) {
    const target = props.targetBlank ? "_blank" : undefined;

    const styles = makeStyles((theme: Theme) => ({
        badge: {
            width: props.iconSize,
        },
    }))();

    return (
        <Box className={props.className}>
            <Link href={props.googlePlayLink} target={target} underline="hover">
                <img src={GooglePlay} alt="google play" className={styles.badge} />
            </Link>
            <Link href={props.appleStoreLink} target={target} underline="hover">
                <img src={AppleStore} alt="apple store" className={styles.badge} />
            </Link>
        </Box>
    );
};

export default AppStoreBadges;
