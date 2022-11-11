import React from "react";

import { Button, Grid, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { SettingsRoute, LogoutRoute as SignOutRoute } from "@constants/Routes";
import LoginLayout from "@layouts/LoginLayout";

export interface Props {
    name: string;
}

const AuthenticatedView = function (props: Props) {
    const styles = useStyles();
    const navigate = useNavigate();
    const { t: translate } = useTranslation();

    const handleSettingsClick = () => {
        navigate(SettingsRoute);
    };
    const handleLogoutClick = () => {
        navigate(SignOutRoute);
    };

    return (
        <LoginLayout id="authenticated-stage" title={`${translate("Hi")} ${props.name}`} showBrand>
            <Grid container>
                <Grid item xs={12} className={styles.mainContainer}>
                    <Grid container justifyContent="center" spacing={1}>
                        <Grid item xs={12} className="method-option">
                            <Button
                                className={styles.item}
                                classes={{ root: styles.buttonRoot }}
                                color="primary"
                                variant="contained"
                                onClick={handleSettingsClick}
                                id="settings-button"
                            >
                                {translate("Settings")}
                            </Button>
                        </Grid>
                        <Grid item xs={12} className="method-option">
                            <Button
                                className={styles.item}
                                classes={{ root: styles.buttonRoot }}
                                color="secondary"
                                variant="outlined"
                                onClick={handleLogoutClick}
                                id="logout-button"
                            >
                                {translate("Logout")}
                            </Button>
                        </Grid>
                    </Grid>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default AuthenticatedView;

const useStyles = makeStyles((theme: Theme) => ({
    mainContainer: {
        border: "1px solid #d6d6d6",
        borderRadius: "10px",
        padding: theme.spacing(4),
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
    item: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
        width: "100%",
    },
    icon: {
        display: "inline-block",
        fill: "white",
    },
    buttonRoot: {
        display: "block",
    },
}));
