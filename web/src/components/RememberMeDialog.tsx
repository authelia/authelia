// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogTitle } from "@components/UI/Dialog";

export interface Props {
    open: boolean;

    onChoice: (_rememberMe: boolean) => void;
}

// RememberMeDialog asks the user whether the session that has just been established should be remembered. It is
// intended for any sign in flow which completes without the user having first filled in the login form, and therefore
// without them having had the opportunity to tick the remember me checkbox on it.
const RememberMeDialog = function (props: Props) {
    const { t: translate } = useTranslation();

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) props.onChoice(false);
            }}
        >
            <DialogContent id="remember-me-dialog" className="sm:max-w-md" showCloseButton={false}>
                <DialogTitle>{translate("Remember me?")}</DialogTitle>
                <DialogDescription>{translate("Would you like to stay signed in on this device?")}</DialogDescription>
                <DialogFooter>
                    <Button id="dialog-remember-me-no" variant="outline" onClick={() => props.onChoice(false)}>
                        {translate("No")}
                    </Button>
                    <Button id="dialog-remember-me-yes" variant="default" onClick={() => props.onChoice(true)}>
                        {translate("Yes")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default RememberMeDialog;
