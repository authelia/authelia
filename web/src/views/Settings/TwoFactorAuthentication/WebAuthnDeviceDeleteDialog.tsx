import React from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

import { WebAuthnDevice } from "@models/WebAuthn";

interface Props {
    open: boolean;
    device: WebAuthnDevice;
    handleClose: (ok: boolean) => void;
}

export default function WebAuthnDeviceDeleteDialog(props: Props) {
    const { t: translate } = useTranslation("settings");

    const handleCancel = () => {
        props.handleClose(false);
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{translate("Remove WebAuthn Credential")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {translate("Are you sure you want to remove the WebAuthn credential from from your account", {
                        description: props.device.description,
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
