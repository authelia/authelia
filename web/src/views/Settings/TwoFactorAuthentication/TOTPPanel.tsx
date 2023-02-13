import React, { Fragment, useState } from "react";

import { Button, Grid, Paper, Stack, Tooltip, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import { UserInfoTOTPConfiguration } from "@models/TOTPConfiguration";
import TOTPDevice from "@views/Settings/TwoFactorAuthentication/TOTPDevice";
import TOTPRegisterDialogController from "@views/Settings/TwoFactorAuthentication/TOTPRegisterDialogController";

interface Props {
    config: UserInfoTOTPConfiguration | undefined | null;
    handleRefreshState: () => void;
}

export default function TOTPPanel(props: Props) {
    const { t: translate } = useTranslation("settings");

    const [showRegisterDialog, setShowRegisterDialog] = useState<boolean>(false);

    return (
        <Fragment>
            <TOTPRegisterDialogController
                open={showRegisterDialog}
                setClosed={() => {
                    setShowRegisterDialog(false);
                    props.handleRefreshState();
                }}
            />
            <Paper variant="outlined" sx={{ p: 3 }}>
                <Grid container spacing={1}>
                    <Grid item xs={12}>
                        <Typography variant="h5">{translate("One Time Password")}</Typography>
                    </Grid>
                    {props.config === undefined || props.config === null ? (
                        <Fragment>
                            <Grid item xs={12}>
                                <Tooltip
                                    title={translate("Click to add a Time-based One Time Password to your account")}
                                >
                                    <Button
                                        variant="outlined"
                                        color="primary"
                                        onClick={() => {
                                            setShowRegisterDialog(true);
                                        }}
                                    >
                                        {translate("Add")}
                                    </Button>
                                </Tooltip>
                            </Grid>
                            <Grid item xs={12}>
                                <Typography variant={"subtitle2"}>
                                    {translate(
                                        "The One Time Password has not been registered. If you'd like to register it click add.",
                                    )}
                                </Typography>
                            </Grid>
                        </Fragment>
                    ) : (
                        <Grid item xs={12}>
                            <Stack spacing={2}>
                                <TOTPDevice index={0} config={props.config} handleRefresh={props.handleRefreshState} />
                            </Stack>
                        </Grid>
                    )}
                </Grid>
            </Paper>
        </Fragment>
    );
}
