import { Fragment, useCallback, useEffect, useState } from "react";

import { Box, Button, Container, List, ListItem, Paper, Stack, Typography, useTheme } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { useUserInfoPOST } from "@hooks/UserInfo";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import ChangePasswordDialog from "@views/Settings/Security/ChangePasswordDialog";

export interface Props {}

const SettingsView = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const theme = useTheme();

    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();
    const [elevation, setElevation] = useState<UserSessionElevation>();
    const [dialogSFOpening, setDialogSFOpening] = useState(false);
    const [dialogIVOpening, setDialogIVOpening] = useState(false);
    const [dialogPWChangeOpen, setDialogPWChangeOpen] = useState(false);
    const [dialogPWChangeOpening, setDialogPWChangeOpening] = useState(false);
    const { createErrorNotification } = useNotifications();

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
        const result = await getUserSessionElevation();

        setElevation(result);
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
    }, [fetchUserInfoError, createErrorNotification, translate]);

    useEffect(() => {
        fetchUserInfo();
    }, [fetchUserInfo]);

    //console.log(userInfo);

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
                maxWidth="md"
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
                        <Box sx={{ p: 5 }}>
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
                                    {translate("Username: ")} {userInfo?.display_name || ""}
                                </Typography>
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
                                <Box display="flex" alignItems="center">
                                    <Typography sx={{ mr: 1 }}>Email:</Typography>
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
                                sx={{ p: 1.25, mb: 1, border: `1px solid ${theme.palette.grey[600]}`, borderRadius: 1 }}
                            >
                                <Typography>{translate("Password")}: ●●●●●●●●</Typography>
                            </Box>
                            <Button variant="contained" sx={{ p: 1, width: "100%" }} onClick={handleChangePassword}>
                                {translate("Reset Password")}
                            </Button>
                        </Box>
                    </Stack>
                </Paper>
            </Container>
        </Fragment>
    );
};

export default SettingsView;
