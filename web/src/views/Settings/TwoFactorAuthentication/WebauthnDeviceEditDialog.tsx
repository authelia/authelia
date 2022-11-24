import React, { MutableRefObject, useRef, useState } from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

import FixedTextField from "@components/FixedTextField";
import { WebauthnDevice } from "@root/models/Webauthn";

interface Props {
    open: boolean;
    device: WebauthnDevice | undefined;
    handleClose: (ok: boolean, name: string) => void;
}

export default function WebauthnDeviceEditDialog(props: Props) {
    const { t: translate } = useTranslation();
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
            <DialogTitle>{`Edit ${props.device ? props.device.description : "(unknown)"}`}</DialogTitle>
            <DialogContent>
                <DialogContentText>Enter a new name for this device:</DialogContentText>
                <FixedTextField
                    // TODO (PR: #806, Issue: #511) potentially refactor
                    autoFocus
                    inputRef={nameRef}
                    id="name-textfield"
                    label={translate("Name")}
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
                    onKeyPress={(ev) => {
                        if (ev.key === "Enter") {
                            handleConfirm();
                            ev.preventDefault();
                        }
                    }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancel}>Cancel</Button>
                <Button onClick={handleConfirm}>Update</Button>
            </DialogActions>
        </Dialog>
    );
}
