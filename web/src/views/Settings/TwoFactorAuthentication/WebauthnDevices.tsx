import React, { useEffect, useState } from "react";

import {
    Box,
    Button,
    Paper,
    Skeleton,
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
import { WebauthnDevice } from "@root/models/Webauthn";
import { initiateWebauthnRegistrationProcess } from "@root/services/RegisterDevice";
import { AutheliaState, AuthenticationLevel } from "@root/services/State";
import { getWebauthnDevices } from "@root/services/UserWebauthnDevices";
import { deleteDevice } from "@root/services/Webauthn";

import WebauthnDeviceItem from "./WebauthnDeviceItem";

interface Props {
    state: AutheliaState;
}

interface WebauthnDeviceDisplay extends WebauthnDevice {
    deleting: boolean;
}

export default function WebauthnDevices(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [webauthnShowDetails, setWebauthnShowDetails] = useState<number>(-1);
    const [registrationInProgress, setRegistrationInProgress] = useState(false);
    const [ready, setReady] = useState(false);

    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDeviceDisplay[]>([]);

    useEffect(() => {
        (async function () {
            const devices = await getWebauthnDevices();
            const devicesDisplay = devices.map((x, idx) => {
                return { ...x, deleting: false } as WebauthnDeviceDisplay;
            });
            setWebauthnDevices(devicesDisplay);
            setReady(true);
        })();
    }, []);

    const handleWebAuthnDetailsChange = (idx: number) => {
        if (webauthnShowDetails === idx) {
            setWebauthnShowDetails(-1);
        } else {
            setWebauthnShowDetails(idx);
        }
    };

    const handleDeleteItem = async (idx: number) => {
        webauthnDevices[idx].deleting = true;
        const status = await deleteDevice(webauthnDevices[idx].id);
        webauthnDevices[idx].deleting = false;
        if (status !== 200) {
            createErrorNotification(translate("There was a problem deleting the device"));
            return;
        }
        let updatedDevices = [...webauthnDevices];
        updatedDevices.splice(idx, 1);
        setWebauthnDevices(updatedDevices);
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
                        {ready ? (
                            <>
                                {webauthnDevices ? (
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
                                            {webauthnDevices.map((x, idx) => {
                                                return (
                                                    <WebauthnDeviceItem
                                                        device={x}
                                                        deleting={x.deleting}
                                                        webauthnShowDetails={webauthnShowDetails === idx}
                                                        handleWebAuthnDetailsChange={() => {
                                                            handleWebAuthnDetailsChange(idx);
                                                        }}
                                                        handleDelete={() => {
                                                            handleDeleteItem(idx);
                                                        }}
                                                        key={`webauthn-device-${idx}`}
                                                    />
                                                );
                                            })}
                                        </TableBody>
                                    </Table>
                                ) : null}
                            </>
                        ) : (
                            <>
                                <Skeleton height={20} />
                                <Skeleton height={40} />
                            </>
                        )}
                    </Box>
                </Stack>
            </Box>
        </Paper>
    );
}
