// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Card, CardContent } from "@components/UI/Card";
import { Spinner } from "@components/UI/Spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { useNotifications } from "@contexts/NotificationsContext";
import { UserInfo } from "@models/UserInfo";
import { AppAuthenticatorDevice, deleteAppAuthenticatorDevice } from "@services/AppAuthenticator";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import AppAuthenticatorDeviceEditDialog from "@views/Settings/TwoFactorAuthentication/AppAuthenticatorDeviceEditDialog";
import AppAuthenticatorDeviceItem from "@views/Settings/TwoFactorAuthentication/AppAuthenticatorDeviceItem";
import AppAuthenticatorRegisterDialog from "@views/Settings/TwoFactorAuthentication/AppAuthenticatorRegisterDialog";
import DeleteDialog from "@views/Settings/TwoFactorAuthentication/DeleteDialog";

type Action = "delete" | "edit" | "register";

interface Props {
    info?: UserInfo;
    devices: AppAuthenticatorDevice[] | undefined;
    registrationAvailable: boolean;
    handleRefreshState: () => void;
}

const AppAuthenticatorPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [elevation, setElevation] = useState<UserSessionElevation>();

    const [dialogSFOpening, setDialogSFOpening] = useState(false);
    const [dialogIVOpening, setDialogIVOpening] = useState(false);

    const [pending, setPending] = useState<Action>();
    const [open, setOpen] = useState<Action>();
    const [index, setIndex] = useState(-1);

    const device = index === -1 || !props.devices ? undefined : props.devices[index];

    const handleResetState = useCallback(() => {
        setDialogSFOpening(false);
        setDialogIVOpening(false);
        setElevation(undefined);
        setPending(undefined);
        setOpen(undefined);
        setIndex(-1);
    }, []);

    const handleOpenPending = useCallback(() => {
        setDialogSFOpening(false);
        setDialogIVOpening(false);
        setElevation(undefined);
        setOpen(pending);
        setPending(undefined);
    }, [pending]);

    const handleClosed = useCallback(() => {
        handleResetState();
        props.handleRefreshState();
    }, [handleResetState, props]);

    const handleElevationRefresh = async () => {
        const result = await getUserSessionElevation();

        setElevation(result);

        return result;
    };

    const handleSFDialogClosed = (ok: boolean, changed: boolean) => {
        if (!ok) {
            console.warn("Second Factor dialog close callback failed, it was likely cancelled by the user.");

            handleResetState();

            return;
        }

        const proceed = (current?: UserSessionElevation) => {
            if (current && (current.elevated || current.skip_second_factor)) {
                handleOpenPending();
            } else {
                setDialogIVOpening(true);
            }
        };

        if (changed) {
            handleElevationRefresh()
                .then((refreshed) => proceed(refreshed))
                .catch(console.error);
        } else {
            proceed(elevation);
        }
    };

    const handleIVDialogClosed = (ok: boolean) => {
        if (!ok) {
            console.warn("Identity Verification dialog close callback failed, it was likely cancelled by the user.");

            handleResetState();

            return;
        }

        handleOpenPending();
    };

    const handleElevation = (action: Action, target: number) => {
        setPending(action);
        setIndex(target);

        handleElevationRefresh().catch(console.error);

        setDialogSFOpening(true);
    };

    const handleRegister = () => {
        handleElevation("register", -1);
    };

    const handleEdit = (target: number) => {
        if (!props.devices || target < 0 || target >= props.devices.length) return;

        handleElevation("edit", target);
    };

    const handleDelete = (target: number) => {
        if (!props.devices || target < 0 || target >= props.devices.length) return;

        handleElevation("delete", target);
    };

    const handleRemove = async () => {
        if (!device) return;

        const item = translate("Authelia app device");

        try {
            const response = await deleteAppAuthenticatorDevice(device.id);

            if (response.data.status === "KO") {
                if (response.data.elevation) {
                    createErrorNotification(
                        translate("You must be elevated to {{action}} a {{item}}", {
                            action: translate("delete"),
                            item,
                        }),
                    );
                } else if (response.data.authentication) {
                    createErrorNotification(
                        translate("You must have a higher authentication level to {{action}} a {{item}}", {
                            action: translate("delete"),
                            item,
                        }),
                    );
                } else {
                    createErrorNotification(
                        translate("There was a problem {{action}} the {{item}}", {
                            action: translate("deleting"),
                            item,
                        }),
                    );
                }

                return;
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(
                translate("There was a problem {{action}} the {{item}}", { action: translate("deleting"), item }),
            );

            return;
        }

        createSuccessNotification(
            translate("Successfully {{action}} the {{item}}", { action: translate("deleted"), item }),
        );

        handleClosed();
    };

    const registering = pending === "register" || open === "register";

    return (
        <Fragment>
            <SecondFactorDialog
                info={props.info}
                elevation={elevation}
                opening={dialogSFOpening}
                handleClosed={handleSFDialogClosed}
                handleOpened={() => setDialogSFOpening(false)}
            />
            <IdentityVerificationDialog
                elevation={elevation}
                opening={dialogIVOpening}
                handleClosed={handleIVDialogClosed}
                handleOpened={() => setDialogIVOpening(false)}
            />
            <AppAuthenticatorRegisterDialog open={open === "register"} setClosed={handleClosed} />
            <AppAuthenticatorDeviceEditDialog open={open === "edit"} device={device} handleClose={handleClosed} />
            <DeleteDialog
                open={open === "delete"}
                onConfirm={() => handleRemove().catch(console.error)}
                onCancel={handleResetState}
                title={translate("Remove {{item}}", { item: translate("Authelia app device") })}
                text={translate("Are you sure you want to remove the Authelia app device from your account", {
                    name: device?.name,
                })}
            />
            <Card id={"app-authenticator-panel"} data-loading={props.devices === undefined ? "true" : "false"}>
                <CardContent className="grid grid-cols-12 gap-4 p-4">
                    <div className="col-span-12">
                        <h5 className="text-xl font-semibold">{translate("Authelia app")}</h5>
                    </div>
                    <div className="col-span-12">
                        {props.registrationAvailable ? (
                            <TooltipProvider>
                                <Tooltip>
                                    <TooltipTrigger
                                        render={
                                            <Button
                                                id={"app-authenticator-device-add"}
                                                variant={"outline"}
                                                color={"primary"}
                                                onClick={handleRegister}
                                                disabled={registering}
                                            >
                                                {pending === "register" ? <Spinner size={20} /> : null}
                                                {translate("Add")}
                                            </Button>
                                        }
                                    />
                                    <TooltipContent>
                                        {translate("Click to add a {{item}} to your account", {
                                            item: translate("Authelia app device"),
                                        })}
                                    </TooltipContent>
                                </Tooltip>
                            </TooltipProvider>
                        ) : (
                            <p
                                id={"app-authenticator-registration-unavailable"}
                                className="text-sm text-muted-foreground"
                            >
                                {translate("Registration of the Authelia app is not configured on this server")}
                            </p>
                        )}
                    </div>
                    <div className="col-span-12">
                        {props.devices === undefined ? (
                            <Spinner size={20} />
                        ) : props.devices.length === 0 ? (
                            <p className={props.registrationAvailable ? "text-sm text-muted-foreground" : "hidden"}>
                                {translate(
                                    "No Authelia app devices have been registered if you'd like to register one click add",
                                )}
                            </p>
                        ) : (
                            <div className="grid grid-cols-1 gap-3">
                                {props.devices.map((value, i) => (
                                    <AppAuthenticatorDeviceItem
                                        key={value.id}
                                        index={i}
                                        device={value}
                                        handleEdit={handleEdit}
                                        handleDelete={handleDelete}
                                    />
                                ))}
                            </div>
                        )}
                    </div>
                </CardContent>
            </Card>
        </Fragment>
    );
};

export default AppAuthenticatorPanel;
