import React, { MutableRefObject, useRef, useState } from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, TextField } from "@mui/material";
import { useTranslation } from "react-i18next";

import { WebAuthnDevice } from "@models/WebAuthn";

interface Props {
    open: boolean;
    device: WebAuthnDevice;
    handleClose: (ok: boolean, name: string) => void;
}
const WebAuthnDeviceEditDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [deviceName, setName] = useState("");
    const nameRef = useRef() as MutableRefObject<HTMLInputElement>;
    const [nameError, setNameError] = useState(false);

    const handleConfirm = () => {
        if (!deviceName.length) {
            setNameError(true);
        } else {
            props.handleClose(true, deviceName);
        }
        setName("");
    };

    const handleCancel = () => {
        props.handleClose(false, "");
        setName("");
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{translate("Edit WebAuthn Credential")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {translate("Enter a new description for this WebAuthn credential")}
                </DialogContentText>
                <TextField
                    autoFocus
                    inputRef={nameRef}
                    id="name-textfield"
                    label={translate("Description")}
                    variant="standard"
                    required
                    value={deviceName}
                    error={nameError}
                    fullWidth
                    disabled={false}
                    onChange={(v) => setName(v.target.value.substring(0, 30))}
                    onFocus={() => {
                        setNameError(false);
                    }}
                    autoCapitalize="none"
                    autoComplete="webauthn-name"
                    onKeyDown={(ev) => {
                        if (ev.key === "Enter") {
                            handleConfirm();
                            ev.preventDefault();
                        }
                    }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancel}>{translate("Cancel")}</Button>
                <Button onClick={handleConfirm}>{translate("Update")}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default WebAuthnDeviceEditDialog;
