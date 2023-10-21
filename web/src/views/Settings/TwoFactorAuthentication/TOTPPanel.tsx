import React, { Fragment, useCallback, useState } from "react";

import { Button, Paper, Stack, Tooltip, Typography } from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2/Grid2";
import { useTranslation } from "react-i18next";

import { UserInfoTOTPConfiguration } from "@models/TOTPConfiguration";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import TOTPConfiguration from "@views/Settings/TwoFactorAuthentication/TOTPConfiguration";
import TOTPDeleteDialog from "@views/Settings/TwoFactorAuthentication/TOTPDeleteDialog";
import TOTPRegisterDialog from "@views/Settings/TwoFactorAuthentication/TOTPRegisterDialog";

interface Props {
    config: UserInfoTOTPConfiguration | undefined | null;
    handleRefreshState: () => void;
}

const TOTPPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [dialogIdentityVerificationPendingOpen, setDialogIdentityVerificationPendingOpen] = useState(false);
    const [dialogRegistrationOpen, setDialogRegistrationOpen] = useState(false);
    const [dialogRegistrationPendingOpen, setDialogRegistrationPendingOpen] = useState(false);
    const [dialogDeleteOpen, setDialogDeleteOpen] = useState(false);
    const [dialogDeletePendingOpen, setDialogDeletePendingOpen] = useState(false);

    const handleResetState = () => {
        setDialogIdentityVerificationPendingOpen(false);
        setDialogRegistrationOpen(false);
        setDialogRegistrationPendingOpen(false);
        setDialogDeleteOpen(false);
        setDialogDeletePendingOpen(false);
    };

    const handleOpenDialogRegistration = () => {
        setDialogRegistrationPendingOpen(false);
        setDialogRegistrationOpen(true);
    };

    const handleOpenDialogDelete = () => {
        setDialogDeletePendingOpen(false);
        setDialogDeleteOpen(true);
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
            }
        },
        [dialogDeletePendingOpen, dialogRegistrationPendingOpen],
    );

    const handleDelete = () => {
        if (!props.config) {
            return;
        }

        setDialogDeletePendingOpen(true);
        setDialogIdentityVerificationPendingOpen(true);
    };

    return (
        <Fragment>
            <IdentityVerificationDialog
                opening={dialogIdentityVerificationPendingOpen}
                handleClosed={handleIVDClosed}
                handleOpened={() => setDialogIdentityVerificationPendingOpen(false)}
            />
            <TOTPRegisterDialog
                open={dialogRegistrationOpen}
                setClosed={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <TOTPDeleteDialog
                open={dialogDeleteOpen}
                handleClose={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <Paper variant={"outlined"}>
                <Grid container spacing={2} padding={2}>
                    <Grid xs={12} lg={8}>
                        <Typography variant={"h5"}>{translate("One-Time Password")}</Typography>
                    </Grid>
                    {props.config === undefined || props.config === null ? (
                        <Fragment>
                            <Grid xs={2}>
                                <Tooltip
                                    title={translate("Click to add a {{item}} to your account", {
                                        item: translate("One-Time Password"),
                                    })}
                                >
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
                                <Typography variant={"subtitle2"}>
                                    {translate(
                                        "The One-Time Password has not been registered. If you'd like to register it click add",
                                    )}
                                </Typography>
                            </Grid>
                        </Fragment>
                    ) : (
                        <Grid xs={12}>
                            <Stack spacing={3}>
                                <TOTPConfiguration
                                    config={props.config}
                                    handleRefresh={props.handleRefreshState}
                                    handleDelete={handleDelete}
                                />
                            </Stack>
                        </Grid>
                    )}
                </Grid>
            </Paper>
        </Fragment>
    );
};

export default TOTPPanel;
