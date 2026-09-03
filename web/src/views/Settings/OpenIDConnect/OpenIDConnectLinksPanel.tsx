import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Card, CardContent } from "@components/UI/Card";
import { Spinner } from "@components/UI/Spinner";
import { OpenIDConnectLink, OpenIDConnectProvider } from "@models/OpenIDConnectRelyingParty";
import OpenIDConnectLinkItem from "@views/Settings/OpenIDConnect/OpenIDConnectLinkItem";

interface Props {
    links: OpenIDConnectLink[];
    providers: OpenIDConnectProvider[];
    linking?: string;
    onDelete: (_id: number) => void;
    onLink: (_id: string) => void;
}

const OpenIDConnectLinksPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Card id={"openid-connect-links-panel"}>
            <CardContent className="grid grid-cols-12 gap-4 p-4">
                <div className="col-span-12">
                    <h5 className="text-xl font-semibold">{translate("Linked Accounts")}</h5>
                </div>
                <div className="col-span-12">
                    {props.links.length === 0 ? (
                        <p className="text-sm text-muted-foreground">{translate("No external accounts are linked")}</p>
                    ) : (
                        <div className="grid grid-cols-12 gap-6">
                            {props.links.map((link) => (
                                <div className="col-span-12 md:col-span-6 xl:col-span-3" key={link.id}>
                                    <OpenIDConnectLinkItem link={link} onDelete={() => props.onDelete(link.id)} />
                                </div>
                            ))}
                        </div>
                    )}
                </div>
                {props.providers.length === 0 ? null : (
                    <div className="col-span-12 flex flex-wrap gap-2">
                        {props.providers.map((provider) => (
                            <Button
                                id={`openid-connect-link-start-${provider.id}`}
                                key={provider.id}
                                variant={"outline"}
                                color={"primary"}
                                disabled={props.linking !== undefined}
                                onClick={() => props.onLink(provider.id)}
                            >
                                {translate("Link your {{name}} account", { name: provider.name })}
                                {props.linking === provider.id ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                            </Button>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
};

export default OpenIDConnectLinksPanel;
