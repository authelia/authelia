import { useTranslation } from "react-i18next";

import { useNotifications } from "@contexts/NotificationsContext";
import { OpenIDConnectLink } from "@models/OpenIDConnectRelyingParty";
import { deleteOpenIDConnectLink } from "@services/OpenIDConnectRelyingParty";
import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";

interface Props {
    open: boolean;
    link?: OpenIDConnectLink;
    handleClose: () => void;
}

const OpenIDConnectLinkDeleteDialog = function (props: Props) {
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
            await deleteOpenIDConnectLink(props.link.id);
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

export default OpenIDConnectLinkDeleteDialog;
