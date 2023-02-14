import React from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

import { WebauthnDevice } from "@models/Webauthn";

interface Props {
    open: boolean;
    device: WebauthnDevice;
    handleClose: (ok: boolean) => void;
}

export default function WebauthnDeviceDeleteDialog(props: Props) {
    const { t: translate } = useTranslation("settings");

    const handleCancel = () => {
        props.handleClose(false);
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{translate("Remove Webauthn Credential")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {translate("Are you sure you want to remove the Webauthn credential from from your account", {
                        description: props.device.displayname,
                    })}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancel}>{translate("Cancel")}</Button>
                <Button
                    onClick={() => {
                        props.handleClose(true);
                    }}
                    autoFocus
                >
                    {translate("Remove")}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
