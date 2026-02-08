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
    const { createErrorNotification, createSuccessNotification } = useNotifications();

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
                createErrorNotification(
                    translate("You must be elevated to {{action}} a {{item}}", {
                        action: translate("delete"),
                        item: translate("WebAuthn Credential"),
                    }),
                );
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to {{action}} a {{item}}", {
                        action: translate("delete"),
                        item: "WebAuthn Credential",
                    }),
                );
            } else {
                createErrorNotification(
                    translate("There was a problem {{action}} the {{item}}", {
                        action: translate("deleting"),
                        item: translate("WebAuthn Credential"),
                    }),
                );
            }

            return;
        }

        createSuccessNotification(
            translate("Successfully {{action}} the {{item}}", {
                action: translate("deleted"),
                item: translate("WebAuthn Credential"),
            }),
        );

        props.handleClose();
    };

    return (
        <DeleteDialog
            open={props.open}
            onConfirm={() => handleRemove().catch(console.error)}
            onCancel={handleCancel}
            title={translate("Remove {{item}}", { item: translate("WebAuthn Credential") })}
            text={translate("Are you sure you want to remove the WebAuthn Credential from your account", {
                description: props.credential?.description,
            })}
        />
    );
};

export default WebAuthnCredentialDeleteDialog;
