import React, { useState, useEffect, useCallback } from "react";
import LoginLayout from "../../layouts/LoginLayout";
import FingerTouchIcon from "../../components/FingerTouchIcon";
import { makeStyles, Typography, Button } from "@material-ui/core";
import { useHistory, useLocation } from "react-router";
import { FirstFactorPath } from "../../services/Api";
import { extractIdentityToken } from "../../utils/IdentityToken";
import { completeU2FRegistrationProcessStep1, completeU2FRegistrationProcessStep2 } from "../../services/RegisterDevice";
import { useNotifications } from "../../hooks/NotificationsContext";
import u2fApi from "u2f-api";

export default function () {
    const style = useStyles();
    const history = useHistory();
    const location = useLocation();
    const { createErrorNotification } = useNotifications();
    const [, setRegistrationInProgress] = useState(false);

    const processToken = extractIdentityToken(location.search);


    const handleBackClick = () => {
        history.push(FirstFactorPath);
    }

    const registerStep1 = useCallback(async () => {
        if (!processToken) {
            return;
        }
        try {
            setRegistrationInProgress(true);
            const res = await completeU2FRegistrationProcessStep1(processToken);
            const registerRequests: u2fApi.RegisterRequest[] = [];
            for (var i in res.registerRequests) {
                const r = res.registerRequests[i];
                registerRequests.push({
                    appId: res.appId,
                    challenge: r.challenge,
                    version: r.version,
                })
            }
            const registerResponse = await u2fApi.register(registerRequests, [], 60);
            await completeU2FRegistrationProcessStep2(registerResponse);
            setRegistrationInProgress(false);
            history.push(FirstFactorPath);
        } catch (err) {
            console.error(err);
            createErrorNotification("Failed to register your security key. " +
                "The identity verification process might have timed out.");
        }
    }, [processToken, createErrorNotification, history]);

    useEffect(() => {
        registerStep1();
    }, [registerStep1]);

    return (
        <LoginLayout title="Touch Security Key">
            <div className={style.icon}>
                <FingerTouchIcon size={64} animated />
            </div>
            <Typography className={style.instruction}>Touch the token on your security key</Typography>
            <Button color="primary" onClick={handleBackClick}>Retry</Button>
            <Button color="primary" onClick={handleBackClick}>Cancel</Button>
        </LoginLayout>
    )
}

const useStyles = makeStyles(theme => ({
    icon: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
    },
    instruction: {
        paddingBottom: theme.spacing(4),
    }
}))