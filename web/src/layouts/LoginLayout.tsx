import React, { ReactNode, useEffect } from "react";

import SettingsIcon from "@mui/icons-material/Settings";
import { AppBar, Box, Container, Grid, IconButton, Theme, Toolbar, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import UserSvg from "@assets/images/user.svg?react";
import Brand from "@components/Brand";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { SettingsRoute } from "@constants/Routes";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string | null;
    titleTooltip?: string | null;
    subtitle?: string | null;
    subtitleTooltip?: string | null;
    showBrand?: boolean;
    showSettings?: boolean;
}

const LoginLayout = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useNavigate();
    const styles = useStyles();

    const logo = getLogoOverride() ? (
        <img src="./static/media/logo.png" alt="Logo" className={styles.icon} />
    ) : (
        <UserSvg className={styles.icon} />
    );

    useEffect(() => {
        document.title = `${translate("Login")} - Authelia`;
    }, [translate]);

    const handleSettingsClick = () => {
        navigate({
            pathname: SettingsRoute,
        });
    };

    return (
        <Box>
            <AppBar position="static" color="transparent" elevation={0}>
                <Toolbar variant="dense">
                    <Typography style={{ flexGrow: 1 }} />
                    {props.showSettings ? (
                        <IconButton
                            size="large"
                            edge="start"
                            color="inherit"
                            aria-label="menu"
                            sx={{ mr: 2 }}
                            onClick={handleSettingsClick}
                        >
                            <SettingsIcon />
                        </IconButton>
                    ) : null}
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
                        <Grid item xs={12}>
                            {logo}
                        </Grid>
                        {props.title ? (
                            <Grid item xs={12}>
                                <TypographyWithTooltip
                                    variant={"h5"}
                                    value={props.title}
                                    tooltip={props.titleTooltip !== null ? props.titleTooltip : undefined}
                                />
                            </Grid>
                        ) : null}
                        {props.subtitle ? (
                            <Grid item xs={12}>
                                <TypographyWithTooltip
                                    variant={"h6"}
                                    value={props.subtitle}
                                    tooltip={props.subtitleTooltip !== null ? props.subtitleTooltip : undefined}
                                />
                            </Grid>
                        ) : null}
                        <Grid item xs={12} className={styles.body}>
                            {props.children}
                        </Grid>
                        {props.showBrand ? <Brand /> : null}
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
