// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Smartphone } from "lucide-react";
import { useTranslation } from "react-i18next";

import { AppAuthenticatorDevice, appAuthenticatorAttestationStatus } from "@services/AppAuthenticator";
import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";

interface Props {
    index: number;
    device: AppAuthenticatorDevice;
    handleEdit: (_index: number) => void;
    handleDelete: (_index: number) => void;
}

const AppAuthenticatorDeviceItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const id = `app-authenticator-device-${props.index}`;
    const attestation = appAuthenticatorAttestationStatus(props.device);

    let attestationLabel: string;
    switch (attestation.kind) {
        case "official":
            attestationLabel = translate("Verified Authelia app");
            break;
        case "verified":
            attestationLabel = translate("Verified: {{identity}}", { identity: attestation.identity });
            break;
        default:
            attestationLabel = translate("Not attested");
    }

    return (
        <CredentialItem
            id={id}
            icon={<Smartphone className="size-7 text-sky-500" />}
            description={props.device.name}
            qualifier=""
            details={
                <div className="flex flex-col">
                    <span id={`${id}-model`} className="text-sm">
                        {`${props.device.model}, ${props.device.platform}`}
                    </span>
                    <span id={`${id}-attestation`} className="text-xs text-muted-foreground">
                        {attestationLabel}
                    </span>
                </div>
            }
            created_at={new Date(props.device.created_at)}
            last_used_at={props.device.last_used_at ? new Date(props.device.last_used_at) : undefined}
            tooltipEdit={translate("Edit this {{item}}", { item: translate("Authelia app device") })}
            tooltipDelete={translate("Remove this {{item}}", { item: translate("Authelia app device") })}
            handleEdit={() => props.handleEdit(props.index)}
            handleDelete={() => props.handleDelete(props.index)}
        />
    );
};

export default AppAuthenticatorDeviceItem;
