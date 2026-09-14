// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Card, CardContent } from "@components/UI/Card";
import { Spinner } from "@components/UI/Spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { useElevationFlow } from "@hooks/ElevationFlow";
import { UserInfoTOTPConfiguration } from "@models/TOTPConfiguration";
import { UserInfo } from "@models/UserInfo";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import ReauthenticationDialog from "@views/Settings/Common/ReauthenticationDialog";
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

    const [dialogInformationOpen, setDialogInformationOpen] = useState(false);
    const [dialogRegisterOpen, setDialogRegisterOpen] = useState(false);
    const [dialogDeleteOpen, setDialogDeleteOpen] = useState(false);

    const handleElevated = useCallback((action: "delete" | "register") => {
        if (action === "register") {
            setDialogRegisterOpen(true);
        } else {
            setDialogDeleteOpen(true);
        }
    }, []);

    const handleCancelled = useCallback(() => {
        setDialogRegisterOpen(false);
        setDialogDeleteOpen(false);
    }, []);

    const {
        identityVerificationDialogProps,
        pending,
        reauthenticationDialogProps,
        reset,
        secondFactorDialogProps,
        start,
    } = useElevationFlow<"delete" | "register">({ onCancelled: handleCancelled, onElevated: handleElevated });

    const handleResetState = useCallback(() => {
        reset();
        handleCancelled();
    }, [handleCancelled, reset]);

    const handleInformation = () => {
        setDialogInformationOpen(true);
    };

    const handleRegister = () => {
        start("register");
    };

    const handleDelete = () => {
        if (!props.config) return;

        start("delete");
    };

    const dialogRegisterOpening = pending === "register";

    const registered = props.config !== null && props.config !== undefined;

    return (
        <Fragment>
            <ReauthenticationDialog info={props.info} {...reauthenticationDialogProps} />
            <SecondFactorDialog info={props.info} {...secondFactorDialogProps} />
            <IdentityVerificationDialog {...identityVerificationDialogProps} />
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
            <Card id={"one-time-password-panel"} data-loading={props.config === undefined ? "true" : "false"}>
                <CardContent className="grid grid-cols-12 gap-4 p-4">
                    <div className="col-span-12">
                        <h5 className="text-xl font-semibold">{translate("One-Time Password")}</h5>
                    </div>
                    <div className="col-span-12">
                        <TooltipProvider>
                            <Tooltip>
                                <TooltipTrigger
                                    render={
                                        <span>
                                            <Button
                                                id={"one-time-password-add"}
                                                variant={"outline"}
                                                color={"primary"}
                                                onClick={handleRegister}
                                                disabled={
                                                    props.config === undefined ||
                                                    registered ||
                                                    dialogRegisterOpening ||
                                                    dialogRegisterOpen
                                                }
                                            >
                                                {dialogRegisterOpening ? <Spinner size={20} /> : null}
                                                {translate("Add")}
                                            </Button>
                                        </span>
                                    }
                                />
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
                    {props.config === undefined ? (
                        <div className="col-span-12">
                            <Spinner size={20} />
                        </div>
                    ) : props.config === null ? (
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
