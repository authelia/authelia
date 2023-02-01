import React, { Fragment, ReactNode, useEffect } from "react";

import { Container, Divider, Grid, Link, Theme } from "@mui/material";
import { grey } from "@mui/material/colors";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import { ReactComponent as UserSvg } from "@assets/images/user.svg";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import PrivacyPolicyLink from "@components/PrivacyPolicyLink";
import TypographyWithTooltip from "@components/TypographyWithTootip";
import { getLogoOverride, getPrivacyPolicyEnabled } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string;
    titleTooltip?: string;
    subtitle?: string;
    subtitleTooltip?: string;
    showBrand?: boolean;
}

const url = "https://www.authelia.com";

const LoginLayout = function (props: Props) {
    const styles = useStyles();
    const { t: translate } = useTranslation();

    const logo = getLogoOverride() ? (
        <img src="./static/media/logo.png" alt="Logo" className={styles.icon} />
    ) : (
        <UserSvg className={styles.icon} />
    );

    const privacyEnabled = getPrivacyPolicyEnabled();

    useEffect(() => {
        document.title = `${translate("Login")} - Authelia`;
    }, [translate]);

    return (
        <Grid id={props.id} className={styles.root} container spacing={0} alignItems="center" justifyContent="center">
            <Container maxWidth="xs" className={styles.rootContainer}>
                <Grid container>
                    <Grid item xs={12}>
                        {logo}
                    </Grid>
                    {props.title ? (
                        <Grid item xs={12}>
                            <TypographyWithTooltip variant={"h5"} value={props.title} tooltip={props.titleTooltip} />
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
                    {props.showBrand ? (
                        <Grid item container xs={12} alignItems="center" justifyContent="center">
                            <Grid item xs={4}>
                                <Link href={url} target="_blank" underline="hover" className={styles.footerLinks}>
                                    {translate("Powered by")} Authelia
                                </Link>
                            </Grid>
                            {privacyEnabled ? (
                                <Fragment>
                                    <Grid item xs={4}>
                                        <PrivacyPolicyLink className={styles.footerLinks} />
                                    </Grid>
                                </Fragment>
                            ) : null}
                        </Grid>
                    ) : null}
                </Grid>
            </Container>
            <PrivacyPolicyDrawer />
        </Grid>
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
    footerLinks: {
        fontSize: "0.7em",
        color: grey[500],
    },
}));
