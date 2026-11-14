// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/UI/Dialog";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { useNotifications } from "@contexts/NotificationsContext";
import { AppAuthenticatorDevice, updateAppAuthenticatorDevice } from "@services/AppAuthenticator";

const MaxNameLength = 64;

function truncate(value: string) {
    return Array.from(value).slice(0, MaxNameLength).join("");
}

interface Props {
    open: boolean;
    device?: AppAuthenticatorDevice;
    handleClose: () => void;
}

const AppAuthenticatorDeviceEditDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [name, setName] = useState("");
    const [errorName, setErrorName] = useState(false);

    const handleReset = () => {
        setErrorName(false);
        setName("");
    };

    const handleCancel = () => {
        props.handleClose();
        handleReset();
    };

    const handleRename = async (value: string) => {
        if (!props.device) return;

        const item = translate("Authelia app device");

        try {
            const response = await updateAppAuthenticatorDevice(props.device.id, value);

            if (response.data.status === "KO") {
                if (response.data.elevation) {
                    createErrorNotification(
                        translate("You must be elevated to {{action}} a {{item}}", {
                            action: translate("update"),
                            item,
                        }),
                    );
                } else if (response.data.authentication) {
                    createErrorNotification(
                        translate("You must have a higher authentication level to {{action}} a {{item}}", {
                            action: translate("update"),
                            item,
                        }),
                    );
                } else {
                    createErrorNotification(
                        translate("There was a problem {{action}} the {{item}}", {
                            action: translate("updating"),
                            item,
                        }),
                    );
                }

                return;
            }

            createSuccessNotification(
                translate("Successfully {{action}} the {{item}}", { action: translate("updated"), item }),
            );
        } catch (err) {
            console.error(err);
            createErrorNotification(
                translate("There was a problem {{action}} the {{item}}", { action: translate("updating"), item }),
            );
        }
    };

    const handleUpdate = async () => {
        const value = truncate(name.trim());

        if (value.length === 0) {
            setErrorName(true);

            return;
        }

        await handleRename(value);

        props.handleClose();
        handleReset();
    };

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleCancel();
            }}
        >
            <DialogContent showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Rename device")}</DialogTitle>
                    <DialogDescription>{translate("Enter a new name for this device")}</DialogDescription>
                </DialogHeader>
                <div className="grid gap-2">
                    <Label htmlFor="app-authenticator-device-name">{translate("Name")}</Label>
                    <Input
                        id="app-authenticator-device-name"
                        required
                        value={name}
                        error={errorName}
                        placeholder={props.device?.name}
                        onChange={(v) => {
                            setName(truncate(v.target.value));
                            setErrorName(false);
                        }}
                        autoCapitalize="none"
                        onKeyDown={(ev) => {
                            if (ev.key === "Enter") {
                                ev.preventDefault();
                                handleUpdate().catch(console.error);
                            }
                        }}
                    />
                </div>
                <DialogFooter>
                    <Button id={"dialog-cancel"} variant={"ghost"} color={"primary"} onClick={handleCancel}>
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id={"dialog-update"}
                        variant={"ghost"}
                        color={"primary"}
                        disabled={name.trim().length === 0}
                        onClick={() => handleUpdate().catch(console.error)}
                    >
                        {translate("Update")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default AppAuthenticatorDeviceEditDialog;
