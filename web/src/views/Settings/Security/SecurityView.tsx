import { Fragment, useCallback, useEffect, useState } from "react";

import { Box, Button, Container, List, ListItem, Paper, Stack, Tooltip, Typography, useTheme } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useConfiguration } from "@hooks/Configuration";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoGET } from "@hooks/UserInfo";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import ChangePasswordDialog from "@views/Settings/Security/ChangePasswordDialog";

const SettingsView = function () {
    const { t: translate } = useTranslation("settings");
    const theme = useTheme();
    const { createErrorNotification } = useNotifications();

    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoGET();
    const [elevation, setElevation] = useState<UserSessionElevation>();
    const [dialogSFOpening, setDialogSFOpening] = useState(false);
    const [dialogIVOpening, setDialogIVOpening] = useState(false);
    const [dialogPWChangeOpen, setDialogPWChangeOpen] = useState(false);
    const [dialogPWChangeOpening, setDialogPWChangeOpening] = useState(false);
    const [configuration, fetchConfiguration, , fetchConfigurationError] = useConfiguration();

    const handleResetStateOpening = () => {
        setDialogSFOpening(false);
        setDialogIVOpening(false);
        setDialogPWChangeOpening(false);
    };

    const handleResetState = useCallback(() => {
        handleResetStateOpening();

        setElevation(undefined);
        setDialogPWChangeOpen(false);
    }, []);

    const handleOpenChangePWDialog = useCallback(() => {
        handleResetStateOpening();
        setDialogPWChangeOpen(true);
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
            if (dialogPWChangeOpening) {
                handleOpenChangePWDialog();
            }
        },
        [dialogPWChangeOpening, handleOpenChangePWDialog, handleResetState],
    );

    const handleIVDialogOpened = () => {
        setDialogIVOpening(false);
    };

    const handleElevationRefresh = async () => {
        try {
            const result = await getUserSessionElevation();
            setElevation(result);
        } catch {
            createErrorNotification(translate("Failed to get session elevation status"));
        }
    };

    const handleElevation = () => {
        handleElevationRefresh().catch(console.error);

        setDialogSFOpening(true);
    };

    const handleChangePassword = () => {
        setDialogPWChangeOpening(true);

        handleElevation();
    };

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(translate("There was an issue retrieving user preferences"));
        }
        if (fetchConfigurationError) {
            createErrorNotification(translate("There was an issue retrieving configuration"));
        }
    }, [fetchUserInfoError, fetchConfigurationError, createErrorNotification, translate]);

    useEffect(() => {
        fetchUserInfo();
        fetchConfiguration();
    }, [fetchUserInfo, fetchConfiguration]);

    const PasswordChangeButton = () => {
        const buttonContent = (
            <Button
                id="change-password-button"
                variant="contained"
                sx={{ p: 1, width: "100%" }}
                onClick={handleChangePassword}
                disabled={configuration?.password_change_disabled || false}
            >
                {translate("Change Password")}
            </Button>
        );

        return configuration?.password_change_disabled ? (
            <Tooltip title={translate("This is disabled by your administrator.")}>
                <span>{buttonContent}</span>
            </Tooltip>
        ) : (
            buttonContent
        );
    };

    return (
        <Fragment>
            <SecondFactorDialog
                info={userInfo}
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
            <ChangePasswordDialog
                username={userInfo?.display_name || ""}
                open={dialogPWChangeOpen}
                setClosed={() => {
                    handleResetState();
                }}
            />

            <Container
                sx={{
                    display: "flex",
                    justifyContent: "center",
                    alignItems: "flex-start",
                    height: "100vh",
                    pt: 8,
                }}
            >
                <Paper
                    variant="outlined"
                    sx={{
                        display: "flex",
                        justifyContent: "center",
                        alignItems: "center",
                        height: "auto",
                    }}
                >
                    <Stack spacing={2} sx={{ m: 2, width: "100%" }}>
                        <Box sx={{ p: { xs: 1, md: 3 } }}>
                            <Box
                                sx={{
                                    width: "100%",
                                    p: 1.25,
                                    mb: 1,
                                    border: `1px solid ${theme.palette.grey[600]}`,
                                    borderRadius: 1,
                                }}
                            >
                                <Box display="flex" alignItems="center">
                                    <Typography sx={{ mr: 1 }}>{translate("Email")}:</Typography>
                                    <Typography>{userInfo?.emails?.[0] || ""}</Typography>
                                </Box>
                                {userInfo?.emails && userInfo.emails.length > 1 && (
                                    <List sx={{ width: "100%", padding: 0, pl: 4 }}>
                                        {" "}
                                        {userInfo.emails.slice(1).map((email: string, index: number) => (
                                            <ListItem key={index} sx={{ paddingTop: 0, paddingBottom: 0 }}>
                                                <Typography>{email}</Typography>
                                            </ListItem>
                                        ))}
                                    </List>
                                )}
                            </Box>
                            <Box
                                sx={{
                                    width: "100%",
                                    p: 1.25,
                                    mb: 1,
                                    border: `1px solid ${theme.palette.grey[600]}`,
                                    borderRadius: 1,
                                }}
                            >
                                <Typography>
                                    {translate("Username")}: {userInfo?.display_name || ""}
                                </Typography>
                            </Box>
                            <Box
                                sx={{ p: 1.25, mb: 1, border: `1px solid ${theme.palette.grey[600]}`, borderRadius: 1 }}
                            >
                                <Typography>{translate("Password")}: ●●●●●●●●</Typography>
                            </Box>
                            <PasswordChangeButton />
                        </Box>
                    </Stack>
                </Paper>
            </Container>
        </Fragment>
    );
};

export default SettingsView;
