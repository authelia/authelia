import React, { useState } from "react";

import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Stack,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { WebauthnDevice } from "@models/Webauthn";

interface Props {
    open: boolean;
    device: WebauthnDevice | undefined;
    handleClose: () => void;
}

export default function WebauthnDetailsDeleteDialog(props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog open={props.open} onClose={props.handleClose}>
            <DialogTitle>Security key details</DialogTitle>
            <DialogContent>
                <DialogContentText sx={{ mb: 3 }}>{`Extended information for security key "${
                    props.device ? props.device.description : "(unknown)"
                }"`}</DialogContentText>
                {props.device && (
                    <Stack spacing={0} sx={{ minWidth: 400 }}>
                        <PropertyText
                            name={translate("Credential Identifier")}
                            value={props.device.kid.toString()}
                            clipboard={true}
                        />
                        <PropertyText
                            name={translate("Public Key")}
                            value={props.device.public_key.toString()}
                            clipboard={true}
                        />
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
                )}
            </DialogContent>
            <DialogActions>
                <Button onClick={props.handleClose}>Close</Button>
            </DialogActions>
        </Dialog>
    );
}

interface PropertyTextProps {
    name: string;
    value: string;
    clipboard?: boolean;
}

function PropertyText(props: PropertyTextProps) {
    const [copied, setCopied] = useState(false);

    const handleCopyToClipboard = () => {
        navigator.clipboard.writeText(props.value);
        setCopied(true);
        setTimeout(() => {
            setCopied(false);
        }, 3000);
    };

    return (
        <Box onClick={props.clipboard ? handleCopyToClipboard : undefined}>
            <Typography display="inline" sx={{ fontWeight: "bold" }}>
                {`${props.name}: `}
            </Typography>
            <Typography display="inline">
                {props.clipboard ? (copied ? "(copied to clipboard)" : "(click to copy)") : props.value}
            </Typography>
        </Box>
    );
}
