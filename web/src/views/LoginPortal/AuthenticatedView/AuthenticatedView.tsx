import React from "react";

import { Button, Grid, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { LogoutRoute as SignOutRoute } from "@constants/Routes";
import MinimalLayout from "@layouts/MinimalLayout";
import { UserInfo } from "@models/UserInfo";
import Authenticated from "@views/LoginPortal/Authenticated";

export interface Props {
    userInfo: UserInfo;
}

const AuthenticatedView = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useNavigate();

    const styles = useStyles();

    const handleLogoutClick = () => {
        navigate(SignOutRoute);
    };

    return (
        <MinimalLayout
            id="authenticated-stage"
            title={`${translate("Hi")} ${props.userInfo.display_name}`}
            userInfo={props.userInfo}
        >
            <Grid container>
                <Grid item xs={12}>
                    <Button color="secondary" onClick={handleLogoutClick} id="logout-button">
                        {translate("Logout")}
                    </Button>
                </Grid>
                <Grid item xs={12} className={styles.mainContainer}>
                    <Authenticated />
                </Grid>
            </Grid>
        </MinimalLayout>
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    mainContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));

export default AuthenticatedView;
