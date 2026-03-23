import { useCallback } from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

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

    if (!targetURL) {
        return null;
    }

    return (
        <Dialog open={props.open} onClose={props.setClosed}>
            <DialogTitle>{translate("Multi-Factor Authentication Registered")}</DialogTitle>
            <DialogContent>
                <DialogContentText>{targetURL}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={props.setClosed}>{translate("Close")}</Button>
                <Button onClick={handleContinue}>{translate("Continue")}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default RedirectAfterEnrollmentDialog;
