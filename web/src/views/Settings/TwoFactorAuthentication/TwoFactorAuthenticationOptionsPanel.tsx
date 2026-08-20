import { ChangeEvent, Fragment, useEffect, useMemo, useReducer } from "react";

import { useTranslation } from "react-i18next";

import { Card, CardContent } from "@components/UI/Card";
import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useNotifications } from "@contexts/NotificationsContext";
import { Configuration } from "@models/Configuration";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { Method2FA, isMethod2FA, setPreferred2FAMethod, toSecondFactorMethod } from "@services/UserInfo";
import TwoFactorAuthenticationOptionsMethodsRadioGroup from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsMethodsRadioGroup";

interface Props {
    refresh: () => void;
    config: Configuration;
    info: UserInfo;
}

type ComponentState = {
    method: SecondFactorMethod | undefined;
};

type Action = { type: "setMethod"; method: SecondFactorMethod };

const initialState: ComponentState = {
    method: undefined,
};

function reducer(state: ComponentState, action: Action): ComponentState {
    if (action.type === "setMethod") {
        return { ...state, method: action.method };
    }
    return state;
}

const TwoFactorAuthenticationOptionsPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification } = useNotifications();
    const { localStorageMethod, localStorageMethodAvailable, setLocalStorageMethod } = useLocalStorageMethodContext();

    const [state, dispatch] = useReducer(reducer, initialState);
    const { method } = state;

    const hasMethods = props.info.has_totp || props.info.has_webauthn || props.info.has_duo;

    useEffect(() => {
        if (props.info === undefined) return;

        dispatch({ method: props.info.method, type: "setMethod" });
    }, [props.info]);

    const methods = useMemo(() => {
        if (!hasMethods) return [];

        return Array.from(props.config.available_methods).filter((method) => {
            switch (method) {
                case SecondFactorMethod.WebAuthn:
                    return props.info.has_webauthn;
                case SecondFactorMethod.TOTP:
                    return props.info.has_totp;
                case SecondFactorMethod.MobilePush:
                    return props.info.has_duo;
                default:
                    return false;
            }
        });
    }, [props.config, hasMethods, props.info.has_webauthn, props.info.has_totp, props.info.has_duo]);

    const handleMethodAccountChanged = (event: ChangeEvent<HTMLInputElement>) => {
        if (isMethod2FA(event.target.value)) {
            const value = toSecondFactorMethod(event.target.value as Method2FA);

            setPreferred2FAMethod(value)
                .then(() => {
                    dispatch({ method: value, type: "setMethod" });
                })
                .catch((err) => {
                    console.error(err);
                    createErrorNotification(translate("There was an issue updating preferred second factor method"));
                })
                .finally(() => {
                    props.refresh();
                });
        }
    };

    const handleMethodBrowserChanged = (event: ChangeEvent<HTMLInputElement>) => {
        if (isMethod2FA(event.target.value)) {
            setLocalStorageMethod(toSecondFactorMethod(event.target.value as Method2FA));
        }
    };

    return (
        <Fragment>
            {!props.config || !hasMethods ? undefined : (
                <Card>
                    <CardContent className="p-4 space-y-4">
                        <div className="w-full">
                            <h5 className="text-xl font-semibold">{translate("Options")}</h5>
                        </div>
                        <div className="w-full">
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 p-4">
                                {method === undefined ? null : (
                                    <div>
                                        <TwoFactorAuthenticationOptionsMethodsRadioGroup
                                            id={"account"}
                                            name={"Default Method"}
                                            method={method}
                                            methods={methods}
                                            handleMethodChanged={handleMethodAccountChanged}
                                        />
                                    </div>
                                )}
                                {!localStorageMethodAvailable || localStorageMethod === undefined ? null : (
                                    <div>
                                        <TwoFactorAuthenticationOptionsMethodsRadioGroup
                                            id={"local"}
                                            name={"Default Method (Browser)"}
                                            method={localStorageMethod}
                                            methods={methods}
                                            handleMethodChanged={handleMethodBrowserChanged}
                                        />
                                    </div>
                                )}
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}
        </Fragment>
    );
};

export default TwoFactorAuthenticationOptionsPanel;
