import React, { Fragment, useState } from "react";

import { Button, Paper, Stack, Tooltip, Typography } from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2";
import { useTranslation } from "react-i18next";

import { UserInfoTOTPConfiguration } from "@models/TOTPConfiguration";
import TOTPDevice from "@views/Settings/TwoFactorAuthentication/TOTPDevice";
import TOTPRegisterDialogController from "@views/Settings/TwoFactorAuthentication/TOTPRegisterDialogController";

interface Props {
    config: UserInfoTOTPConfiguration | undefined | null;
    handleRefreshState: () => void;
}

const TOTPPanel = function (props: Props) {
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
            <Paper variant={"outlined"}>
                <Grid container spacing={2} padding={2}>
                    <Grid xs={12} lg={8}>
                        <Typography variant={"h5"}>{translate("One Time Password")}</Typography>
                    </Grid>
                    {props.config === undefined || props.config === null ? (
                        <Fragment>
                            <Grid xs={2}>
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
                            <Grid xs={12}>
                                <Typography variant={"subtitle2"}>
                                    {translate(
                                        "The One Time Password has not been registered. If you'd like to register it click add.",
                                    )}
                                </Typography>
                            </Grid>
                        </Fragment>
                    ) : (
                        <Grid xs={12}>
                            <Stack spacing={3}>
                                <TOTPDevice config={props.config} handleRefresh={props.handleRefreshState} />
                            </Stack>
                        </Grid>
                    )}
                </Grid>
            </Paper>
        </Fragment>
    );
};

export default TOTPPanel;
