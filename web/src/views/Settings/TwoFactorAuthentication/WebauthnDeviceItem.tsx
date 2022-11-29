import React from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import KeyRoundedIcon from "@mui/icons-material/KeyRounded";
import ToggleOffOutlinedIcon from "@mui/icons-material/ToggleOffOutlined";
import ToggleOnOutlinedIcon from "@mui/icons-material/ToggleOnOutlined";
import { Box, Button, CircularProgress, Stack, Typography } from "@mui/material";
import { ButtonProps } from "@mui/material/Button";
import { useTranslation } from "react-i18next";

import { WebauthnDevice } from "@models/Webauthn";

interface Props {
    device: WebauthnDevice;
    deleting: boolean;
    editing: boolean;
    enabling: boolean;
    enabled: boolean;
    webauthnShowDetails: boolean;
    handleWebAuthnDetailsChange: () => void;
    handleDetails: () => void;
    handleDelete: () => void;
    handleEdit: () => void;
    handleEnable: () => void;
}

export default function WebauthnDeviceItem(props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
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
            <ToggleButton
                enabled={props.enabled}
                loading={props.enabling}
                variant="outlined"
                color="primary"
                enabledIcon={<ToggleOnOutlinedIcon />}
                disabledIcon={<ToggleOffOutlinedIcon />}
                enabledText={"Disable"}
                disabledText={"Enable"}
                onClick={props.handleEnable}
            />
            <Button variant="outlined" color="primary" startIcon={<InfoOutlinedIcon />} onClick={props.handleDetails}>
                {translate("Info")}
            </Button>
            <LoadingButton
                loading={props.editing}
                variant="outlined"
                color="primary"
                startIcon={<EditIcon />}
                onClick={props.handleEdit}
            >
                {translate("Edit")}
            </LoadingButton>
            <LoadingButton
                loading={props.deleting}
                variant="outlined"
                color="secondary"
                startIcon={<DeleteIcon />}
                onClick={props.handleDelete}
            >
                {translate("Remove")}
            </LoadingButton>
        </Stack>
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

interface ToggleButtonProps extends LoadingButtonProps {
    enabledIcon: React.ReactNode;
    disabledIcon: React.ReactNode;
    enabledText: string;
    disabledText: string;
    enabled: boolean;
}

function ToggleButton(props: ToggleButtonProps) {
    let { enabled, enabledIcon, disabledIcon, enabledText, disabledText, ...childProps } = props;
    return (
        <LoadingButton startIcon={enabled ? enabledIcon : disabledIcon} {...childProps}>
            {enabled ? enabledText : disabledText}
        </LoadingButton>
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
