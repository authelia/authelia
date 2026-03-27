import { useCallback } from "react";

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useRedirector } from "@hooks/Redirector";
import { checkSafeRedirection } from "@services/SafeRedirection";

interface Props {
    open: boolean;
    setClosed: () => void;
}

const RedirectAfterEnrollmentDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const redirectionURL = useQueryParam(RedirectionURL);
    const redirect = useRedirector();
    const { createErrorNotification } = useNotifications();

    const targetURL = props.open ? redirectionURL : null;

    const handleContinue = useCallback(async () => {
        if (!targetURL) {
            props.setClosed();

            return;
        }

        try {
            const res = await checkSafeRedirection(targetURL);

            if (res) {
                redirect(targetURL);
            } else {
                createErrorNotification(
                    translate(
                        "Redirection was determined to be unsafe and aborted ensure the redirection URL is correct",
                        {
                            ns: "portal",
                        },
                    ),
                );
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(
                translate("Redirection was determined to be unsafe and aborted ensure the redirection URL is correct", {
                    ns: "portal",
                }),
            );
        }

        props.setClosed();
    }, [targetURL, redirect, createErrorNotification, translate, props]);

    const handleStayHere = useCallback(() => {
        props.setClosed();
    }, [props]);

    if (!targetURL) {
        return null;
    }

    return (
        <Dialog open={props.open} onClose={handleStayHere} maxWidth={"sm"} fullWidth={true}>
            <DialogTitle>{translate("Multi-Factor Authentication Registered")}</DialogTitle>
            <DialogContent>
                <SuccessIcon />
                <DialogContentText>
                    {translate("You have successfully added a multi-factor authentication method")}
                </DialogContentText>
                <DialogContentText>
                    {translate("Would you like to continue to your originally requested resource?")}
                </DialogContentText>
                <Typography variant={"body2"}>{targetURL}</Typography>
            </DialogContent>
            <DialogActions>
                <Button id={"dialog-stay-here"} onClick={handleStayHere}>{translate("Stay Here")}</Button>
                <Button id={"dialog-continue"} onClick={handleContinue}>{translate("Continue")}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default RedirectAfterEnrollmentDialog;
