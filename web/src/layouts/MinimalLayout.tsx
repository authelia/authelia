import React, { ReactNode, useEffect } from "react";

import { Box, Container, Theme } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import UserSvg from "@assets/images/user.svg?react";
import AppBarLoginPortal from "@components/AppBarLoginPortal";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { EncodedName } from "@constants/constants";
import { UserInfo } from "@models/UserInfo";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: null | string;
    userInfo?: UserInfo;
}

const MinimalLayout = function (props: Props) {
    const { t: translate } = useTranslation();

    const { classes } = useStyles();

    const logo = getLogoOverride() ? (
        <Box component={"img"} src="./static/media/logo.png" alt="Logo" className={classes.icon} />
    ) : (
        <UserSvg className={classes.icon} />
    );

    useEffect(() => {
        document.title = translate("Login - {{authelia}}", { authelia: atob(String.fromCharCode(...EncodedName)) });
    }, [translate]);

    return (
        <Box>
            <AppBarLoginPortal userInfo={props.userInfo} />
            <Grid
                id={props.id}
                className={classes.root}
                container
                spacing={0}
                alignItems={"center"}
                justifyContent={"center"}
            >
                <Container maxWidth={"xs"} className={classes.rootContainer}>
                    <Grid container>
                        <Grid size={{ xs: 12 }}>{logo}</Grid>
                        {props.title ? (
                            <Grid size={{ xs: 12 }}>
                                <TypographyWithTooltip variant={"h5"} value={props.title} />
                            </Grid>
                        ) : null}
                        <Grid size={{ xs: 12 }} className={classes.body}>
                            {props.children}
                        </Grid>
                    </Grid>
                </Container>
                <PrivacyPolicyDrawer />
            </Grid>
        </Box>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    body: {
        marginTop: theme.spacing(),
        paddingBottom: theme.spacing(),
        paddingTop: theme.spacing(),
    },
    icon: {
        fill: theme.custom.icon,
        margin: theme.spacing(),
        width: "64px",
    },
    root: {
        minHeight: "90vh",
        textAlign: "center",
    },
    rootContainer: {
        paddingLeft: 32,
        paddingRight: 32,
    },
}));

export default MinimalLayout;
