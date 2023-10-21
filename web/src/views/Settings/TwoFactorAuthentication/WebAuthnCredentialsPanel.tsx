import React, { Fragment, useCallback, useState } from "react";

import { Button, Paper, Tooltip, Typography } from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2/Grid2";
import { useTranslation } from "react-i18next";

import { WebAuthnCredential } from "@models/WebAuthn";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import WebAuthnCredentialDeleteDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDeleteDialog";
import WebAuthnCredentialEditDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialEditDialog";
import WebAuthnCredentialInformationDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialInformationDialog";
import WebAuthnCredentialRegisterDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialRegisterDialog";
import WebAuthnCredentialsGrid from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsGrid";

interface Props {
    credentials: WebAuthnCredential[] | undefined;
    handleRefreshState: () => void;
}

const WebAuthnCredentialsPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [dialogIdentityVerificationPendingOpen, setDialogIdentityVerificationPendingOpen] = useState(false);
    const [dialogRegistrationOpen, setDialogRegistrationOpen] = useState(false);
    const [dialogRegistrationPendingOpen, setDialogRegistrationPendingOpen] = useState(false);
    const [dialogInformationOpen, setDialogInformationOpen] = useState(false);
    const [indexInformation, setIndexInformation] = useState(-1);
    const [dialogEditOpen, setDialogEditOpen] = useState(false);
    const [dialogEditPendingOpen, setDialogEditPendingOpen] = useState(false);
    const [indexEdit, setIndexEdit] = useState(-1);
    const [dialogDeleteOpen, setDialogDeleteOpen] = useState(false);
    const [dialogDeletePendingOpen, setDialogDeletePendingOpen] = useState(false);
    const [indexDelete, setIndexDelete] = useState(-1);

    const handleResetState = () => {
        setDialogIdentityVerificationPendingOpen(false);
        setDialogRegistrationOpen(false);
        setDialogRegistrationPendingOpen(false);
        setDialogEditOpen(false);
        setDialogEditPendingOpen(false);
        setIndexEdit(-1);
        setDialogDeleteOpen(false);
        setDialogDeletePendingOpen(false);
        setIndexDelete(-1);
    };

    const handleOpenDialogRegistration = () => {
        setDialogRegistrationPendingOpen(false);
        setDialogRegistrationOpen(true);
    };

    const handleOpenDialogDelete = () => {
        setDialogDeletePendingOpen(false);
        setDialogDeleteOpen(true);
    };

    const handleOpenDialogEdit = () => {
        setDialogEditPendingOpen(false);
        setDialogEditOpen(true);
    };

    const handleIVDClosed = useCallback(
        (ok: boolean) => {
            if (!ok) {
                console.warn(
                    "Identity Verification dialog close callback was not ok which should probably mean it was cancelled by the user.",
                );

                handleResetState();

                return;
            }

            if (dialogRegistrationPendingOpen) {
                handleOpenDialogRegistration();
            } else if (dialogDeletePendingOpen) {
                handleOpenDialogDelete();
            } else if (dialogEditPendingOpen) {
                handleOpenDialogEdit();
            }
        },
        [dialogDeletePendingOpen, dialogEditPendingOpen, dialogRegistrationPendingOpen],
    );

    const handleInformation = (index: number) => {
        if (!props.credentials) {
            return;
        }

        if (props.credentials.length + 1 < index) {
            return;
        }

        setIndexInformation(index);
        setDialogInformationOpen(true);
    };

    const handleEdit = (index: number) => {
        if (!props.credentials) {
            return;
        }

        if (props.credentials.length + 1 < index) {
            return;
        }

        setDialogEditPendingOpen(true);
        setIndexEdit(index);
        setDialogIdentityVerificationPendingOpen(true);
    };

    const handleDelete = (index: number) => {
        if (!props.credentials) {
            return;
        }

        if (props.credentials.length + 1 < index) {
            return;
        }

        setDialogDeletePendingOpen(true);
        setIndexDelete(index);
        setDialogIdentityVerificationPendingOpen(true);
    };

    return (
        <Fragment>
            <IdentityVerificationDialog
                opening={dialogIdentityVerificationPendingOpen}
                handleClosed={handleIVDClosed}
                handleOpened={() => setDialogIdentityVerificationPendingOpen(false)}
            />
            <WebAuthnCredentialRegisterDialog
                open={dialogRegistrationOpen}
                setClosed={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <WebAuthnCredentialInformationDialog
                credential={
                    indexInformation === -1 || !props.credentials ? undefined : props.credentials[indexInformation]
                }
                open={dialogInformationOpen}
                handleClose={() => {
                    setDialogInformationOpen(false);
                }}
            />
            <WebAuthnCredentialEditDialog
                credential={indexEdit === -1 || !props.credentials ? undefined : props.credentials[indexEdit]}
                open={dialogEditOpen}
                handleClose={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <WebAuthnCredentialDeleteDialog
                open={dialogDeleteOpen}
                credential={indexDelete === -1 || !props.credentials ? undefined : props.credentials[indexDelete]}
                handleClose={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <Paper variant={"outlined"}>
                <Grid container spacing={2} padding={2}>
                    <Grid xs={12}>
                        <Typography variant="h5">{translate("WebAuthn Credentials")}</Typography>
                    </Grid>
                    <Grid xs={4} md={2}>
                        <Tooltip title={translate("Click to add a WebAuthn credential to your account")}>
                            <Button
                                variant="outlined"
                                color="primary"
                                onClick={() => {
                                    setDialogRegistrationPendingOpen(true);
                                    setDialogIdentityVerificationPendingOpen(true);
                                }}
                            >
                                {translate("Add")}
                            </Button>
                        </Tooltip>
                    </Grid>
                    <Grid xs={12}>
                        {props.credentials === undefined || props.credentials.length === 0 ? (
                            <Typography variant={"subtitle2"}>
                                {translate(
                                    "No WebAuthn Credentials have been registered. If you'd like to register one click add.",
                                )}
                            </Typography>
                        ) : (
                            <WebAuthnCredentialsGrid
                                credentials={props.credentials}
                                handleRefreshState={props.handleRefreshState}
                                handleInformation={handleInformation}
                                handleEdit={handleEdit}
                                handleDelete={handleDelete}
                            />
                        )}
                    </Grid>
                </Grid>
            </Paper>
        </Fragment>
    );
};

export default WebAuthnCredentialsPanel;
