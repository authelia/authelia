import { ChangeEvent } from "react";

import { useTranslation } from "react-i18next";

import { Label } from "@components/UI/Label";
import { RadioGroup, RadioGroupItem } from "@components/UI/RadioGroup";
import { SecondFactorMethod } from "@models/Methods";
import { toMethod2FA } from "@services/UserInfo";

interface Props {
    id: string;
    methods: SecondFactorMethod[];
    method: SecondFactorMethod;
    name: string;
    handleMethodChanged: (_event: ChangeEvent<HTMLInputElement>) => void;
}

const TwoFactorAuthenticationOptionsMethodsRadioGroup = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const handleValueChange = (value: string) => {
        const syntheticEvent = {
            target: { value },
        } as ChangeEvent<HTMLInputElement>;
        props.handleMethodChanged(syntheticEvent);
    };

    return (
        <fieldset>
            <legend className="text-sm font-medium text-muted-foreground mb-3">{translate(props.name)}</legend>
            <RadioGroup
                value={toMethod2FA(props.method)}
                onValueChange={handleValueChange}
                className="flex flex-row gap-4"
            >
                {props.methods.map((value, _index) => {
                    const v = toMethod2FA(value);

                    switch (value) {
                        case SecondFactorMethod.WebAuthn:
                            return (
                                <div
                                    key={v}
                                    className="flex items-center gap-2"
                                    id={`method-${props.id}-default-webauthn`}
                                >
                                    <RadioGroupItem value={v} id={`${props.id}-webauthn`} />
                                    <Label htmlFor={`${props.id}-webauthn`}>{translate("WebAuthn")}</Label>
                                </div>
                            );
                        case SecondFactorMethod.TOTP:
                            return (
                                <div
                                    key={v}
                                    className="flex items-center gap-2"
                                    id={`method-${props.id}-default-one-time-password`}
                                >
                                    <RadioGroupItem value={v} id={`${props.id}-otp`} />
                                    <Label htmlFor={`${props.id}-otp`}>{translate("One-Time Password")}</Label>
                                </div>
                            );
                        case SecondFactorMethod.MobilePush:
                            return (
                                <div key={v} className="flex items-center gap-2" id={`method-${props.id}-default-duo`}>
                                    <RadioGroupItem value={v} id={`${props.id}-duo`} />
                                    <Label htmlFor={`${props.id}-duo`}>{translate("Mobile Push")}</Label>
                                </div>
                            );
                        default:
                            return null;
                    }
                })}
            </RadioGroup>
        </fieldset>
    );
};

export default TwoFactorAuthenticationOptionsMethodsRadioGroup;
