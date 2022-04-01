import React, { CSSProperties } from "react";

import { Grid, Button, useTheme, Theme } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { LogoutRoute as SignOutRoute } from "@constants/Routes";
import LoginLayout from "@layouts/LoginLayout";
import Authenticated from "@views/LoginPortal/Authenticated";

export interface Props {
    name: string;
}

const AuthenticatedView = function (props: Props) {
    const theme = useTheme();
    const style = useStyles(theme);

    const navigate = useNavigate();
    const { t: translate } = useTranslation("Portal");

    const handleLogoutClick = () => {
        navigate(SignOutRoute);
    };

    return (
        <LoginLayout id="authenticated-stage" title={`${translate("Hi")} ${props.name}`} showBrand>
            <Grid container>
                <Grid item xs={12}>
                    <Button color="secondary" onClick={handleLogoutClick} id="logout-button">
                        {translate("Logout")}
                    </Button>
                </Grid>
                <Grid item xs={12} sx={style.mainContainer}>
                    <Authenticated />
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default AuthenticatedView;

const useStyles = (theme: Theme): { [key: string]: CSSProperties } => ({
    mainContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
});
