import { Fingerprint } from "@mui/icons-material";
import { useTranslation } from "react-i18next";

import { WebAuthnCredential } from "@models/WebAuthn";
import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";

interface Props {
    index: number;
    credential: WebAuthnCredential;
    handleInformation: (_index: number) => void;
    handleEdit: (_index: number) => void;
    handleDelete: (_index: number) => void;
}

const WebAuthnCredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

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
        <CredentialItem
            id={`webauthn-credential-${props.index}`}
            icon={<Fingerprint fontSize="large" color={"warning"} />}
            description={props.credential.description}
            qualifier={` (${props.credential.attestation_type.toUpperCase()})`}
            created_at={new Date(props.credential.created_at)}
            problem={props.credential.legacy}
            last_used_at={props.credential.last_used_at ? new Date(props.credential.last_used_at) : undefined}
            tooltipInformation={translate("Display extended information for this WebAuthn Credential")}
            tooltipInformationProblem={translate(
                "There is an issue with this Credential to find out more click to display extended information for this WebAuthn Credential",
            )}
            tooltipEdit={translate("Edit this {{item}}", { item: translate("WebAuthn Credential") })}
            tooltipDelete={translate("Remove this {{item}}", { item: translate("WebAuthn Credential") })}
            handleInformation={handleInformation}
            handleEdit={handleEdit}
            handleDelete={handleDelete}
        />
    );
};

export default WebAuthnCredentialItem;
