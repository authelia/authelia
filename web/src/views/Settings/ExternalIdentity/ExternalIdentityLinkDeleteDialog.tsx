// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useTranslation } from "react-i18next";

import { useNotifications } from "@contexts/NotificationsContext";
import { ExternalIdentityLink } from "@models/ExternalIdentity";
import { deleteExternalIdentityLink } from "@services/ExternalIdentity";
import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";

interface Props {
    open: boolean;
    link?: ExternalIdentityLink;
    handleClose: () => void;
}

const ExternalIdentityLinkDeleteDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const handleCancel = () => {
        props.handleClose();
    };

    const handleRemove = async () => {
        if (!props.link) {
            return;
        }

        const item = `${props.link.provider_name} account`;

        try {
            await deleteExternalIdentityLink(props.link.id);
        } catch (err) {
            console.error(err);

            createErrorNotification(
                translate("There was a problem {{action}} the {{item}}", {
                    action: translate("deleting"),
                    item,
                }),
            );

            return;
        }

        createSuccessNotification(
            translate("Successfully {{action}} the {{item}}", {
                action: translate("deleted"),
                item,
            }),
        );

        props.handleClose();
    };

    return (
        <DeleteDialog
            open={props.open}
            onConfirm={() => handleRemove().catch(console.error)}
            onCancel={handleCancel}
            title={translate("Remove {{item}}", {
                item: props.link ? `${props.link.provider_name} account` : translate("Linked Accounts"),
            })}
            text={translate("Remove the link to your {{name}} account", { name: props.link?.provider_name })}
        />
    );
};

export default ExternalIdentityLinkDeleteDialog;
