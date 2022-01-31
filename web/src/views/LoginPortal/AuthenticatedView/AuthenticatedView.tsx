import React from "react";

import { Grid, makeStyles, Button } from "@material-ui/core";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { LogoutRoute as SignOutRoute } from "@constants/Routes";
import LoginLayout from "@layouts/LoginLayout";
import Authenticated from "@views/LoginPortal/Authenticated";

export interface Props {
    name: string;
}

const AuthenticatedView = function (props: Props) {
    const style = useStyles();
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
                <Grid item xs={12} className={style.mainContainer}>
                    <Authenticated />
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default AuthenticatedView;

const useStyles = makeStyles((theme) => ({
    mainContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
