import React, { useRef, useState } from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, TextField } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { WebAuthnCredential } from "@models/WebAuthn";
import { updateUserWebAuthnCredential } from "@services/WebAuthn";

interface Props {
    open: boolean;
    credential?: WebAuthnCredential;
    handleClose: () => void;
}

const WebAuthnCredentialEditDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [credentialDescription, setCredentialDescription] = useState("");
    const descriptionRef = useRef<HTMLInputElement>(null);
    const [errorDescription, setErrorDescription] = useState(false);

    const handleReset = () => {
        setErrorDescription(false);
        setCredentialDescription("");
    };

    const handleUpdate = () => {
        if (!credentialDescription.length) {
            setErrorDescription(true);
        } else {
            handleEdit(credentialDescription).catch(console.error);
            props.handleClose();
        }
        handleReset();
    };

    const handleCancel = () => {
        props.handleClose();
        handleReset();
    };

    const handleEdit = async (name: string) => {
        if (!props.credential) {
            createErrorNotification(translate("An error occurred when attempting to update the WebAuthn credential"));
            return;
        }

        const response = await updateUserWebAuthnCredential(props.credential.id, name);

        if (!response) {
            createErrorNotification(translate("An error occurred when attempting to update the WebAuthn credential"));
            return;
        }

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(translate("You must be elevated to update WebAuthn credentials"));
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to update WebAuthn credentials"),
                );
            } else {
                createErrorNotification(translate("There was a problem updating the WebAuthn credential"));
            }

            return;
        }

        createSuccessNotification(translate("Successfully updated the WebAuthn credential"));

        handleReset();
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
                    inputRef={descriptionRef}
                    id="name-textfield"
                    label={translate("Description")}
                    variant="standard"
                    required
                    value={credentialDescription}
                    error={errorDescription}
                    fullWidth
                    disabled={false}
                    inputProps={{ maxLength: 30 }}
                    onChange={(v) => {
                        setCredentialDescription(v.target.value.substring(0, 30));
                        setErrorDescription(false);
                    }}
                    autoCapitalize="none"
                    autoComplete="webauthn-name"
                    onKeyDown={(ev) => {
                        if (ev.key === "Enter") {
                            handleUpdate();
                            ev.preventDefault();
                        }
                    }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancel}>{translate("Cancel")}</Button>
                <Button onClick={handleUpdate}>{translate("Update")}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default WebAuthnCredentialEditDialog;
