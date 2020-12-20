import React, { useEffect, useCallback, useState } from "react";
import LoginLayout from "../../layouts/LoginLayout";
import classnames from "classnames";
import { makeStyles, Typography, Button, IconButton, Link, CircularProgress, TextField } from "@material-ui/core";
import QRCode from 'qrcode.react';
import AppStoreBadges from "../../components/AppStoreBadges";
import { GoogleAuthenticator } from "../../constants";
import { useHistory, useLocation } from "react-router";
import { completeTOTPRegistrationProcess } from "../../services/RegisterDevice";
import { useNotifications } from "../../hooks/NotificationsContext";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCopy, faTimesCircle } from "@fortawesome/free-solid-svg-icons";
import { red } from "@material-ui/core/colors";
import { extractIdentityToken } from "../../utils/IdentityToken";
import { FirstFactorRoute } from "../../Routes";

const RegisterOneTimePassword = function () {
    const style = useStyles();
    const history = useHistory();
    const location = useLocation();

    // The secret retrieved from the API is all is ok.
    const [secretURL, setSecretURL] = useState("empty");
    const [secretBase32, setSecretBase32] = useState(undefined as string | undefined);
    const { createErrorNotification } = useNotifications();
    const [hasErrored, setHasErrored] = useState(false);
    const [isLoading, setIsLoading] = useState(false);

    // Get the token from the query param to give it back to the API when requesting
    // the secret for OTP.
    const processToken = extractIdentityToken(location.search);

    const handleDoneClick = () => {
        history.push(FirstFactorRoute);
    }

    const completeRegistrationProcess = useCallback(async () => {
        if (!processToken) {
            return;
        }

        setIsLoading(true);
        try {
            const secret = await completeTOTPRegistrationProcess(processToken);
            setSecretURL(secret.otpauth_url);
            setSecretBase32(secret.base32_secret);
        } catch (err) {
            console.error(err);
            createErrorNotification("Failed to generate the code to register your device", 10000);
            setHasErrored(true);
        }
        setIsLoading(false);
    }, [processToken, createErrorNotification]);

    useEffect(() => { completeRegistrationProcess() }, [completeRegistrationProcess]);
    const CopyButton = () => (
      <IconButton
          color="primary"
          onClick={() =>  navigator.clipboard.writeText(`${secretBase32}`)}
      >
          <FontAwesomeIcon icon={faCopy} />
      </IconButton>
    )
    const qrcodeFuzzyStyle = (isLoading || hasErrored) ? style.fuzzy : undefined

    return (
        <LoginLayout title="Scan QRCode">
            <div className={style.root}>
                <div className={style.googleAuthenticator}>
                    <Typography className={style.googleAuthenticatorText}>Need Google Authenticator?</Typography>
                    <AppStoreBadges
                        iconSize={128}
                        targetBlank
                        className={style.googleAuthenticatorBadges}
                        googlePlayLink={GoogleAuthenticator.googlePlay}
                        appleStoreLink={GoogleAuthenticator.appleStore} />
                </div>
                <div className={style.qrcodeContainer}>
                    <Link href={secretURL}>
                        <QRCode
                            value={secretURL}
                            className={classnames(qrcodeFuzzyStyle, style.qrcode)}
                            size={256} />
                        {!hasErrored && isLoading ? <CircularProgress className={style.loader} size={128} /> : null}
                        {hasErrored ? <FontAwesomeIcon className={style.failureIcon} icon={faTimesCircle} /> : null}
                    </Link>
                </div>
                {secretBase32
                    ? <TextField
                        id="base32-secret"
                        label="Secret"
                        className={style.secret}
                        value={secretBase32}
                        InputProps={{
                            autoFocus: true,
                            endAdornment: <CopyButton />,
                            readOnly: true
                        }} /> : null}
                <Button
                    variant="contained"
                    color="primary"
                    className={style.doneButton}
                    onClick={handleDoneClick}
                    disabled={isLoading}>
                    Done
                </Button>
            </div>
        </LoginLayout>
    )
}

export default RegisterOneTimePassword

const useStyles = makeStyles(theme => ({
    root: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    qrcode: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
    fuzzy: {
        filter: "blur(10px)"
    },
    secret: {
        marginTop: theme.spacing(1),
        marginBottom: theme.spacing(1),
    },
    googleAuthenticator: {},
    googleAuthenticatorText: {
        fontSize: theme.typography.fontSize * 0.8,
    },
    googleAuthenticatorBadges: {},
    doneButton: {
        width: "256px",
    },
    qrcodeContainer: {
        position: "relative",
        display: "inline-block",
    },
    loader: {
        position: "absolute",
        top: "calc(128px - 64px)",
        left: "calc(128px - 64px)",
        color: "rgba(255, 255, 255, 0.5)",
    },
    failureIcon: {
        position: "absolute",
        top: "calc(128px - 64px)",
        left: "calc(128px - 64px)",
        color: red[400],
        fontSize: "128px",
    }
}))