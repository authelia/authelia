import React, { useEffect, useState } from "react";

import {
    Box,
    Button,
    Chip,
    Paper,
    Stack,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { RegisterWebauthnRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { Configuration } from "@models/Configuration";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { WebauthnDevice } from "@models/Webauthn";
import { initiateWebauthnRegistrationProcess } from "@root/services/RegisterDevice";
import { AutheliaState, AuthenticationLevel } from "@root/services/State";
import { getWebauthnDevices } from "@root/services/UserWebauthnDevices";

import WebauthnDeviceItem from "./WebauthnDeviceItem";

interface Props {
    configuration: Configuration;
    state: AutheliaState;
    userInfo: UserInfo;
}

export default function WebauthnDevices(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [webauthnShowDetails, setWebauthnShowDetails] = useState<number>(-1);
    const [registrationInProgress, setRegistrationInProgress] = useState(false);

    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDevice[] | undefined>();

    useEffect(() => {
        (async function () {
            const devices = await getWebauthnDevices();
            setWebauthnDevices(devices);
        })();
    }, []);

    const handleWebAuthnDetailsChange = (idx: number) => {
        if (webauthnShowDetails === idx) {
            setWebauthnShowDetails(-1);
        } else {
            setWebauthnShowDetails(idx);
        }
    };

    const initiateRegistration = async (initiateRegistrationFunc: () => Promise<void>, redirectRoute: string) => {
        if (props.state.authentication_level >= AuthenticationLevel.TwoFactor) {
            navigate(redirectRoute);
        } else {
            if (registrationInProgress) {
                return;
            }
            setRegistrationInProgress(true);
            try {
                await initiateRegistrationFunc();
                createInfoNotification(translate("An email has been sent to your address to complete the process"));
            } catch (err) {
                console.error(err);
                createErrorNotification(translate("There was a problem initiating the registration process"));
            }
            setRegistrationInProgress(false);
        }
    };

    const handleAddKeyButtonClick = () => {
        initiateRegistration(initiateWebauthnRegistrationProcess, RegisterWebauthnRoute);
    };

    return (
        <Paper variant="outlined">
            <Box sx={{ p: 3 }}>
                <Stack spacing={2}>
                    <Stack direction="row" spacing={1}>
                        <Typography variant="h5" style={{ flexGrow: 1 }}>
                            Webauthn Devices
                        </Typography>
                        {!props.configuration.available_methods.has(SecondFactorMethod.Webauthn) && (
                            <Chip label="not available" color="secondary" variant="outlined" />
                        )}
                        {props.userInfo.has_webauthn && (
                            <>
                                <Chip label="enabled" color="primary" variant="outlined" />
                                {props.userInfo.method === SecondFactorMethod.Webauthn && (
                                    <Chip label="default" color="primary" />
                                )}
                            </>
                        )}
                    </Stack>
                    <Box>
                        <Button variant="outlined" color="primary" onClick={handleAddKeyButtonClick}>
                            {"Add new device"}
                        </Button>
                    </Box>
                    <Box>
                        <Table>
                            <TableHead>
                                <TableRow>
                                    <TableCell />
                                    <TableCell>{translate("Name")}</TableCell>
                                    <TableCell>{translate("Enabled")}</TableCell>
                                    <TableCell align="center">{translate("Actions")}</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {webauthnDevices
                                    ? webauthnDevices.map((x, idx) => {
                                          return (
                                              <WebauthnDeviceItem
                                                  device={x}
                                                  idx={idx}
                                                  webauthnShowDetails={webauthnShowDetails}
                                                  handleWebAuthnDetailsChange={handleWebAuthnDetailsChange}
                                                  key={`webauthn-device-${idx}`}
                                              />
                                          );
                                      })
                                    : null}
                            </TableBody>
                        </Table>
                    </Box>
                </Stack>
            </Box>
        </Paper>
    );
}
