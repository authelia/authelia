import React, { useState } from "react";

import { Box, Button, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { RegisterWebauthnRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { WebauthnDevice } from "@root/models/Webauthn";
import { initiateWebauthnRegistrationProcess } from "@services/RegisterDevice";
import { AutheliaState, AuthenticationLevel } from "@services/State";

import WebauthnDeviceItem from "./WebauthnDeviceItem";

interface Props {
    state: AutheliaState;
    webauthnDevices: WebauthnDevice[] | undefined;
}

export default function TwoFactorAuthSettings(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [webauthnShowDetails, setWebauthnShowDetails] = useState<number>(-1);
    const [registrationInProgress, setRegistrationInProgress] = useState(false);

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
                    <Box>
                        <Typography variant="h5">Webauthn Devices</Typography>
                    </Box>
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
                                {props.webauthnDevices
                                    ? props.webauthnDevices.map((x, idx) => {
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
