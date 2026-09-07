import { useTranslation } from "react-i18next";

import { Item, ItemActions, ItemContent, ItemDescription, ItemTitle } from "@components/UI/Item";
import { Label } from "@components/UI/Label";
import { Switch } from "@components/UI/Switch";

export interface Props {
    checked: boolean;
    onChangePreConfiguration: (_value: boolean) => void;
    pre_configuration: boolean;
}

function DecisionFormPreConfiguration({ checked, onChangePreConfiguration, pre_configuration }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    if (!pre_configuration) {
        return null;
    }

    return (
        <Item variant={"muted"} size={"sm"} className="w-full text-left">
            <ItemContent>
                <ItemTitle>
                    <Label htmlFor="pre-configure">{translate("Remember Consent")}</Label>
                </ItemTitle>
                <ItemDescription className="text-xs">
                    {translate("This saves this consent as a pre-configured consent for future use")}
                </ItemDescription>
            </ItemContent>
            <ItemActions>
                <Switch
                    id="pre-configure"
                    checked={checked}
                    onCheckedChange={(value) => onChangePreConfiguration(value)}
                    aria-label={translate("Remember Consent")}
                />
            </ItemActions>
        </Item>
    );
}

export default DecisionFormPreConfiguration;
