import { ChangeEvent, Fragment, useEffect, useMemo, useReducer } from "react";

import { Paper, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useNotifications } from "@hooks/NotificationsContext";
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
    const { localStorageMethod, setLocalStorageMethod, localStorageMethodAvailable } = useLocalStorageMethodContext();

    const [state, dispatch] = useReducer(reducer, initialState);
    const { method } = state;

    const hasMethods = props.info.has_totp || props.info.has_webauthn || props.info.has_duo;

    useEffect(() => {
        if (props.info === undefined) return;

        dispatch({ type: "setMethod", method: props.info.method });
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
                    dispatch({ type: "setMethod", method: value });
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
                <Paper variant={"outlined"}>
                    <Grid container spacing={2} padding={2}>
                        <Grid size={{ xs: 12 }}>
                            <Typography variant={"h5"}>{translate("Options")}</Typography>
                        </Grid>
                        <Grid size={{ xs: 12 }}>
                            <Grid container spacing={2} padding={2}>
                                {method === undefined ? null : (
                                    <Grid size={{ xs: 12, md: 4 }}>
                                        <TwoFactorAuthenticationOptionsMethodsRadioGroup
                                            id={"account"}
                                            name={"Default Method"}
                                            method={method}
                                            methods={methods}
                                            handleMethodChanged={handleMethodAccountChanged}
                                        />
                                    </Grid>
                                )}
                                {!localStorageMethodAvailable || localStorageMethod === undefined ? null : (
                                    <Grid size={{ xs: 12, md: 4 }}>
                                        <TwoFactorAuthenticationOptionsMethodsRadioGroup
                                            id={"local"}
                                            name={"Default Method (Browser)"}
                                            method={localStorageMethod}
                                            methods={methods}
                                            handleMethodChanged={handleMethodBrowserChanged}
                                        />
                                    </Grid>
                                )}
                            </Grid>
                        </Grid>
                    </Grid>
                </Paper>
            )}
        </Fragment>
    );
};

export default TwoFactorAuthenticationOptionsPanel;
