// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Link2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { ExternalIdentityLink } from "@models/ExternalIdentity";
import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";

interface Props {
    link: ExternalIdentityLink;
    onDelete: () => void;
}

const ExternalIdentityLinkItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const { link } = props;

    return (
        <CredentialItem
            id={`external-identity-link-${link.id}`}
            icon={<Link2 className="size-7 text-blue-500" />}
            description={link.provider_name}
            qualifier={link.remote_username ? ` (${link.remote_username})` : ""}
            created_at={new Date(link.created_at)}
            last_used_at={link.last_used_at ? new Date(link.last_used_at) : undefined}
            tooltipDelete={translate("Remove the link to your {{name}} account", { name: link.provider_name })}
            handleDelete={props.onDelete}
        />
    );
};

export default ExternalIdentityLinkItem;
