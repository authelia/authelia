import { Fragment, useCallback, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Card, CardContent } from "@components/UI/Card";
import { Spinner } from "@components/UI/Spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
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
    config: null | undefined | UserInfoTOTPConfiguration;
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
            }
        },
        [
            handleResetState,
            handleOpenDialogRegister,
            handleOpenDialogDelete,
            dialogRegisterOpening,
            dialogDeleteOpening,
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
            <Card>
                <CardContent className="grid grid-cols-12 gap-4 p-4">
                    <div className="col-span-12">
                        <h5 className="text-xl font-semibold">{translate("One-Time Password")}</h5>
                    </div>
                    <div className="col-span-12">
                        <TooltipProvider>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <span>
                                        <Button
                                            id={"one-time-password-add"}
                                            variant="outline"
                                            onClick={handleRegister}
                                            disabled={registered || dialogRegisterOpening || dialogRegisterOpen}
                                        >
                                            {dialogRegisterOpening ? <Spinner size={20} /> : null}
                                            {translate("Add")}
                                        </Button>
                                    </span>
                                </TooltipTrigger>
                                <TooltipContent>
                                    {registered
                                        ? translate("You can only register a single One-Time Password")
                                        : translate("Click to add a {{item}} to your account", {
                                              item: translate("One-Time Password"),
                                          })}
                                </TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    </div>
                    {props.config === null || props.config === undefined ? (
                        <div className="col-span-12">
                            <p className="text-sm text-muted-foreground">
                                {translate(
                                    "The One-Time Password has not been registered if you'd like to register it click add",
                                )}
                            </p>
                        </div>
                    ) : (
                        <div className="col-span-12 md:col-span-6 xl:col-span-3">
                            <OneTimePasswordConfiguration
                                config={props.config}
                                handleInformation={handleInformation}
                                handleDelete={handleDelete}
                            />
                        </div>
                    )}
                </CardContent>
            </Card>
        </Fragment>
    );
};

export default OneTimePasswordPanel;
