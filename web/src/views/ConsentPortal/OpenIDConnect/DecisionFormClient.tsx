import { AppWindow } from "lucide-react";
import { useTranslation } from "react-i18next";

export interface Props {
    client_id: string;
    client_description: string;
}

function DecisionFormClient({ client_description, client_id }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    const named = client_description !== "";

    return (
        <div className="flex w-full items-center gap-3 text-left">
            <span className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                <AppWindow className="size-5" />
            </span>
            <span className="min-w-0 flex-1">
                <span
                    id={"openid-consent-client-name"}
                    data-testid={"openid-consent-client-name"}
                    className="block truncate font-semibold"
                >
                    {named ? client_description : client_id}
                </span>
                {named ? (
                    <span
                        id={"openid-consent-client-id"}
                        className="block truncate font-mono text-xs text-muted-foreground"
                        title={translate("Client ID", { client_id }) || `Client ID: ${client_id}`}
                    >
                        {client_id}
                    </span>
                ) : null}
            </span>
        </div>
    );
}

export default DecisionFormClient;
