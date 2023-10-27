import React, { Fragment, useState } from "react";

import { QrCode2 } from "@mui/icons-material";
import DeleteIcon from "@mui/icons-material/Delete";
import { Box, Button, CircularProgress, Paper, Stack, Tooltip, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { UserInfoTOTPConfiguration, toAlgorithmString } from "@models/TOTPConfiguration";
import { deleteUserTOTPConfiguration } from "@services/UserInfoTOTPConfiguration";
import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";

interface Props {
    config: UserInfoTOTPConfiguration;
    handleRefresh: () => void;
}

const TOTPDevice = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [showDialogDelete, setShowDialogDelete] = useState<boolean>(false);

    const [loadingDelete, setLoadingDelete] = useState<boolean>(false);

    const handleDelete = async (ok: boolean) => {
        setShowDialogDelete(false);

        if (!ok) {
            return;
        }

        setLoadingDelete(true);

        const response = await deleteUserTOTPConfiguration();

        setLoadingDelete(false);

        if (response.data.status === "KO") {
            if (response.data.elevation) {
                createErrorNotification(translate("You must be elevated to delete the One-Time Password"));
            } else if (response.data.authentication) {
                createErrorNotification(
                    translate("You must have a higher authentication level to delete the One-Time Password"),
                );
            } else {
                createErrorNotification(translate("There was a problem deleting the One-Time Password"));
            }

            return;
        }

        createSuccessNotification(translate("Successfully deleted the One Time Password configuration"));

        props.handleRefresh();
    };

    return (
        <Fragment>
            <Paper variant="outlined">
                <Box sx={{ p: 3 }}>
                    <DeleteDialog
                        open={showDialogDelete}
                        handleClose={handleDelete}
                        title={translate("Remove One Time Password")}
                        text={translate(
                            "Are you sure you want to remove the Time-based One Time Password from from your account",
                        )}
                    />
                    <Stack direction={"row"} spacing={1} alignItems={"center"}>
                        <QrCode2 fontSize="large" />
                        <Stack spacing={0} sx={{ minWidth: 400 }}>
                            <Box>
                                <Typography display={"inline"} sx={{ fontWeight: "bold" }}>
                                    {props.config.issuer}
                                </Typography>
                                <Typography display={"inline"} variant={"body2"}>
                                    {" (" +
                                        translate("{{algorithm}}, {{digits}} digits, {{seconds}} seconds", {
                                            algorithm: toAlgorithmString(props.config.algorithm),
                                            digits: props.config.digits,
                                            seconds: props.config.period,
                                        }) +
                                        ")"}
                                </Typography>
                            </Box>
                            <Typography variant={"caption"}>
                                {translate("Added when", {
                                    when: props.config.created_at,
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
                            </Typography>
                            <Typography variant={"caption"}>
                                {props.config.last_used_at === undefined
                                    ? translate("Never used")
                                    : translate("Last Used when", {
                                          when: props.config.last_used_at,
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
                            </Typography>
                        </Stack>
                        <Tooltip title={translate("Remove the Time-based One Time Password configuration")}>
                            <Button
                                variant={"outlined"}
                                color={"error"}
                                startIcon={
                                    loadingDelete ? <CircularProgress color="inherit" size={20} /> : <DeleteIcon />
                                }
                                onClick={loadingDelete ? undefined : () => setShowDialogDelete(true)}
                            >
                                {translate("Remove")}
                            </Button>
                        </Tooltip>
                    </Stack>
                </Box>
            </Paper>
        </Fragment>
    );
};

export default TOTPDevice;
