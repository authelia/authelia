import React, { Fragment } from "react";

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    Typography,
} from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2";
import { useTranslation } from "react-i18next";

import CopyButton from "@components/CopyButton";
import { WebAuthnDevice, toTransportName } from "@models/WebAuthn";

interface Props {
    open: boolean;
    device: WebAuthnDevice;
    handleClose: () => void;
}

const WebAuthnDeviceDetailsDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog open={props.open} onClose={props.handleClose} aria-labelledby="webauthn-device-details-dialog-title">
            <DialogTitle id="webauthn-device-details-dialog-title">
                {translate("WebAuthn Credential Details")}
            </DialogTitle>
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
                <CopyButton
                    variant={"contained"}
                    tooltip={`${translate("Click to copy the")} ${translate("KID")}`}
                    value={props.device.kid.toString()}
                    fullWidth={false}
                    childrenCopied={translate("Copied")}
                >
                    {translate("KID")}
                </CopyButton>
                <CopyButton
                    variant={"contained"}
                    tooltip={`${translate("Click to copy the")} ${translate("Public Key")}`}
                    value={props.device.public_key.toString()}
                    fullWidth={false}
                    childrenCopied={translate("Copied")}
                >
                    {translate("Public Key")}
                </CopyButton>
                <Button onClick={props.handleClose}>{translate("Close")}</Button>
            </DialogActions>
        </Dialog>
    );
};

interface PropertyTextProps {
    name: string;
    value: string;
    xs?: number;
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

export default WebAuthnDeviceDetailsDialog;
