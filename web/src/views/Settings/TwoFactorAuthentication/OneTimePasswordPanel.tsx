import React, { Fragment, useCallback, useState } from "react";

import { Box, Button, CircularProgress, Paper, Tooltip, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import { UserInfoTOTPConfiguration } from "@models/TOTPConfiguration";
import { UserInfo } from "@models/UserInfo";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import OneTimePasswordConfiguration from "@views/Settings/TwoFactorAuthentication/OneTimePasswordConfiguration";
import OneTimePasswordDeleteDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordDeleteDialog";
import OneTimePasswordInformationDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordInformationDialog";
import OneTimePasswordRegisterDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordRegisterDialog";

interface Props {
    info?: UserInfo;
    config: UserInfoTOTPConfiguration | undefined | null;
    handleRefreshState: () => void;
}

const OneTimePasswordPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [elevation, setElevation] = useState<UserSessionElevation>();

    const [dialogInformationOpen, setDialogInformationOpen] = useState(false);

    const [dialogSFOpening, setDialogSFOpening] = useState(false);
    const [dialogIVOpening, setDialogIVOpening] = useState(false);

    const [dialogRegisterOpen, setDialogRegisterOpen] = useState(false);
    const [dialogRegisterOpening, setDialogRegisterOpening] = useState(false);

    const [dialogDeleteOpen, setDialogDeleteOpen] = useState(false);
    const [dialogDeleteOpening, setDialogDeleteOpening] = useState(false);

    const handleResetStateOpening = () => {
        setDialogSFOpening(false);
        setDialogIVOpening(false);
        setDialogRegisterOpening(false);
        setDialogDeleteOpening(false);
    };

    const handleResetState = useCallback(() => {
        handleResetStateOpening();

        setElevation(undefined);

        setDialogRegisterOpen(false);
        setDialogDeleteOpen(false);
    }, []);

    const handleOpenDialogRegister = useCallback(() => {
        handleResetStateOpening();
        setDialogRegisterOpen(true);
    }, []);

    const handleOpenDialogDelete = useCallback(() => {
        handleResetStateOpening();
        setDialogDeleteOpen(true);
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
                .then(() => {
                    setDialogIVOpening(true);
                });
        } else {
            setDialogIVOpening(true);
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
            }
        },
        [
            dialogDeleteOpening,
            dialogRegisterOpening,
            handleOpenDialogDelete,
            handleOpenDialogRegister,
            handleResetState,
        ],
    );

    const handleIVDialogOpened = () => {
        setDialogIVOpening(false);
    };

    const handleElevationRefresh = async () => {
        const result = await getUserSessionElevation();

        setElevation(result);
    };

    const handleElevation = () => {
        handleElevationRefresh().catch(console.error);

        setDialogSFOpening(true);
    };

    const handleInformation = () => {
        setDialogInformationOpen(true);
    };

    const handleRegister = () => {
        setDialogRegisterOpening(true);

        handleElevation();
    };

    const handleDelete = () => {
        if (!props.config) return;

        setDialogDeleteOpening(true);

        handleElevation();
    };

    const registered = props.config !== null && props.config !== undefined;

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
                opening={dialogIVOpening}
                elevation={elevation}
                handleClosed={handleIVDialogClosed}
                handleOpened={handleIVDialogOpened}
            />
            <OneTimePasswordRegisterDialog
                open={dialogRegisterOpen}
                setClosed={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <OneTimePasswordInformationDialog
                open={dialogInformationOpen}
                handleClose={() => {
                    setDialogInformationOpen(false);
                }}
                config={props.config}
            />
            <OneTimePasswordDeleteDialog
                open={dialogDeleteOpen}
                handleClose={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <Paper variant={"outlined"}>
                <Grid container spacing={2} padding={2}>
                    <Grid size={{ xs: 12 }}>
                        <Typography variant={"h5"}>{translate("One-Time Password")}</Typography>
                    </Grid>
                    <Grid size={{ xs: 12 }}>
                        <Tooltip
                            title={
                                !registered
                                    ? translate("Click to add a {{item}} to your account", {
                                          item: translate("One-Time Password"),
                                      })
                                    : translate("You can only register a single One-Time Password")
                            }
                        >
                            <Box component={"span"}>
                                <Button
                                    id={"one-time-password-add"}
                                    variant="outlined"
                                    color="primary"
                                    onClick={handleRegister}
                                    disabled={registered || dialogRegisterOpening || dialogRegisterOpen}
                                    endIcon={
                                        dialogRegisterOpening ? <CircularProgress color="inherit" size={20} /> : null
                                    }
                                    data-1p-ignore
                                >
                                    {translate("Add")}
                                </Button>
                            </Box>
                        </Tooltip>
                    </Grid>
                    {props.config === null || props.config === undefined ? (
                        <Grid size={{ xs: 12 }}>
                            <Typography variant={"subtitle2"}>
                                {translate(
                                    "The One-Time Password has not been registered if you'd like to register it click add",
                                )}
                            </Typography>
                        </Grid>
                    ) : (
                        <Grid size={{ xs: 12, md: 6, xl: 3 }}>
                            <OneTimePasswordConfiguration
                                config={props.config}
                                handleInformation={handleInformation}
                                handleRefresh={props.handleRefreshState}
                                handleDelete={handleDelete}
                            />
                        </Grid>
                    )}
                </Grid>
            </Paper>
        </Fragment>
    );
};

export default OneTimePasswordPanel;
