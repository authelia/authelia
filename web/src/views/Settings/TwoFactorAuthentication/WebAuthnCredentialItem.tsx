import React, { useState } from "react";

import { Fingerprint } from "@mui/icons-material";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import { Paper, Stack, Tooltip, Typography } from "@mui/material";
import IconButton from "@mui/material/IconButton";
import Grid from "@mui/material/Unstable_Grid2";
import { useTranslation } from "react-i18next";

import { FormatDateHumanReadable } from "@i18n/formats";
import { WebAuthnCredential } from "@models/WebAuthn";
import WebAuthnCredentialDetailsDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDetailsDialog";

interface Props {
    index: number;
    credential: WebAuthnCredential;
    handleDelete: (index: number) => void;
    handleEdit: (index: number) => void;
}

const WebAuthnCredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [showDialogDetails, setShowDialogDetails] = useState<boolean>(false);

    const handleShowEditDialog = () => {
        // loadingEdit ? undefined : () => setShowDialogEdit(true)
        props.handleEdit(props.index);
    };

    const handleShowDeleteDialog = () => {
        // loadingDelete ? undefined : () => setShowDialogDelete(true);
        props.handleDelete(props.index);
    };

    return (
        <Grid xs={12} md={6} xl={3}>
            <WebAuthnCredentialDetailsDialog
                credential={props.credential}
                open={showDialogDetails}
                handleClose={() => {
                    setShowDialogDetails(false);
                }}
            />
            <Paper variant="outlined">
                <Grid container spacing={1} alignItems="center" padding={3}>
                    <Grid xs={12} sm={6} md={6}>
                        <Grid container>
                            <Grid xs={12}>
                                <Stack direction={"row"} spacing={1} alignItems={"center"}>
                                    <Fingerprint fontSize="large" color={"warning"} />
                                    <Typography display="inline" sx={{ fontWeight: "bold" }}>
                                        {props.credential.description}
                                    </Typography>
                                    <Typography
                                        display="inline"
                                        variant="body2"
                                    >{` (${props.credential.attestation_type.toUpperCase()})`}</Typography>
                                </Stack>
                            </Grid>
                            <Grid xs={12} sx={{ display: { xs: "none", md: "block" } }}>
                                <Stack direction={"row"} spacing={1} alignItems={"center"}>
                                    <Typography variant={"caption"} sx={{ display: { xs: "none", md: "block" } }}>
                                        {translate("Added when", {
                                            when: new Date(props.credential.created_at),
                                            formatParams: { when: FormatDateHumanReadable },
                                        })}
                                    </Typography>
                                    <Typography variant={"caption"} sx={{ display: { xs: "none", md: "block" } }}>
                                        {props.credential.last_used_at === undefined
                                            ? translate("Never used")
                                            : translate("Last Used when", {
                                                  when: new Date(props.credential.last_used_at),
                                                  formatParams: { when: FormatDateHumanReadable },
                                              })}
                                    </Typography>
                                </Stack>
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
                                <IconButton color="primary" onClick={handleShowEditDialog}>
                                    <EditIcon />
                                </IconButton>
                            </Tooltip>
                            <Tooltip title={translate("Remove this WebAuthn credential")}>
                                <IconButton color="primary" onClick={handleShowDeleteDialog}>
                                    <DeleteIcon />
                                </IconButton>
                            </Tooltip>
                        </Stack>
                    </Grid>
                </Grid>
            </Paper>
        </Grid>
    );
};

export default WebAuthnCredentialItem;
