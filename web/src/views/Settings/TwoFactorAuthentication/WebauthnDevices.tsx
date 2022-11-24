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
import { deleteDevice, updateDevice } from "@root/services/Webauthn";

import WebauthnDeviceDeleteDialog from "./WebauthnDeviceDeleteDialog";
import WebauthnDeviceEditDialog from "./WebauthnDeviceEditDialog";
import WebauthnDeviceItem from "./WebauthnDeviceItem";

interface Props {
    state: AutheliaState;
}

interface WebauthnDeviceDisplay extends WebauthnDevice {
    deleting: boolean;
    editing: boolean;
}

export default function WebauthnDevices(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [webauthnShowDetails, setWebauthnShowDetails] = useState<number>(-1);
    const [deletingIdx, setDeletingIdx] = useState<number>(-1);
    const [editingIdx, setEditingIdx] = useState<number>(-1);
    const [registrationInProgress, setRegistrationInProgress] = useState(false);
    const [ready, setReady] = useState(false);

    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDeviceDisplay[]>([]);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [editDialogOpen, setEditDialogOpen] = useState(false);

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
        setDeletingIdx(idx);
        setDeleteDialogOpen(true);
    };

    const handleDeleteItemConfirm = async (ok: boolean) => {
        setDeleteDialogOpen(false);
        const idx = deletingIdx;
        if (ok !== true) {
            return;
        }
        webauthnDevices[idx].deleting = true;
        const status = await deleteDevice(webauthnDevices[idx].id);
        if (status !== 200) {
            webauthnDevices[idx].deleting = false;
            createErrorNotification(translate("There was a problem deleting the device"));
            return;
        }
        let updatedDevices = [...webauthnDevices];
        updatedDevices.splice(idx, 1);
        setWebauthnDevices(updatedDevices);
    };

    const handleEditItem = async (idx: number) => {
        setEditingIdx(idx);
        setEditDialogOpen(true);
    };

    const handleEditItemConfirm = async (ok: boolean, name: string) => {
        setEditDialogOpen(false);
        const idx = editingIdx;
        if (ok !== true) {
            return;
        }
        webauthnDevices[idx].editing = true;
        const status = await updateDevice(webauthnDevices[idx].id, name);
        webauthnDevices[idx].editing = false;
        if (status !== 200) {
            createErrorNotification(translate("There was a problem updating the device"));
            return;
        }
        let updatedDevices = [...webauthnDevices];
        updatedDevices[idx].description = name;
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
        <>
            <WebauthnDeviceEditDialog
                device={editingIdx > -1 ? webauthnDevices[editingIdx] : undefined}
                open={editDialogOpen}
                handleClose={handleEditItemConfirm}
            />
            <WebauthnDeviceDeleteDialog
                device={deletingIdx > -1 ? webauthnDevices[deletingIdx] : undefined}
                open={deleteDialogOpen}
                handleClose={handleDeleteItemConfirm}
            />
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
                                                            editing={x.editing}
                                                            webauthnShowDetails={webauthnShowDetails === idx}
                                                            handleWebAuthnDetailsChange={() => {
                                                                handleWebAuthnDetailsChange(idx);
                                                            }}
                                                            handleEdit={() => {
                                                                handleEditItem(idx);
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
        </>
    );
}
