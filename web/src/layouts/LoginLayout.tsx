import React, { ReactNode, useEffect } from "react";

import SettingsIcon from "@mui/icons-material/Settings";
import { AppBar, Box, Container, Grid, IconButton, Theme, Toolbar, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { ReactComponent as UserSvg } from "@assets/images/user.svg";
import Brand from "@components/Brand";
import TypographyWithTooltip from "@components/TypographyWithTootip";
import { SettingsRoute } from "@root/constants/Routes";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string;
    titleTooltip?: string;
    subtitle?: string;
    subtitleTooltip?: string;
    showBrand?: boolean;
    showSettings?: boolean;
}

const LoginLayout = function (props: Props) {
    const navigate = useNavigate();
    const styles = useStyles();
    const logo = getLogoOverride() ? (
        <img src="./static/media/logo.png" alt="Logo" className={styles.icon} />
    ) : (
        <UserSvg className={styles.icon} />
    );
    const { t: translate } = useTranslation();
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
                container
                id={props.id}
                className={styles.root}
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
                                    tooltip={props.titleTooltip}
                                />
                            </Grid>
                        ) : null}
                        {props.subtitle ? (
                            <Grid item xs={12}>
                                <TypographyWithTooltip
                                    variant={"h6"}
                                    value={props.subtitle}
                                    tooltip={props.subtitleTooltip}
                                />
                            </Grid>
                        ) : null}
                        <Grid item xs={12} className={styles.body}>
                            {props.children}
                        </Grid>
                        {props.showBrand ? <Brand /> : null}
                    </Grid>
                </Container>
            </Grid>
        </Box>
    );
};

export default LoginLayout;

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
