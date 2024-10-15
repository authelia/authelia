import React, { ReactNode, useEffect } from "react";

import { AppBar, Box, Container, Theme, Toolbar, Typography } from "@mui/material";
import Grid from "@mui/material/Grid2";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import UserSvg from "@assets/images/user.svg?react";
import AccountSettingsMenu from "@components/AccountSettingsMenu";
import Brand from "@components/Brand";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { UserInfo } from "@models/UserInfo";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string | null;
    titleTooltip?: string | null;
    subtitle?: string | null;
    subtitleTooltip?: string | null;
    userInfo?: UserInfo;
}

const LoginLayout = function (props: Props) {
    const { t: translate } = useTranslation();

    const styles = useStyles();

    const logo = getLogoOverride() ? (
        <img src="./static/media/logo.png" alt="Logo" className={styles.icon} />
    ) : (
        <UserSvg className={styles.icon} />
    );

    useEffect(() => {
        document.title = `${translate("Login")} - Authelia`;
    }, [translate]);

    return (
        <Box>
            <AppBar position="static" color="transparent" elevation={0}>
                <Toolbar variant="regular">
                    <Typography style={{ flexGrow: 1 }} />
                    {props.userInfo ? <AccountSettingsMenu userInfo={props.userInfo} /> : null}
                </Toolbar>
            </AppBar>
            <Grid
                id={props.id}
                className={styles.root}
                container
                spacing={0}
                alignItems="center"
                justifyContent="center"
            >
                <Container maxWidth="xs" className={styles.rootContainer}>
                    <Grid container>
                        <Grid size={{ xs: 12 }}>{logo}</Grid>
                        {props.title ? (
                            <Grid size={{ xs: 12 }}>
                                <TypographyWithTooltip
                                    variant={"h5"}
                                    value={props.title}
                                    tooltip={props.titleTooltip !== null ? props.titleTooltip : undefined}
                                />
                            </Grid>
                        ) : null}
                        {props.subtitle ? (
                            <Grid size={{ xs: 12 }}>
                                <TypographyWithTooltip
                                    variant={"h6"}
                                    value={props.subtitle}
                                    tooltip={props.subtitleTooltip !== null ? props.subtitleTooltip : undefined}
                                />
                            </Grid>
                        ) : null}
                        <Grid size={{ xs: 12 }} className={styles.body}>
                            {props.children}
                        </Grid>
                        <Brand />
                    </Grid>
                </Container>
                <PrivacyPolicyDrawer />
            </Grid>
        </Box>
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    root: {
        minHeight: "90vh",
        textAlign: "center",
    },
    rootContainer: {
        paddingLeft: 32,
        paddingRight: 32,
    },
    title: {},
    subtitle: {},
    icon: {
        margin: theme.spacing(),
        width: "64px",
        fill: theme.custom.icon,
    },
    body: {
        marginTop: theme.spacing(),
        paddingTop: theme.spacing(),
        paddingBottom: theme.spacing(),
    },
}));

export default LoginLayout;
