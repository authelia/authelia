import React, { useState } from "react";

import { Check, ContentCopy } from "@mui/icons-material";
import {
    Box,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Stack,
    Tooltip,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { WebauthnDevice } from "@models/Webauthn";

interface Props {
    open: boolean;
    device: WebauthnDevice;
    handleClose: () => void;
}

export default function WebauthnDetailsDeleteDialog(props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog open={props.open} onClose={props.handleClose}>
            <DialogTitle>{translate("Webauthn Credential Details")}</DialogTitle>
            <DialogContent>
                <DialogContentText sx={{ mb: 3 }}>
                    {translate("Extended Webauthn credential information for security key", {
                        description: props.device.description,
                    })}
                </DialogContentText>
                <Stack spacing={0} sx={{ minWidth: 400 }}>
                    <Box paddingBottom={2}>
                        <Stack direction="row" spacing={1} alignItems="center">
                            <PropertyCopyButton name={translate("Identifier")} value={props.device.kid.toString()} />
                            <PropertyCopyButton
                                name={translate("Public Key")}
                                value={props.device.public_key.toString()}
                            />
                        </Stack>
                    </Box>
                    <PropertyText name={translate("Description")} value={props.device.description} />
                    <PropertyText name={translate("Relying Party ID")} value={props.device.rpid} />
                    <PropertyText
                        name={translate("Authenticator Attestation GUID")}
                        value={props.device.aaguid === undefined ? "N/A" : props.device.aaguid}
                    />
                    <PropertyText name={translate("Attestation Type")} value={props.device.attestation_type} />
                    <PropertyText
                        name={translate("Transports")}
                        value={props.device.transports.length === 0 ? "N/A" : props.device.transports.join(", ")}
                    />
                    <PropertyText
                        name={translate("Clone Warning")}
                        value={props.device.clone_warning ? translate("Yes") : translate("No")}
                    />
                    <PropertyText name={translate("Usage Count")} value={`${props.device.sign_count}`} />
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={props.handleClose}>{translate("Close")}</Button>
            </DialogActions>
        </Dialog>
    );
}

interface PropertyTextProps {
    name: string;
    value: string;
}

function PropertyCopyButton(props: PropertyTextProps) {
    const { t: translate } = useTranslation("settings");

    const [copied, setCopied] = useState(false);
    const [copying, setCopying] = useState(false);

    const handleCopyToClipboard = () => {
        if (copied) {
            return;
        }

        (async () => {
            setCopying(true);

            await navigator.clipboard.writeText(props.value);

            setTimeout(() => {
                setCopying(false);
                setCopied(true);
            }, 500);

            setTimeout(() => {
                setCopied(false);
            }, 2000);
        })();
    };

    return (
        <Tooltip title={`${translate("Click to copy the")} ${props.name}`}>
            <Button
                variant="outlined"
                color={copied ? "success" : "primary"}
                onClick={copying ? undefined : handleCopyToClipboard}
                startIcon={
                    copying ? <CircularProgress color="inherit" size={20} /> : copied ? <Check /> : <ContentCopy />
                }
            >
                {copied ? translate("Copied") : props.name}
            </Button>
        </Tooltip>
    );
}

function PropertyText(props: PropertyTextProps) {
    return (
        <Box>
            <Typography display="inline" sx={{ fontWeight: "bold" }}>
                {`${props.name}: `}
            </Typography>
            <Typography display="inline">{props.value}</Typography>
        </Box>
    );
}
