import { Fragment, useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Card } from "@components/UI/Card";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
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
            variant="default"
            className="p-2 w-full"
            onClick={handleChangePassword}
            disabled={!configuration || configuration.password_change_disabled}
        >
            {translate("Change Password")}
        </Button>
    );

    return !configuration || configuration.password_change_disabled ? (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>
                    <span>{buttonContent}</span>
                </TooltipTrigger>
                <TooltipContent>{translate("This is disabled by your administrator")}</TooltipContent>
            </Tooltip>
        </TooltipProvider>
    ) : (
        buttonContent
    );
};

const SettingsView = function () {
    const { t: translate } = useTranslation(["settings", "portal"]);
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

            <div className="flex items-start justify-center h-screen pt-16">
                <Card className="flex items-center justify-center h-auto">
                    <div className="flex flex-col gap-4 m-4 w-full">
                        <div className="p-2 md:p-6">
                            <div className="border border-muted-foreground/50 rounded mb-2 p-2.5 w-full">
                                <p>
                                    {translate("Name")}: {userInfo?.display_name || ""}
                                </p>
                            </div>
                            <div className="border border-muted-foreground/50 rounded mb-2 p-2.5 w-full">
                                <div className="flex items-center">
                                    <p className="mr-2">{translate("Email")}:</p>
                                    <p>{userInfo?.emails?.[0] || ""}</p>
                                </div>
                                {userInfo?.emails && userInfo.emails.length > 1 && (
                                    <ul className="p-0 pl-8 w-full">
                                        {" "}
                                        {userInfo.emails.slice(1).map((email: string) => (
                                            <li key={email} className="py-0">
                                                <p>{email}</p>
                                            </li>
                                        ))}
                                    </ul>
                                )}
                            </div>
                            <div className="border border-muted-foreground/50 rounded mb-2 p-2.5">
                                <p>{translate("Password")}: ●●●●●●●●</p>
                            </div>
                            <PasswordChangeButton
                                configuration={configuration}
                                translate={translate}
                                handleChangePassword={handleChangePassword}
                            />
                        </div>
                    </div>
                </Card>
            </div>
        </Fragment>
    );
};

export default SettingsView;
