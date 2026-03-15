import { FC, Fragment, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Checkbox } from "@components/UI/Checkbox";
import { Label } from "@components/UI/Label";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";

export interface Props {
    pre_configuration: boolean;
    onChangePreConfiguration: (_value: boolean) => void;
}

const DecisionFormPreConfiguration: FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    const [preConfigure, setPreConfigure] = useState(false);

    const handlePreConfigureChanged = () => {
        setPreConfigure((preConfigure) => !preConfigure);
    };

    useEffect(() => {
        props.onChangePreConfiguration(preConfigure);
    }, [preConfigure, props]);

    return (
        <Fragment>
            {props.pre_configuration ? (
                <div className="w-full">
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="flex items-center gap-2">
                                    <Checkbox
                                        id="pre-configure"
                                        checked={preConfigure}
                                        onCheckedChange={handlePreConfigureChanged}
                                    />
                                    <Label htmlFor="pre-configure">{translate("Remember Consent")}</Label>
                                </div>
                            </TooltipTrigger>
                            <TooltipContent>
                                {translate("This saves this consent as a pre-configured consent for future use")}
                            </TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                </div>
            ) : null}
        </Fragment>
    );
};

export default DecisionFormPreConfiguration;
