import { Fragment, useCallback, useState } from "react";

import { Button, CircularProgress, Paper, Tooltip, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import { UserInfo } from "@models/UserInfo";
import { WebAuthnCredential } from "@models/WebAuthn";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import WebAuthnCredentialDeleteDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDeleteDialog";
import WebAuthnCredentialEditDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialEditDialog";
import WebAuthnCredentialInformationDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialInformationDialog";
import WebAuthnCredentialRegisterDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialRegisterDialog";
import WebAuthnCredentialsGrid from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsGrid";

interface Props {
    info?: UserInfo;
    credentials: undefined | WebAuthnCredential[];
    handleRefreshState: () => void;
}

const WebAuthnCredentialsPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [elevation, setElevation] = useState<UserSessionElevation>();

    const [dialogSFOpening, setDialogSFOpening] = useState(false);
    const [dialogIVOpening, setDialogIVOpening] = useState(false);

    const [dialogRegisterOpen, setDialogRegisterOpen] = useState(false);
    const [dialogRegisterOpening, setDialogRegisterOpening] = useState(false);

    const [dialogInformationOpen, setDialogInformationOpen] = useState(false);
    const [indexInformation, setIndexInformation] = useState(-1);

    const [dialogEditOpen, setDialogEditOpen] = useState(false);
    const [dialogEditOpening, setDialogEditOpening] = useState(false);
    const [indexEdit, setIndexEdit] = useState(-1);

    const [dialogDeleteOpen, setDialogDeleteOpen] = useState(false);
    const [dialogDeleteOpening, setDialogDeleteOpening] = useState(false);
    const [indexDelete, setIndexDelete] = useState(-1);

    const handleResetStateOpening = () => {
        setDialogSFOpening(false);
        setDialogIVOpening(false);
        setDialogRegisterOpening(false);
        setDialogEditOpening(false);
        setDialogDeleteOpening(false);
    };

    const handleResetState = useCallback(() => {
        handleResetStateOpening();

        setElevation(undefined);

        setDialogRegisterOpen(false);
        setDialogEditOpen(false);
        setIndexEdit(-1);
        setDialogDeleteOpen(false);
        setIndexDelete(-1);
    }, []);

    const handleOpenDialogRegister = useCallback(() => {
        handleResetStateOpening();
        setDialogRegisterOpen(true);
    }, []);

    const handleOpenDialogDelete = useCallback(() => {
        handleResetStateOpening();
        setDialogDeleteOpen(true);
    }, []);

    const handleOpenDialogEdit = useCallback(() => {
        handleResetStateOpening();
        setDialogEditOpen(true);
    }, []);

    const handleSFDialogClosed = (ok: boolean, changed: boolean) => {
        if (!ok) {
            console.warn("Second Factor dialog close callback failed, it was likely cancelled by the user.");

            handleResetState();

            return;
        }

        if (changed) {
            handleElevationRefresh()
                .catch(console.error)
                .then((refreshedElevation) => {
                    if (refreshedElevation) {
                        const isElevatedFromRefresh =
                            refreshedElevation.elevated || refreshedElevation.skip_second_factor;
                        if (isElevatedFromRefresh) {
                            setElevation(undefined);
                            if (dialogRegisterOpening) {
                                handleOpenDialogRegister();
                            } else if (dialogDeleteOpening) {
                                handleOpenDialogDelete();
                            } else if (dialogEditOpening) {
                                handleOpenDialogEdit();
                            }
                        } else {
                            setDialogIVOpening(true);
                        }
                    }
                });
        } else {
            const isElevated = elevation && (elevation.elevated || elevation.skip_second_factor);
            if (isElevated) {
                setElevation(undefined);
                if (dialogRegisterOpening) {
                    handleOpenDialogRegister();
                } else if (dialogDeleteOpening) {
                    handleOpenDialogDelete();
                } else if (dialogEditOpening) {
                    handleOpenDialogEdit();
                }
            } else {
                setDialogIVOpening(true);
            }
        }
    };

    const handleSFDialogOpened = () => {
        setDialogSFOpening(false);
    };

    const handleIVDialogClosed = useCallback(
        (ok: boolean) => {
            if (!ok) {
                console.warn(
                    "Identity Verification dialog close callback failed, it was likely cancelled by the user.",
                );

                handleResetState();

                return;
            }

            setElevation(undefined);

            if (dialogRegisterOpening) {
                handleOpenDialogRegister();
            } else if (dialogDeleteOpening) {
                handleOpenDialogDelete();
            } else if (dialogEditOpening) {
                handleOpenDialogEdit();
            }
        },
        [
            handleResetState,
            handleOpenDialogRegister,
            handleOpenDialogDelete,
            handleOpenDialogEdit,
            dialogRegisterOpening,
            dialogDeleteOpening,
            dialogEditOpening,
        ],
    );

    const handleIVDialogOpened = useCallback(() => {
        setDialogIVOpening(false);
    }, []);

    const handleElevationRefresh = async () => {
        const result = await getUserSessionElevation();

        setElevation(result);
        return result;
    };

    const handleElevation = () => {
        handleElevationRefresh().catch(console.error);

        setDialogSFOpening(true);
    };

    const handleRegister = () => {
        setDialogRegisterOpening(true);

        handleElevation();
    };

    const handleInformation = (index: number) => {
        if (!props.credentials) return;

        if (props.credentials.length + 1 < index) return;

        setIndexInformation(index);
        setDialogInformationOpen(true);
    };

    const handleEdit = (index: number) => {
        if (!props.credentials) return;

        if (props.credentials.length + 1 < index) return;

        setDialogEditOpening(true);
        setIndexEdit(index);

        handleElevation();
    };

    const handleDelete = (index: number) => {
        if (!props.credentials) return;

        if (props.credentials.length + 1 < index) return;

        setDialogDeleteOpening(true);
        setIndexDelete(index);

        handleElevation();
    };

    return (
        <Fragment>
            <SecondFactorDialog
                info={props.info}
                elevation={elevation}
                opening={dialogSFOpening}
                handleClosed={handleSFDialogClosed}
                handleOpened={handleSFDialogOpened}
            />
            <IdentityVerificationDialog
                elevation={elevation}
                opening={dialogIVOpening}
                handleClosed={handleIVDialogClosed}
                handleOpened={handleIVDialogOpened}
            />
            <WebAuthnCredentialRegisterDialog
                open={dialogRegisterOpen}
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
                    <Grid size={{ xs: 12 }}>
                        <Typography variant="h5">{translate("WebAuthn Credentials")}</Typography>
                    </Grid>
                    <Grid size={{ md: 2, xs: 4 }}>
                        <Tooltip
                            title={translate("Click to add a {{item}} to your account", {
                                item: translate("WebAuthn Credential"),
                            })}
                        >
                            <Button
                                id={"webauthn-credential-add"}
                                variant="outlined"
                                color="primary"
                                onClick={handleRegister}
                                disabled={dialogRegisterOpening || dialogRegisterOpen}
                                endIcon={dialogRegisterOpening ? <CircularProgress color="inherit" size={20} /> : null}
                            >
                                {translate("Add")}
                            </Button>
                        </Tooltip>
                    </Grid>
                    <Grid size={{ xs: 12 }}>
                        {props.credentials === undefined || props.credentials.length === 0 ? (
                            <Typography variant={"subtitle2"}>
                                {translate(
                                    "No WebAuthn Credentials have been registered if you'd like to register one click add",
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
