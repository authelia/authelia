import React, { ReactNode } from "react";

import { Grid, Container, Typography, Link, Theme, useTheme, SvgIcon, Box } from "@mui/material";
import { grey } from "@mui/material/colors";
import { CSSProperties } from "@mui/styles";

import { ReactComponent as UserSvg } from "@assets/images/user.svg";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string;
    showBrand?: boolean;
}

const LoginLayout = function (props: Props) {
    const theme = useTheme();
    const style = useStyles(theme);

    const logo = getLogoOverride() ? (
        <Box component="img" src="./static/media/logo.png" alt="Logo" sx={style.icon} />
    ) : (
        <SvgIcon component={UserSvg} sx={style.icon} />
    );
    return (
        <Grid id={props.id} sx={style.root} container spacing={0} alignItems="center" justifyContent="center">
            <Container maxWidth="xs" sx={style.rootContainer}>
                <Grid container>
                    <Grid item xs={12}>
                        {logo}
                    </Grid>
                    {props.title ? (
                        <Grid item xs={12}>
                            <Typography variant="h5" sx={style.title}>
                                {props.title}
                            </Typography>
                        </Grid>
                    ) : null}
                    <Grid item xs={12} sx={style.body}>
                        {props.children}
                    </Grid>
                    {props.showBrand ? (
                        <Grid item xs={12}>
                            <Link
                                href="https://github.com/authelia/authelia"
                                target="_blank"
                                sx={style.poweredBy}
                                underline="hover"
                            >
                                Powered by Authelia
                            </Link>
                        </Grid>
                    ) : null}
                </Grid>
            </Container>
        </Grid>
    );
};

export default LoginLayout;

const useStyles = (theme: Theme): { [key: string]: CSSProperties } => ({
    root: {
        minHeight: "90vh",
        textAlign: "center",
    },
    rootContainer: {
        paddingLeft: 32,
        paddingRight: 32,
    },
    title: {},
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
});
