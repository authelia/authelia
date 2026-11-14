// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useEffectEvent, useRef, useState } from "react";

import axios from "axios";
import { QRCodeSVG } from "qrcode.react";
import { useTranslation } from "react-i18next";

import CopyButton from "@components/CopyButton";
import SuccessIcon from "@components/SuccessIcon";
import { Button } from "@components/UI/Button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/UI/Dialog";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { useNotifications } from "@contexts/NotificationsContext";
import { useAbortSignal } from "@hooks/Abort";
import { useSecondsUntil } from "@hooks/Countdown";
import {
    AppAuthenticatorRegistrationDevice,
    AppAuthenticatorRegistrationStart,
    AppAuthenticatorRegistrationStatus,
    cancelAppAuthenticatorRegistration,
    getAppAuthenticatorRegistration,
    startAppAuthenticatorRegistration,
} from "@services/AppAuthenticator";

const PollIntervalMs = 2000;

type Status = "confirm" | "failed" | "registered" | "starting" | "waiting";

interface Confirmation {
    number: string;
    device: AppAuthenticatorRegistrationDevice;
    expires_at: number;
}

interface Props {
    open: boolean;
    setClosed: () => void;
}

const AppAuthenticatorRegisterDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const getSignal = useAbortSignal();
    const [status, setStatus] = useState<Status>("starting");
    const [registration, setRegistration] = useState<AppAuthenticatorRegistrationStart>();
    const [confirmation, setConfirmation] = useState<Confirmation>();
    const [poll, setPoll] = useState(0);

    const pendingRef = useRef(false);
    const timeoutSuccessRef = useRef<null | ReturnType<typeof setTimeout>>(null);

    const expiresAt =
        status === "waiting" ? registration?.expires_at : status === "confirm" ? confirmation?.expires_at : undefined;
    const seconds = useSecondsUntil(expiresAt);

    useEffect(() => {
        return () => {
            if (timeoutSuccessRef.current !== null) {
                clearTimeout(timeoutSuccessRef.current);
                timeoutSuccessRef.current = null;
            }
        };
    }, []);

    const resetState = useCallback(() => {
        pendingRef.current = false;
        setRegistration(undefined);
        setConfirmation(undefined);
        setStatus("starting");
    }, []);

    const handleRegistered = useEffectEvent(() => {
        pendingRef.current = false;
        setStatus("registered");

        createSuccessNotification(
            translate("Successfully {{action}} the {{item}}", {
                action: translate("added"),
                item: translate("Authelia app device"),
            }),
        );

        timeoutSuccessRef.current = setTimeout(() => {
            timeoutSuccessRef.current = null;
            props.setClosed();
            resetState();
        }, 1500);
    });

    const handleStatus = useEffectEvent((res: AppAuthenticatorRegistrationStatus) => {
        switch (res.state) {
            case "waiting":
                setPoll((value) => value + 1);
                break;
            case "confirm_number":
                setConfirmation({ device: res.device, expires_at: res.expires_at, number: res.number });
                setStatus("confirm");
                setPoll((value) => value + 1);
                break;
            case "registered":
                handleRegistered();
                break;
            default:
                pendingRef.current = false;
                setStatus("failed");
        }
    });

    useEffect(() => {
        if (!props.open || status !== "starting") return;

        let active = true;

        startAppAuthenticatorRegistration(getSignal())
            .then((res) => {
                if (!active) return;

                pendingRef.current = true;
                setRegistration(res);
                setStatus("waiting");
            })
            .catch((err) => {
                if (!active || axios.isCancel(err)) return;

                console.error(err);
                createErrorNotification(
                    translate("There was a problem {{action}} the {{item}}", {
                        action: translate("adding"),
                        item: translate("Authelia app device"),
                    }),
                );
                setStatus("failed");
            });

        return () => {
            active = false;
        };
    }, [props.open, status, getSignal, createErrorNotification, translate]);

    useEffect(() => {
        if (!props.open || (status !== "waiting" && status !== "confirm")) return;

        let active = true;

        const timeout = setTimeout(() => {
            getAppAuthenticatorRegistration(getSignal())
                .then((res) => {
                    if (active) handleStatus(res);
                })
                .catch((err) => {
                    if (!active || axios.isCancel(err)) return;

                    console.error(err);
                    pendingRef.current = false;
                    setStatus("failed");
                });
        }, PollIntervalMs);

        return () => {
            active = false;
            clearTimeout(timeout);
        };
    }, [props.open, status, poll, getSignal]);

    const handleClose = () => {
        if (timeoutSuccessRef.current !== null) {
            clearTimeout(timeoutSuccessRef.current);
            timeoutSuccessRef.current = null;
        }

        if (pendingRef.current) {
            cancelAppAuthenticatorRegistration().catch(console.error);
        }

        props.setClosed();
        resetState();
    };

    const handleStartAgain = () => {
        resetState();
    };

    const renderContent = () => {
        switch (status) {
            case "starting":
                return <Spinner size={32} />;
            case "waiting":
                return (
                    <div className="flex flex-col items-center gap-2">
                        <p>{translate("Scan this QR code with the Authelia app")}</p>
                        {registration ? (
                            <QRCodeSVG value={registration.qr} size={200} className="inline-block my-4 p-2 bg-white" />
                        ) : null}
                        <Label htmlFor="app-authenticator-qr" className="self-start">
                            {translate("QR code text")}
                        </Label>
                        <textarea
                            id="app-authenticator-qr"
                            className="w-full resize-none overflow-hidden rounded-md border border-input bg-transparent px-3 py-2 text-sm font-mono shadow-xs outline-none"
                            value={registration?.qr ?? ""}
                            readOnly
                            rows={registration ? Math.ceil(registration.qr.length / 40) : 1}
                        />
                        <CopyButton
                            variant={"default"}
                            tooltip={translate("Click to Copy")}
                            value={registration?.qr ?? null}
                            childrenCopied={translate("Copied")}
                            fullWidth={true}
                        >
                            {translate("Copy")}
                        </CopyButton>
                        <p id="app-authenticator-register-countdown" className="text-sm text-muted-foreground">
                            {translate("Expires in {{count}} seconds", { count: seconds })}
                        </p>
                    </div>
                );
            case "confirm":
                return (
                    <div className="flex flex-col items-center gap-2">
                        <p
                            id="app-authenticator-register-number"
                            className="font-mono text-5xl font-bold tracking-[0.3em]"
                            aria-live="polite"
                        >
                            {confirmation?.number}
                        </p>
                        <p id="app-authenticator-register-device" className="font-semibold">
                            {confirmation ? `${confirmation.device.name}, ${confirmation.device.platform}` : null}
                        </p>
                        <p>{translate("Enter this number in the Authelia app.")}</p>
                        <p id="app-authenticator-register-countdown" className="text-sm text-muted-foreground">
                            {translate("Expires in {{count}} seconds", { count: seconds })}
                        </p>
                    </div>
                );
            case "registered":
                return <SuccessIcon />;
            case "failed":
                return (
                    <div className="flex flex-col items-center gap-2">
                        <p id="app-authenticator-register-failed">{translate("Registration failed. Start again.")}</p>
                        <Button
                            id="app-authenticator-register-start-again"
                            variant={"outline"}
                            onClick={handleStartAgain}
                        >
                            {translate("Start again")}
                        </Button>
                    </div>
                );
        }
    };

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open && props.open) handleClose();
            }}
        >
            <DialogContent id={"dialog-register-app-authenticator"} showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Register an Authelia app device")}</DialogTitle>
                    <DialogDescription>
                        {translate("Use the Authelia app on your phone to scan the QR code")}
                    </DialogDescription>
                </DialogHeader>
                <div className="flex min-h-[200px] flex-col items-center justify-center">{renderContent()}</div>
                <DialogFooter>
                    <Button id={"dialog-cancel"} variant={"outline"} color={"destructive"} onClick={handleClose}>
                        {translate("Cancel")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default AppAuthenticatorRegisterDialog;
