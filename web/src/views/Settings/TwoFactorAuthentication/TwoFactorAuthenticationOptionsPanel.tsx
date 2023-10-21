import React, { ChangeEvent, Fragment, useEffect, useState } from "react";

import { FormControl, FormControlLabel, FormLabel, Paper, Radio, RadioGroup, Typography } from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2/Grid2";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { Configuration } from "@models/Configuration";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { Method2FA, isMethod2FA, setPreferred2FAMethod, toMethod2FA, toSecondFactorMethod } from "@services/UserInfo";

interface Props {
    refresh: () => void;
    config?: Configuration;
    info?: UserInfo;
}

const TwoFactorAuthenticationOptionsPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const { createErrorNotification } = useNotifications();

    const [method, setMethod] = useState<undefined | string>(undefined);
    const [methods, setMethods] = useState<string[]>([]);

    const hasMethods =
        props.info !== undefined && (props.info.has_totp || props.info.has_webauthn || props.info.has_duo);

    useEffect(() => {
        if (props.info === undefined) return;

        setMethod(toMethod2FA(props.info.method));
    }, [props.info]);

    useEffect(() => {
        if (!props.config || !hasMethods) return;
        let valuesFinal: string[] = [];

        const values = Array.from(props.config.available_methods);

        values.forEach((value) => {
            const v = toMethod2FA(value);

            if (!valuesFinal.includes(v)) {
                switch (value) {
                    case SecondFactorMethod.WebAuthn:
                        valuesFinal.push(v);
                        break;
                    case SecondFactorMethod.TOTP:
                        valuesFinal.push(v);
                        break;
                    case SecondFactorMethod.MobilePush:
                        valuesFinal.push(v);
                        break;
                }
            }
        });

        setMethods(valuesFinal);
    }, [props.config, hasMethods]);

    const handleMethodChanged = (event: ChangeEvent<HTMLInputElement>) => {
        console.log(event.target.value);

        if (isMethod2FA(event.target.value)) {
            const value = toSecondFactorMethod(event.target.value as Method2FA);

            setPreferred2FAMethod(value)
                .catch((err) => {
                    console.error(err);
                    createErrorNotification("There was an issue updating preferred second factor method");
                })
                .then(() => {
                    setMethod(event.target.value);
                })
                .finally(() => {
                    props.refresh();
                });
        }
    };

    useEffect(() => {
        console.table(props.info);
        console.table(props.config);
    }, [props.config, props.info]);

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
                                        <FormControl>
                                            <FormLabel id={"two-factor-method-label"}>
                                                {translate("Default Method")}
                                            </FormLabel>
                                            <RadioGroup value={method} onChange={handleMethodChanged} row>
                                                {methods.map((value, index) => {
                                                    switch (value) {
                                                        case "webauthn":
                                                            return (
                                                                <FormControlLabel
                                                                    control={<Radio />}
                                                                    label={translate("WebAuthn")}
                                                                    key={index}
                                                                    value={value}
                                                                />
                                                            );
                                                        case "totp":
                                                            return (
                                                                <FormControlLabel
                                                                    control={<Radio />}
                                                                    label={translate("One-Time Password")}
                                                                    key={index}
                                                                    value={value}
                                                                />
                                                            );
                                                        case "mobile_push":
                                                            return (
                                                                <FormControlLabel
                                                                    control={<Radio />}
                                                                    label={translate("Mobile Push")}
                                                                    key={index}
                                                                    value={value}
                                                                />
                                                            );
                                                        default:
                                                            return <Fragment />;
                                                    }
                                                })}
                                            </RadioGroup>
                                        </FormControl>
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
