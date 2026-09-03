import { Link2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { OpenIDConnectLink } from "@models/OpenIDConnectRelyingParty";
import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";

interface Props {
    link: OpenIDConnectLink;
    onDelete: () => void;
}

const OpenIDConnectLinkItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const { link } = props;

    return (
        <CredentialItem
            id={`openid-connect-link-${link.id}`}
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

export default OpenIDConnectLinkItem;
