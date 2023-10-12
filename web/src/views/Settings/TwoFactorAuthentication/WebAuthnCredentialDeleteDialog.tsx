import React from "react";

import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { WebAuthnCredential } from "@models/WebAuthn";
import { deleteUserWebAuthnCredential } from "@services/WebAuthn";
import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";

interface Props {
    open: boolean;
    credential?: WebAuthnCredential;
    handleClose: () => void;
}

const WebAuthnCredentialDeleteDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const handleCancel = () => {
        props.handleClose();
    };

    const handleRemove = async () => {
        if (!props.credential) {
            return;
        }

        const response = await deleteUserWebAuthnCredential(props.credential.id);

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(translate("You must be elevated to delete WebAuthn credentials"));
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to delete WebAuthn credentials"),
                );
            } else {
                createErrorNotification(translate("There was a problem deleting the WebAuthn credential"));
            }

            return;
        }

        createSuccessNotification(translate("Successfully deleted the WebAuthn credential"));

        props.handleClose();
    };

    const handleClose = (ok: boolean) => {
        if (ok) {
            handleRemove().catch(console.error);
        } else {
            handleCancel();
        }
    };

    return (
        <DeleteDialog
            open={props.open}
            handleClose={handleClose}
            title={translate("Remove WebAuthn Credential")}
            text={translate("Are you sure you want to remove the WebAuthn credential from from your account", {
                description: props.credential?.description,
            })}
        />
    );
};

export default WebAuthnCredentialDeleteDialog;
