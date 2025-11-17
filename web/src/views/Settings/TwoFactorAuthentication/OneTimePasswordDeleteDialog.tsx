import React from "react";

import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { deleteUserTOTPConfiguration } from "@services/UserInfoTOTPConfiguration";
import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";

interface Props {
    open: boolean;
    handleClose: () => void;
}

const OneTimePasswordDeleteDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const handleCancel = () => {
        props.handleClose();
    };

    const handleRemove = async () => {
        const response = await deleteUserTOTPConfiguration();

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(
                    translate("You must be elevated to {{action}} a {{item}}", {
                        action: translate("delete"),
                        item: translate("One-Time Password"),
                    }),
                );
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to {{action}} a {{item}}", {
                        action: translate("delete"),
                        item: translate("One-Time Password"),
                    }),
                );
            } else {
                createErrorNotification(
                    translate("There was a problem {{action}} the {{item}}", {
                        action: translate("deleting"),
                        item: translate("One-Time Password"),
                    }),
                );
            }

            return;
        }

        createSuccessNotification(
            translate("Successfully {{action}} the {{item}}", {
                action: translate("deleted"),
                item: translate("One-Time Password"),
            }),
        );

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
            title={translate("Remove {{item}}", { item: translate("One-Time Password") })}
            text={translate("Are you sure you want to remove the One-Time Password from your account")}
        />
    );
};

export default OneTimePasswordDeleteDialog;
