import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Card, CardContent } from "@components/UI/Card";
import { OpenIDConnectPendingLink } from "@models/OpenIDConnectRelyingParty";

interface Props {
    pending: OpenIDConnectPendingLink;
    onAccept: () => void;
    onDecline: () => void;
}

const OpenIDConnectLinkPendingPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const { pending } = props;
    const remoteIdentity = pending.display_name || pending.remote_username;

    return (
        <Card id={"openid-connect-link-pending-panel"}>
            <CardContent className="grid grid-cols-12 gap-4 p-4">
                <div className="col-span-12">
                    <h5 className="text-xl font-semibold">
                        {translate("Link your {{name}} account", { name: pending.provider_name })}
                    </h5>
                </div>
                <div className="col-span-12 flex flex-col gap-1">
                    <span className="font-bold">{pending.provider_name}</span>
                    {remoteIdentity ? <span className="text-sm">{remoteIdentity}</span> : null}
                    {pending.email ? <span className="text-sm text-muted-foreground">{pending.email}</span> : null}
                    <span className="text-xs text-muted-foreground">{pending.subject}</span>
                </div>
                <div className="col-span-12 flex gap-2">
                    <Button
                        id={"openid-connect-link-accept"}
                        variant={"outline"}
                        color={"primary"}
                        onClick={props.onAccept}
                    >
                        {translate("Accept")}
                    </Button>
                    <Button
                        id={"openid-connect-link-decline"}
                        variant={"outline"}
                        color={"destructive"}
                        onClick={props.onDecline}
                    >
                        {translate("Decline")}
                    </Button>
                </div>
            </CardContent>
        </Card>
    );
};

export default OpenIDConnectLinkPendingPanel;
