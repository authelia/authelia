import React, { useState } from "react";

import { Grid, Button, makeStyles } from "@material-ui/core";
import classnames from "classnames";
import { useHistory } from "react-router";

import FixedTextField from "../../../components/FixedTextField";
import { useNotifications } from "../../../hooks/NotificationsContext";
import LoginLayout from "../../../layouts/LoginLayout";
import { AuthenticatedRoute } from "../../../Routes";
import { updatePassword } from "../../../services/UpdatePassword";

const UpdatePasswordView = function () {
    const style = useStyles();
    const [formDisabled, setFormDisabled] = useState(false);
    const [oldPassword, setOldPassword] = useState("");
    const [password1, setPassword1] = useState("");
    const [password2, setPassword2] = useState("");
    const [errorOldPassword, setErrorOldPassword] = useState(false);
    const [errorPassword1, setErrorPassword1] = useState(false);
    const [errorPassword2, setErrorPassword2] = useState(false);
    const { createSuccessNotification, createErrorNotification } = useNotifications();
    const history = useHistory();

    const doUpdatePassword = async () => {
        if (password1 === "" || password2 === "") {
            if (oldPassword === "") {
                setErrorOldPassword(true);
            }
            if (password1 === "") {
                setErrorPassword1(true);
            }
            if (password2 === "") {
                setErrorPassword2(true);
            }
            return;
        }
        if (password1 !== password2) {
            setErrorPassword1(true);
            setErrorPassword2(true);
            createErrorNotification("Passwords do not match.");
            return;
        }

        try {
            await updatePassword(oldPassword, password1);
            createSuccessNotification("Password has been udpated.");
            setTimeout(() => history.push(AuthenticatedRoute), 1500);
            setFormDisabled(true);
        } catch (err) {
            console.error(err);
            if (err.message.includes("0000052D.")) {
                createErrorNotification("Your supplied password does not meet the password policy requirements.");
            } else {
                createErrorNotification("There was an issue resetting the password.");
            }
        }
    };

    const handleUpdateClick = () => doUpdatePassword();

    const handleCancelClick = () => history.push(AuthenticatedRoute);

    return (
        <LoginLayout title="Enter new password" id="update-password">
            <Grid container className={style.root} spacing={2}>
                <Grid item xs={12}>
                    <FixedTextField
                        id="password1-textfield"
                        label="Old password"
                        variant="outlined"
                        type="password"
                        value={oldPassword}
                        disabled={formDisabled}
                        onChange={(e) => setOldPassword(e.target.value)}
                        error={errorOldPassword}
                        className={classnames(style.fullWidth)}
                    />
                </Grid>
                <Grid item xs={12}>
                    <FixedTextField
                        id="password1-textfield"
                        label="New password"
                        variant="outlined"
                        type="password"
                        value={password1}
                        disabled={formDisabled}
                        onChange={(e) => setPassword1(e.target.value)}
                        error={errorPassword1}
                        className={classnames(style.fullWidth)}
                    />
                </Grid>
                <Grid item xs={12}>
                    <FixedTextField
                        id="password2-textfield"
                        label="Repeat new password"
                        variant="outlined"
                        type="password"
                        disabled={formDisabled}
                        value={password2}
                        onChange={(e) => setPassword2(e.target.value)}
                        error={errorPassword2}
                        onKeyPress={(ev) => {
                            if (ev.key === "Enter") {
                                doUpdatePassword();
                                ev.preventDefault();
                            }
                        }}
                        className={classnames(style.fullWidth)}
                    />
                </Grid>
                <Grid item xs={6}>
                    <Button
                        id="update-button"
                        variant="contained"
                        color="primary"
                        name="password1"
                        disabled={formDisabled}
                        onClick={handleUpdateClick}
                        className={style.fullWidth}
                    >
                        Update
                    </Button>
                </Grid>
                <Grid item xs={6}>
                    <Button
                        id="cancel-button"
                        variant="contained"
                        color="primary"
                        name="password2"
                        onClick={handleCancelClick}
                        className={style.fullWidth}
                    >
                        Cancel
                    </Button>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default UpdatePasswordView;

const useStyles = makeStyles((theme) => ({
    root: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
    fullWidth: {
        width: "100%",
    },
}));
