import React, { Fragment } from "react";

import {
    Alert,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    Typography,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import CopyButton from "@components/CopyButton";
import { FormatDateHumanReadable } from "@i18n/formats";
import { WebAuthnCredential, toAttachmentName, toTransportName } from "@models/WebAuthn";

interface Props {
    open: boolean;
    credential?: WebAuthnCredential;
    handleClose: () => void;
}

const WebAuthnCredentialInformationDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog open={props.open} onClose={props.handleClose} aria-labelledby="webauthn-credential-info-dialog-title">
            <DialogTitle id="webauthn-credential-info-dialog-title">
                {translate("WebAuthn Credential Information")}
            </DialogTitle>
            <DialogContent>
                {!props.credential ? (
                    <DialogContentText sx={{ mb: 3 }}>
                        {translate("The WebAuthn Credential information is not loaded")}
                    </DialogContentText>
                ) : (
                    <Fragment>
                        <DialogContentText sx={{ mb: 3 }}>
                            {translate("Extended information for WebAuthn Credential", {
                                description: props.credential.description,
                            })}
                        </DialogContentText>
                        {props.credential.legacy ? (
                            <DialogContentText sx={{ mb: 3 }}>
                                <Alert severity={"warning"}>
                                    {translate(
                                        "This is a legacy WebAuthn Credential if it's not operating normally you may need to delete it and register it again",
                                    )}
                                </Alert>
                            </DialogContentText>
                        ) : null}
                        <Grid container spacing={2}>
                            <Grid size={{ md: 3 }} sx={{ display: { md: "block", xs: "none" } }}>
                                <Fragment />
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                                <Divider />
                            </Grid>
                            <PropertyText name={translate("Description")} value={props.credential.description} />
                            <PropertyText name={translate("Relying Party ID")} value={props.credential.rpid} />
                            <PropertyText
                                name={translate("Authenticator GUID")}
                                value={
                                    props.credential.aaguid === undefined
                                        ? translate("Unknown")
                                        : props.credential.aaguid
                                }
                            />
                            <PropertyText
                                name={translate("Attestation Type")}
                                value={props.credential.attestation_type}
                            />
                            <PropertyText
                                name={translate("Attachment")}
                                value={translate(toAttachmentName(props.credential.attachment))}
                            />
                            <PropertyText
                                name={translate("Discoverable")}
                                value={props.credential.discoverable ? translate("Yes") : translate("No")}
                            />
                            <PropertyText
                                name={translate("User Verified")}
                                value={props.credential.verified ? translate("Yes") : translate("No")}
                            />
                            <PropertyText
                                name={translate("Backup State")}
                                value={
                                    props.credential.backup_eligible
                                        ? props.credential.backup_state
                                            ? translate("Backed Up")
                                            : translate("Eligible")
                                        : translate("Not Eligible")
                                }
                            />
                            <PropertyText
                                name={translate("Transports")}
                                value={
                                    props.credential.transports === null || props.credential.transports.length === 0
                                        ? translate("Unknown")
                                        : props.credential.transports
                                              .map((transport) => toTransportName(transport))
                                              .join(", ")
                                }
                            />
                            <PropertyText
                                name={translate("Clone Warning")}
                                value={props.credential.clone_warning ? translate("Yes") : translate("No")}
                            />
                            <PropertyText name={translate("Usage Count")} value={`${props.credential.sign_count}`} />
                            <PropertyText
                                name={translate("Added")}
                                value={translate("{{when, datetime}}", {
                                    formatParams: { when: FormatDateHumanReadable },
                                    when: new Date(props.credential.created_at),
                                })}
                            />
                            <PropertyText
                                name={translate("Last Used")}
                                value={
                                    props.credential.last_used_at
                                        ? translate("{{when, datetime}}", {
                                              formatParams: { when: FormatDateHumanReadable },
                                              when: new Date(props.credential.last_used_at),
                                          })
                                        : translate("Never")
                                }
                            />
                        </Grid>
                    </Fragment>
                )}
            </DialogContent>
            <DialogActions>
                {props.credential ? (
                    <Fragment>
                        {" "}
                        <CopyButton
                            variant={"contained"}
                            tooltip={translate("Click to copy the {{value}}", { value: "KID" })}
                            value={props.credential.kid.toString()}
                            fullWidth={false}
                            childrenCopied={translate("Copied")}
                        >
                            {translate("KID")}
                        </CopyButton>
                        <CopyButton
                            variant={"contained"}
                            tooltip={translate("Click to copy the {{value}}", { value: translate("Public Key") })}
                            value={props.credential.public_key.toString()}
                            fullWidth={false}
                            childrenCopied={translate("Copied")}
                        >
                            {translate("Public Key")}
                        </CopyButton>
                    </Fragment>
                ) : undefined}
                <Button id={"dialog-close"} onClick={props.handleClose}>
                    {translate("Close")}
                </Button>
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
        <Grid size={{ xs: props.xs !== undefined ? props.xs : 12 }}>
            <Typography display="inline" sx={{ fontWeight: "bold" }}>
                {`${props.name}: `}
            </Typography>
            <Typography display="inline">{props.value}</Typography>
        </Grid>
    );
}

export default WebAuthnCredentialInformationDialog;
