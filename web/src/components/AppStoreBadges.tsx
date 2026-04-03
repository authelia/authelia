import { Box, Link } from "@mui/material";

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

    return (
        <Box className={props.className}>
            <Link href={props.googlePlayLink} target={target} underline="hover">
                <Box component={"img"} src={GooglePlay} alt="google play" width={props.iconSize} />
            </Link>
            <Link href={props.appleStoreLink} target={target} underline="hover">
                <Box component={"img"} src={AppleStore} alt="apple store" width={props.iconSize} />
            </Link>
        </Box>
    );
};

export default AppStoreBadges;
