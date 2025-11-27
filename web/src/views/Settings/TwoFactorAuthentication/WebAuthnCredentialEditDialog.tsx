import { useRef, useState } from "react";

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
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [credentialDescription, setCredentialDescription] = useState("");
    const descriptionRef = useRef<HTMLInputElement>(null);
    const [errorDescription, setErrorDescription] = useState(false);

    const handleReset = () => {
        setErrorDescription(false);
        setCredentialDescription("");
    };

    const handleUpdate = () => {
        if (credentialDescription.length === 0) {
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
            createErrorNotification(translate("An error occurred when attempting to update the WebAuthn Credential"));
            return;
        }

        const response = await updateUserWebAuthnCredential(props.credential.id, name);

        if (!response) {
            createErrorNotification(translate("An error occurred when attempting to update the WebAuthn Credential"));
            return;
        }

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(
                    translate("You must be elevated to {{action}} a {{item}}", {
                        action: translate("update"),
                        item: translate("WebAuthn Credential"),
                    }),
                );
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to {{action}} a {{item}}", {
                        action: translate("update"),
                        item: translate("WebAuthn Credential"),
                    }),
                );
            } else {
                createErrorNotification(
                    translate("There was a problem {{action}} the {{item}}", {
                        action: translate("updating"),
                        item: translate("WebAuthn Credential"),
                    }),
                );
            }

            return;
        }

        createSuccessNotification(
            translate("Successfully {{action}} the {{item}}", {
                action: translate("updated"),
                item: translate("WebAuthn Credential"),
            }),
        );

        handleReset();
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{translate("Edit WebAuthn Credential")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {translate("Enter a new description for this WebAuthn Credential")}
                </DialogContentText>
                <TextField
                    inputRef={descriptionRef}
                    id="webauthn-credential-description"
                    label={translate("Description")}
                    variant="standard"
                    required
                    value={credentialDescription}
                    error={errorDescription}
                    fullWidth
                    disabled={false}
                    slotProps={{
                        htmlInput: {
                            maxLength: 30,
                        },
                    }}
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
                <Button id={"dialog-cancel"} onClick={handleCancel}>
                    {translate("Cancel")}
                </Button>
                <Button id={"dialog-update"} onClick={handleUpdate}>
                    {translate("Update")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default WebAuthnCredentialEditDialog;
