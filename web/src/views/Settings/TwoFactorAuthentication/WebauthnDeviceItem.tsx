import React, { Fragment, useState } from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import KeyRoundedIcon from "@mui/icons-material/KeyRounded";
import { Box, Button, CircularProgress, Stack, Typography } from "@mui/material";
import { ButtonProps } from "@mui/material/Button";
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
    handleDeviceEdit(index: number, device: WebauthnDevice): void;
    handleDeviceDelete(device: WebauthnDevice): void;
}

export default function WebauthnDeviceItem(props: Props) {
    const { t: translate } = useTranslation("settings");

    const { createErrorNotification } = useNotifications();

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

        const status = await updateDevice(props.device.id, name);

        setLoadingEdit(false);

        if (status !== 200) {
            createErrorNotification(translate("There was a problem updating the device"));
            return;
        }

        props.handleDeviceEdit(props.index, { ...props.device, description: name });
    };

    const handleDelete = async (ok: boolean) => {
        setShowDialogDelete(false);

        if (!ok) {
            return;
        }

        setLoadingDelete(true);

        const status = await deleteDevice(props.device.id);

        setLoadingDelete(false);

        if (status !== 200) {
            createErrorNotification(translate("There was a problem deleting the device"));
            return;
        }

        props.handleDeviceDelete(props.device);
    };

    return (
        <Fragment>
            <WebauthnDeviceDetailsDialog
                device={props.device}
                open={showDialogDetails}
                handleClose={() => {
                    setShowDialogDetails(false);
                }}
            />
            <WebauthnDeviceEditDialog device={props.device} open={showDialogEdit} handleClose={handleEdit} />
            <WebauthnDeviceDeleteDialog device={props.device} open={showDialogDelete} handleClose={handleDelete} />
            <Stack direction="row" spacing={1} alignItems="center">
                <KeyRoundedIcon fontSize="large" />
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
                    <Typography>Added {props.device.created_at.toString()}</Typography>
                    <Typography>
                        {props.device.last_used_at === undefined
                            ? translate("Never used")
                            : "Last used " + props.device.last_used_at.toString()}
                    </Typography>
                </Stack>
                <Button
                    variant="outlined"
                    color="primary"
                    startIcon={<InfoOutlinedIcon />}
                    onClick={() => setShowDialogDetails(true)}
                >
                    {translate("Info")}
                </Button>
                <LoadingButton
                    loading={loadingEdit}
                    variant="outlined"
                    color="primary"
                    startIcon={<EditIcon />}
                    onClick={() => setShowDialogEdit(true)}
                >
                    {translate("Edit")}
                </LoadingButton>
                <LoadingButton
                    loading={loadingDelete}
                    variant="outlined"
                    color="secondary"
                    startIcon={<DeleteIcon />}
                    onClick={() => setShowDialogDelete(true)}
                >
                    {translate("Remove")}
                </LoadingButton>
            </Stack>
        </Fragment>
    );
}

interface LoadingButtonProps extends ButtonProps {
    loading: boolean;
}

function LoadingButton(props: LoadingButtonProps) {
    let { loading, ...childProps } = props;
    if (loading) {
        childProps = {
            ...childProps,
            startIcon: <CircularProgress color="inherit" size={20} />,
            color: "inherit",
            onClick: undefined,
        };
    }
    return <Button {...childProps}></Button>;
}
