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

            if (res?.ok) {
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
            <DialogContent
                sx={{
                    alignItems: "center",
                    display: "flex",
                    flexDirection: "column",
                    gap: 2,
                    py: 3,
                    textAlign: "center",
                }}
            >
                <SuccessIcon />
                <DialogContentText sx={{ mt: 2 }}>
                    {translate("You have successfully added a multi-factor authentication method")}
                </DialogContentText>
                <DialogContentText>
                    {translate("Would you like to continue to your originally requested resource?")}
                </DialogContentText>
                <Typography
                    variant={"body2"}
                    sx={(theme) => ({
                        color: theme.palette.primary.main,
                        fontWeight: "bold",
                        wordBreak: "break-all",
                    })}
                >
                    {targetURL}
                </Typography>
            </DialogContent>
            <DialogActions sx={{ gap: 1, justifyContent: "center", pb: 2 }}>
                <Button id={"dialog-stay-here"} color={"secondary"} variant={"outlined"} onClick={handleStayHere}>
                    {translate("Stay Here")}
                </Button>
                <Button id={"dialog-continue"} color={"success"} variant={"contained"} onClick={handleContinue}>
                    {translate("Continue")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default RedirectAfterEnrollmentDialog;
