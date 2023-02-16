import React, { Fragment, useState } from "react";

import { Fingerprint } from "@mui/icons-material";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import { Box, Button, CircularProgress, Paper, Stack, Tooltip, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { WebauthnDevice } from "@models/Webauthn";
import { deleteDevice, updateDevice } from "@services/Webauthn";
import WebauthnDeviceDeleteDialog from "@views/Settings/TwoFactorAuthentication/WebauthnDeviceDeleteDialog";
import WebauthnDeviceDetailsDialog from "@views/Settings/TwoFactorAuthentication/WebauthnDeviceDetailsDialog";
import WebauthnDeviceEditDialog from "@views/Settings/TwoFactorAuthentication/WebauthnDeviceEditDialog";

interface Props {
    index: number;
    device: WebauthnDevice;
    handleEdit: () => void;
}

export default function WebauthnDeviceItem(props: Props) {
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

        const response = await updateDevice(props.device.id, name);

        setLoadingEdit(false);

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(translate("You must be elevated to update Webauthn credentials"));
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to update Webauthn credentials"),
                );
            } else {
                createErrorNotification(translate("There was a problem updating the Webauthn credential"));
            }

            return;
        }

        createSuccessNotification(translate("Successfully updated the Webauthn credential"));

        props.handleEdit();
    };

    const handleDelete = async (ok: boolean) => {
        setShowDialogDelete(false);

        if (!ok) {
            return;
        }

        setLoadingDelete(true);

        const response = await deleteDevice(props.device.id);

        setLoadingDelete(false);

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(translate("You must be elevated to delete Webauthn credentials"));
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to delete Webauthn credentials"),
                );
            } else {
                createErrorNotification(translate("There was a problem deleting the Webauthn credential"));
            }

            return;
        }

        createSuccessNotification(translate("Successfully deleted the Webauthn credential"));

        props.handleEdit();
    };

    return (
        <Fragment>
            <Paper variant="outlined">
                <Box sx={{ p: 3 }}>
                    <WebauthnDeviceDetailsDialog
                        device={props.device}
                        open={showDialogDetails}
                        handleClose={() => {
                            setShowDialogDetails(false);
                        }}
                    />
                    <WebauthnDeviceEditDialog device={props.device} open={showDialogEdit} handleClose={handleEdit} />
                    <WebauthnDeviceDeleteDialog
                        device={props.device}
                        open={showDialogDelete}
                        handleClose={handleDelete}
                    />
                    <Stack direction="row" spacing={1} alignItems="center">
                        <Fingerprint fontSize="large" color={"warning"} />
                        <Stack spacing={0} sx={{ minWidth: 400 }}>
                            <Box>
                                <Typography display="inline" sx={{ fontWeight: "bold" }}>
                                    {props.device.description}
                                </Typography>
                                <Typography
                                    display="inline"
                                    variant="body2"
                                >{` (${props.device.attestation_type.toUpperCase()})`}</Typography>
                            </Box>
                            <Typography variant={"caption"}>
                                {translate("Added", {
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
                            <Typography variant={"caption"}>
                                {props.device.last_used_at === undefined
                                    ? translate("Never used")
                                    : translate("Last Used", {
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
                        </Stack>

                        <Tooltip title={translate("Display extended information for this Webauthn credential")}>
                            <Button
                                variant="outlined"
                                color="primary"
                                startIcon={<InfoOutlinedIcon />}
                                onClick={() => setShowDialogDetails(true)}
                            >
                                {translate("Info")}
                            </Button>
                        </Tooltip>
                        <Tooltip title={translate("Edit information for this Webauthn credential")}>
                            <Button
                                variant="outlined"
                                color="primary"
                                startIcon={loadingEdit ? <CircularProgress color="inherit" size={20} /> : <EditIcon />}
                                onClick={loadingEdit ? undefined : () => setShowDialogEdit(true)}
                            >
                                {translate("Edit")}
                            </Button>
                        </Tooltip>
                        <Tooltip title={translate("Remove this Webauthn credential")}>
                            <Button
                                variant="outlined"
                                color="primary"
                                startIcon={
                                    loadingDelete ? <CircularProgress color="inherit" size={20} /> : <DeleteIcon />
                                }
                                onClick={loadingDelete ? undefined : () => setShowDialogDelete(true)}
                            >
                                {translate("Remove")}
                            </Button>
                        </Tooltip>
                    </Stack>
                </Box>
            </Paper>
        </Fragment>
    );
}
