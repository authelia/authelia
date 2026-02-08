import { Fragment, useCallback, useEffect, useState } from "react";

import { Box, Button, Container, List, ListItem, Paper, Stack, Tooltip, Typography, useTheme } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useConfiguration } from "@hooks/Configuration";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoGET } from "@hooks/UserInfo";
import { Configuration } from "@models/Configuration";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import ChangePasswordDialog from "@views/Settings/Security/ChangePasswordDialog";

interface PasswordChangeButtonProps {
    configuration: Configuration | undefined;
    translate: (_key: string) => string;
    handleChangePassword: () => void;
}

const PasswordChangeButton = ({ configuration, handleChangePassword, translate }: PasswordChangeButtonProps) => {
    const buttonContent = (
        <Button
            id="change-password-button"
            variant="contained"
            sx={{ p: 1, width: "100%" }}
            onClick={handleChangePassword}
            disabled={!configuration || configuration.password_change_disabled}
        >
            {translate("Change Password")}
        </Button>
    );

    return !configuration || configuration.password_change_disabled ? (
        <Tooltip title={translate("This is disabled by your administrator")}>
            <Box component={"span"}>{buttonContent}</Box>
        </Tooltip>
    ) : (
        buttonContent
    );
};

const SettingsView = function () {
    const { t: translate } = useTranslation(["settings", "portal"]);
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
                .then((refreshedElevation) => {
                    if (refreshedElevation) {
                        const isElevatedFromRefresh =
                            refreshedElevation.elevated || refreshedElevation.skip_second_factor;
                        if (isElevatedFromRefresh) {
                            setElevation(undefined);
                            if (dialogPWChangeOpening) {
                                handleOpenChangePWDialog();
                            }
                        } else {
                            setDialogIVOpening(true);
                        }
                    }
                })
                .catch((error) => {
                    console.error(error);
                    createErrorNotification(translate("Failed to get session elevation status"));
                });
        } else {
            const isElevated = elevation && (elevation.elevated || elevation.skip_second_factor);
            if (isElevated) {
                setElevation(undefined);
                if (dialogPWChangeOpening) {
                    handleOpenChangePWDialog();
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
        const result = await getUserSessionElevation();
        setElevation(result);
        return result;
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
            createErrorNotification(translate("There was an issue retrieving user preferences", { ns: "portal" }));
        }
        if (fetchConfigurationError) {
            createErrorNotification(translate("There was an issue retrieving configuration"));
        }
    }, [fetchUserInfoError, fetchConfigurationError, createErrorNotification, translate]);

    useEffect(() => {
        fetchUserInfo();
        fetchConfiguration();
    }, [fetchUserInfo, fetchConfiguration]);

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
                    alignItems: "flex-start",
                    display: "flex",
                    height: "100vh",
                    justifyContent: "center",
                    pt: 8,
                }}
            >
                <Paper
                    variant="outlined"
                    sx={{
                        alignItems: "center",
                        display: "flex",
                        height: "auto",
                        justifyContent: "center",
                    }}
                >
                    <Stack spacing={2} sx={{ m: 2, width: "100%" }}>
                        <Box sx={{ p: { md: 3, xs: 1 } }}>
                            <Box
                                sx={{
                                    border: `1px solid ${theme.palette.grey[600]}`,
                                    borderRadius: 1,
                                    mb: 1,
                                    p: 1.25,
                                    width: "100%",
                                }}
                            >
                                <Typography>
                                    {translate("Name")}: {userInfo?.display_name || ""}
                                </Typography>
                            </Box>
                            <Box
                                sx={{
                                    border: `1px solid ${theme.palette.grey[600]}`,
                                    borderRadius: 1,
                                    mb: 1,
                                    p: 1.25,
                                    width: "100%",
                                }}
                            >
                                <Box display="flex" alignItems="center">
                                    <Typography sx={{ mr: 1 }}>{translate("Email")}:</Typography>
                                    <Typography>{userInfo?.emails?.[0] || ""}</Typography>
                                </Box>
                                {userInfo?.emails && userInfo.emails.length > 1 && (
                                    <List sx={{ padding: 0, pl: 4, width: "100%" }}>
                                        {" "}
                                        {userInfo.emails.slice(1).map((email: string) => (
                                            <ListItem key={email} sx={{ paddingBottom: 0, paddingTop: 0 }}>
                                                <Typography>{email}</Typography>
                                            </ListItem>
                                        ))}
                                    </List>
                                )}
                            </Box>
                            <Box
                                sx={{ border: `1px solid ${theme.palette.grey[600]}`, borderRadius: 1, mb: 1, p: 1.25 }}
                            >
                                <Typography>{translate("Password")}: ●●●●●●●●</Typography>
                            </Box>
                            <PasswordChangeButton
                                configuration={configuration}
                                translate={translate}
                                handleChangePassword={handleChangePassword}
                            />
                        </Box>
                    </Stack>
                </Paper>
            </Container>
        </Fragment>
    );
};

export default SettingsView;
