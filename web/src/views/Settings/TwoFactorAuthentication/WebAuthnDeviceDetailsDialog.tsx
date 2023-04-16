import React, { Fragment, useState } from "react";

import { Check, ContentCopy } from "@mui/icons-material";
import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    Tooltip,
    Typography,
} from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2";
import { useTranslation } from "react-i18next";

import { WebAuthnDevice, toTransportName } from "@models/WebAuthn";

interface Props {
    open: boolean;
    device: WebAuthnDevice;
    handleClose: () => void;
}

export default function WebAuthnDeviceDetailsDialog(props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog open={props.open} onClose={props.handleClose}>
            <DialogTitle>{translate("WebAuthn Credential Details")}</DialogTitle>
            <DialogContent>
                <DialogContentText sx={{ mb: 3 }}>
                    {translate("Extended WebAuthn credential information for security key", {
                        description: props.device.description,
                    })}
                </DialogContentText>
                <Grid container spacing={2}>
                    <Grid md={3} sx={{ display: { xs: "none", md: "block" } }}>
                        <Fragment />
                    </Grid>
                    <Grid xs={4} md={2}>
                        <PropertyCopyButton name={translate("KID")} value={props.device.kid.toString()} />
                    </Grid>
                    <Grid xs={8} md={4}>
                        <PropertyCopyButton name={translate("Public Key")} value={props.device.public_key.toString()} />
                    </Grid>
                    <Grid xs={12}>
                        <Divider />
                    </Grid>
                    <PropertyText name={translate("Description")} value={props.device.description} />
                    <PropertyText name={translate("Relying Party ID")} value={props.device.rpid} />
                    <PropertyText
                        name={translate("Authenticator GUID")}
                        value={props.device.aaguid === undefined ? translate("Unknown") : props.device.aaguid}
                    />
                    <PropertyText name={translate("Attestation Type")} value={props.device.attestation_type} />
                    <PropertyText
                        name={translate("Attachment")}
                        value={props.device.attachment === "" ? translate("Unknown") : props.device.attachment}
                    />
                    <PropertyText
                        name={translate("Discoverable")}
                        value={props.device.discoverable ? translate("Yes") : translate("No")}
                    />
                    <PropertyText
                        name={translate("User Verified")}
                        value={props.device.verified ? translate("Yes") : translate("No")}
                    />
                    <PropertyText
                        name={translate("Backup State")}
                        value={
                            props.device.backup_eligible
                                ? props.device.backup_state
                                    ? translate("Backed Up")
                                    : translate("Eligible")
                                : translate("Not Eligible")
                        }
                    />
                    <PropertyText
                        name={translate("Transports")}
                        value={
                            props.device.transports === null || props.device.transports.length === 0
                                ? translate("Unknown")
                                : props.device.transports.map((transport) => toTransportName(transport)).join(", ")
                        }
                    />
                    <PropertyText
                        name={translate("Clone Warning")}
                        value={props.device.clone_warning ? translate("Yes") : translate("No")}
                    />
                    <PropertyText name={translate("Usage Count")} value={`${props.device.sign_count}`} />
                    <PropertyText
                        name={translate("Added")}
                        value={translate("{{when, datetime}}", {
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
                    />
                    <PropertyText
                        name={translate("Last Used")}
                        value={
                            props.device.last_used_at
                                ? translate("{{when, datetime}}", {
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
                                  })
                                : translate("Never")
                        }
                    />
                </Grid>
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
    xs?: number;
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
        <Grid xs={props.xs !== undefined ? props.xs : 12}>
            <Typography display="inline" sx={{ fontWeight: "bold" }}>
                {`${props.name}: `}
            </Typography>
            <Typography display="inline">{props.value}</Typography>
        </Grid>
    );
}
