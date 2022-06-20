import React, { ReactNode, useEffect } from "react";

import { Grid, Container, Link, Theme } from "@mui/material";
import { grey } from "@mui/material/colors";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import { ReactComponent as UserSvg } from "@assets/images/user.svg";
import TypographyWithTooltip from "@components/TypographyWithTootip";
import { getLogoOverride } from "@utils/Configuration";

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
    const logo = getLogoOverride() ? (
        <img src="./static/media/logo.png" alt="Logo" className={styles.icon} />
    ) : (
        <UserSvg className={styles.icon} />
    );
    const { t: translate } = useTranslation();
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
                        <Grid item xs={12}>
                            <Link href={url} target="_blank" underline="hover" className={styles.poweredBy}>
                                {translate("Powered by")} Authelia
                            </Link>
                        </Grid>
                    ) : null}
                </Grid>
            </Container>
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
    poweredBy: {
        fontSize: "0.7em",
        color: grey[500],
    },
}));
