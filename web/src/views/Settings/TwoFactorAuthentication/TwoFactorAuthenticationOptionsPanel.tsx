import React, { ChangeEvent, Fragment, useEffect, useState } from "react";

import { Paper, Typography } from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2/Grid2";
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

const TwoFactorAuthenticationOptionsPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification } = useNotifications();
    const { localStorageMethod, setLocalStorageMethod, localStorageMethodAvailable } = useLocalStorageMethodContext();

    const [method, setMethod] = useState<SecondFactorMethod>();
    const [methods, setMethods] = useState<SecondFactorMethod[]>([]);

    const hasMethods = props.info.has_totp || props.info.has_webauthn || props.info.has_duo;

    useEffect(() => {
        if (props.info === undefined) return;

        setMethod(props.info.method);
    }, [props.info]);

    useEffect(() => {
        if (!hasMethods) return;
        let valuesFinal: SecondFactorMethod[] = [];

        const values = Array.from(props.config.available_methods);

        values.forEach((value) => {
            if (!valuesFinal.includes(value)) {
                switch (value) {
                    case SecondFactorMethod.WebAuthn:
                        if (props.info.has_webauthn) {
                            valuesFinal.push(value);
                        }
                        break;
                    case SecondFactorMethod.TOTP:
                        if (props.info.has_totp) {
                            valuesFinal.push(value);
                        }
                        break;
                    case SecondFactorMethod.MobilePush:
                        if (props.info.has_duo) {
                            valuesFinal.push(value);
                        }
                        break;
                }
            }
        });

        setMethods(valuesFinal);
    }, [props.config, hasMethods, props.info.has_webauthn, props.info.has_totp, props.info.has_duo]);

    const handleMethodAccountChanged = (event: ChangeEvent<HTMLInputElement>) => {
        if (isMethod2FA(event.target.value)) {
            const value = toSecondFactorMethod(event.target.value as Method2FA);

            setPreferred2FAMethod(value)
                .catch((err) => {
                    console.error(err);
                    createErrorNotification(translate("There was an issue updating preferred second factor method"));
                })
                .then(() => {
                    setMethod(value);
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
                        <Grid xs={12}>
                            <Typography variant={"h5"}>{translate("Options")}</Typography>
                        </Grid>
                        <Grid xs={12}>
                            <Grid container spacing={2} padding={2}>
                                {method === undefined ? null : (
                                    <Grid xs={4}>
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
                                    <Grid xs={4}>
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
