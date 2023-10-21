import React, { Fragment, useState } from "react";

import { Fingerprint } from "@mui/icons-material";
import { useTranslation } from "react-i18next";

import { WebAuthnCredential } from "@models/WebAuthn";
import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";
import WebAuthnCredentialInformationDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialInformationDialog";

interface Props {
    index: number;
    credential: WebAuthnCredential;
    handleInformation: (index: number) => void;
    handleEdit: (index: number) => void;
    handleDelete: (index: number) => void;
}

const WebAuthnCredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [showDialogDetails, setShowDialogDetails] = useState<boolean>(false);

    const handleInformation = () => {
        props.handleInformation(props.index);
    };

    const handleEdit = () => {
        props.handleEdit(props.index);
    };

    const handleDelete = () => {
        props.handleDelete(props.index);
    };

    return (
        <Fragment>
            <WebAuthnCredentialInformationDialog
                credential={props.credential}
                open={showDialogDetails}
                handleClose={() => {
                    setShowDialogDetails(false);
                }}
            />
            <CredentialItem
                icon={<Fingerprint fontSize="large" color={"warning"} />}
                description={props.credential.description}
                qualifier={` (${props.credential.attestation_type.toUpperCase()})`}
                created_at={new Date(props.credential.created_at)}
                last_used_at={props.credential.last_used_at ? new Date(props.credential.last_used_at) : undefined}
                tooltipInformation={translate("Display extended information for this WebAuthn credential")}
                tooltipEdit={translate("Edit this WebAuthn credential")}
                tooltipDelete={translate("Remove this WebAuthn credential")}
                handleInformation={handleInformation}
                handleEdit={handleEdit}
                handleDelete={handleDelete}
            />
        </Fragment>
    );
};

export default WebAuthnCredentialItem;
