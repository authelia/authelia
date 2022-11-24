import React from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";

import { WebauthnDevice } from "@root/models/Webauthn";

interface Props {
    open: boolean;
    device: WebauthnDevice | undefined;
    handleClose: (ok: boolean) => void;
}

export default function WebauthnDeviceDeleteDialog(props: Props) {
    const handleCancel = () => {
        props.handleClose(false);
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{`Remove ${props.device ? props.device.description : "(unknown)"}`}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {`Are you sure you want to remove the device "${
                        props.device ? props.device.description : "(unknown)"
                    }" from your account?`}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancel}>Cancel</Button>
                <Button
                    onClick={() => {
                        props.handleClose(true);
                    }}
                    autoFocus
                >
                    Remove
                </Button>
            </DialogActions>
        </Dialog>
    );
}
