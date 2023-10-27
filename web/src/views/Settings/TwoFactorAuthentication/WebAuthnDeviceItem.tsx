import React, { useState } from "react";

import { Fingerprint } from "@mui/icons-material";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import { CircularProgress, Paper, Stack, Tooltip, Typography } from "@mui/material";
import IconButton from "@mui/material/IconButton";
import Grid from "@mui/material/Unstable_Grid2";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { WebAuthnDevice } from "@models/WebAuthn";
import { deleteUserWebAuthnDevice, updateUserWebAuthnDevice } from "@services/WebAuthn";
import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";
import WebAuthnDeviceDetailsDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnDeviceDetailsDialog";
import WebAuthnDeviceEditDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnDeviceEditDialog";

interface Props {
    index: number;
    device: WebAuthnDevice;
    handleEdit: () => void;
}

export default function WebAuthnDeviceItem(props: Props) {
    const { t: translate } = useTranslation("settings");

    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [showDialogDetails, setShowDialogDetails] = useState<boolean>(false);
    const [showDialogEdit, setShowDialogEdit] = useState<boolean>(false);
    const [showDialogDelete, setShowDialogDelete] = useState<boolean>(false);

    const [loadingEdit, setLoadingEdit] = useState<boolean>(false);
    const [loadingDelete, setLoadingDelete] = useState<boolean>(false);

    const handleEdit = async (ok: boolean, name: string) => {
        setShowDialogEdit(false);

        if (!ok) {
            return;
        }

        setLoadingEdit(true);

        const response = await updateUserWebAuthnDevice(props.device.id, name);

        setLoadingEdit(false);

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(translate("You must be elevated to update WebAuthn credentials"));
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to update WebAuthn credentials"),
                );
            } else {
                createErrorNotification(translate("There was a problem updating the WebAuthn credential"));
            }

            return;
        }

        createSuccessNotification(translate("Successfully updated the WebAuthn credential"));

        props.handleEdit();
    };

    const handleDelete = async (ok: boolean) => {
        setShowDialogDelete(false);

        if (!ok) {
            return;
        }

        setLoadingDelete(true);

        const response = await deleteUserWebAuthnDevice(props.device.id);

        setLoadingDelete(false);

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(translate("You must be elevated to delete WebAuthn credentials"));
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to delete WebAuthn credentials"),
                );
            } else {
                createErrorNotification(translate("There was a problem deleting the WebAuthn credential"));
            }

            return;
        }

        createSuccessNotification(translate("Successfully deleted the WebAuthn credential"));

        props.handleEdit();
    };

    return (
        <Grid xs={12} md={6} xl={3}>
            <WebAuthnDeviceDetailsDialog
                device={props.device}
                open={showDialogDetails}
                handleClose={() => {
                    setShowDialogDetails(false);
                }}
            />
            <WebAuthnDeviceEditDialog device={props.device} open={showDialogEdit} handleClose={handleEdit} />
            <DeleteDialog
                open={showDialogDelete}
                handleClose={handleDelete}
                title={translate("Remove WebAuthn Credential")}
                text={translate("Are you sure you want to remove the WebAuthn credential from from your account", {
                    description: props.device.description,
                })}
            />
            <Paper variant="outlined">
                <Grid container spacing={1} alignItems="center" padding={3}>
                    <Grid xs={12} sm={6} md={6}>
                        <Grid container>
                            <Grid xs={12}>
                                <Stack direction={"row"} spacing={1} alignItems={"center"}>
                                    <Fingerprint fontSize="large" color={"warning"} />
                                    <Typography display="inline" sx={{ fontWeight: "bold" }}>
                                        {props.device.description}
                                    </Typography>
                                    <Typography
                                        display="inline"
                                        variant="body2"
                                    >{` (${props.device.attestation_type.toUpperCase()})`}</Typography>
                                </Stack>
                            </Grid>
                            <Grid xs={12}>
                                <Typography variant={"caption"} sx={{ display: { xs: "none", md: "block" } }}>
                                    {translate("Added when", {
                                        when: new Date(props.device.created_at),
                                        formatParams: {
                                            when: {
                                                hour: "numeric",
                                                minute: "numeric",
                                                year: "numeric",
                                                month: "long",
                                                day: "numeric",
                                            },
                                        },
                                    })}
                                </Typography>
                            </Grid>
                            <Grid xs={12}>
                                <Typography variant={"caption"} sx={{ display: { xs: "none", md: "block" } }}>
                                    {props.device.last_used_at === undefined
                                        ? translate("Never used")
                                        : translate("Last Used when", {
                                              when: new Date(props.device.last_used_at),
                                              formatParams: {
                                                  when: {
                                                      hour: "numeric",
                                                      minute: "numeric",
                                                      year: "numeric",
                                                      month: "long",
                                                      day: "numeric",
                                                  },
                                              },
                                          })}
                                </Typography>
                            </Grid>
                        </Grid>
                    </Grid>
                    <Grid xs={12} md={7} xl={5}>
                        <Stack direction={"row"} spacing={1}>
                            <Tooltip title={translate("Display extended information for this WebAuthn credential")}>
                                <IconButton color="primary" onClick={() => setShowDialogDetails(true)}>
                                    <InfoOutlinedIcon />
                                </IconButton>
                            </Tooltip>
                            <Tooltip title={translate("Edit information for this WebAuthn credential")}>
                                <IconButton
                                    color="primary"
                                    onClick={loadingEdit ? undefined : () => setShowDialogEdit(true)}
                                >
                                    {loadingEdit ? <CircularProgress color="inherit" size={20} /> : <EditIcon />}
                                </IconButton>
                            </Tooltip>
                            <Tooltip title={translate("Remove this WebAuthn credential")}>
                                <IconButton
                                    color="primary"
                                    onClick={loadingDelete ? undefined : () => setShowDialogDelete(true)}
                                >
                                    {loadingDelete ? <CircularProgress color="inherit" size={20} /> : <DeleteIcon />}
                                </IconButton>
                            </Tooltip>
                        </Stack>
                    </Grid>
                </Grid>
            </Paper>
        </Grid>
    );
}
