// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback } from "react";

import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";
import { Button } from "@components/UI/Button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/UI/Dialog";
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
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleStayHere();
            }}
        >
            <DialogContent showCloseButton={false} className="sm:max-w-sm w-full">
                <DialogHeader>
                    <DialogTitle>{translate("Multi-Factor Authentication Registered")}</DialogTitle>
                </DialogHeader>
                <div className="flex flex-col items-center gap-2 py-2 text-center">
                    <SuccessIcon />
                    <DialogDescription className="mt-2">
                        {translate("You have successfully added a multi-factor authentication method")}
                    </DialogDescription>
                    <DialogDescription>
                        {translate("Would you like to continue to your originally requested resource?")}
                    </DialogDescription>
                    <p className="font-bold text-primary break-all text-sm">{targetURL}</p>
                </div>
                <DialogFooter className="sm:justify-center">
                    <Button id={"dialog-stay-here"} variant={"outline"} color={"secondary"} onClick={handleStayHere}>
                        {translate("Stay Here")}
                    </Button>
                    <Button id={"dialog-continue"} variant={"default"} color={"success"} onClick={handleContinue}>
                        {translate("Continue")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default RedirectAfterEnrollmentDialog;
