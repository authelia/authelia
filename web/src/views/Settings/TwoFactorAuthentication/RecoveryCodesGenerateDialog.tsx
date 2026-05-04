// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import CopyButton from "@components/CopyButton";
import { Alert, AlertDescription } from "@components/UI/Alert";
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
import { Spinner } from "@components/UI/Spinner";
import { useGenerateRecoveryCodes } from "@hooks/UserRecoveryCodes";

interface Props {
    open: boolean;
    setClosed: () => void;
}

function normalize(input: string): string {
    return input.replace(/[\s\-_]/g, "").toUpperCase();
}

function buildDownloadFilename(): string {
    const today = new Date();
    const yyyy = today.getFullYear();
    const mm = String(today.getMonth() + 1).padStart(2, "0");
    const dd = String(today.getDate()).padStart(2, "0");

    return `authelia-recovery-codes-${yyyy}-${mm}-${dd}.txt`;
}

function buildDownloadBody(codes: string[], generatedAt: Date): string {
    const header = [
        `Authelia recovery codes`,
        `Generated: ${generatedAt.toISOString()}`,
        `Anyone with this file can sign in to your account. Treat it like a password.`,
        ``,
    ].join("\n");

    return header + codes.join("\n") + "\n";
}

const RecoveryCodesGenerateDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [response, triggerGenerate, inProgress, hookError] = useGenerateRecoveryCodes();
    const [confirmInput, setConfirmInput] = useState("");
    const triggeredRef = useRef(false);

    useEffect(() => {
        if (props.open && !triggeredRef.current) {
            triggeredRef.current = true;
            triggerGenerate();
        }
    }, [props.open, triggerGenerate]);

    const codes = response?.codes;
    const notificationSent = response?.notification_sent ?? true;
    const error = hookError ? translate("Failed to generate recovery codes") : undefined;
    const loading = inProgress;

    // Stamp the moment the generation response first becomes available; recomputes only when response identity changes.
    const generatedAt = useMemo(() => (response ? new Date() : undefined), [response]);

    const reset = useCallback(() => {
        setConfirmInput("");
        triggeredRef.current = false;
    }, []);

    const confirmMatches =
        codes !== undefined && confirmInput.length > 0 && codes.some((c) => normalize(c) === normalize(confirmInput));

    const handleClose = useCallback(() => {
        reset();
        props.setClosed();
    }, [props, reset]);

    const handleCopyAll = useCallback(() => {
        if (!codes) return;

        navigator.clipboard.writeText(codes.join("\n")).catch(console.error);
    }, [codes]);

    const handleDownload = useCallback(() => {
        if (!codes || !generatedAt) return;

        const blob = new Blob([buildDownloadBody(codes, generatedAt)], { type: "text/plain" });
        const url = URL.createObjectURL(blob);

        const a = document.createElement("a");
        a.href = url;
        a.download = buildDownloadFilename();
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }, [codes, generatedAt]);

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleClose();
            }}
        >
            <DialogContent showCloseButton={false} aria-labelledby="recovery-codes-generate-dialog-title">
                <DialogHeader>
                    <DialogTitle id="recovery-codes-generate-dialog-title">
                        {translate("Save your recovery codes")}
                    </DialogTitle>
                </DialogHeader>
                {loading ? (
                    <div className="flex justify-center p-8">
                        <Spinner size={32} />
                    </div>
                ) : null}
                {error ? (
                    <Alert variant="destructive">
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                ) : null}
                {codes ? (
                    <Fragment>
                        <DialogDescription>
                            {translate(
                                "These codes can each be used once if you lose access to your second factor device. Save them now; they will not be shown again.",
                            )}
                        </DialogDescription>
                        {!notificationSent ? (
                            <Alert variant="warning">
                                <AlertDescription>
                                    {translate(
                                        "We could not email you a confirmation; if you did not initiate this, change your password immediately.",
                                    )}
                                </AlertDescription>
                            </Alert>
                        ) : null}
                        <div className="grid grid-cols-1 gap-1 rounded-md bg-muted p-4 font-mono">
                            {codes.map((c) => (
                                <div key={c} className="flex items-center justify-between">
                                    <span className="font-mono">{c}</span>
                                    <CopyButton variant="ghost" tooltip={translate("Copy")} value={c}>
                                        {translate("Copy")}
                                    </CopyButton>
                                </div>
                            ))}
                        </div>
                        <div className="flex flex-row gap-2">
                            <Button id={"recovery-codes-copy-all"} variant="outline" onClick={handleCopyAll}>
                                {translate("Copy all")}
                            </Button>
                            <Button id={"recovery-codes-download"} variant="outline" onClick={handleDownload}>
                                {translate("Download")}
                            </Button>
                        </div>
                        <p className="text-xs text-muted-foreground">
                            {translate(
                                "Downloads can be picked up by cloud sync (OneDrive, iCloud), shared clipboards, and chat auto-paste. Prefer printing or pasting into a password manager.",
                            )}
                        </p>
                        <div className="w-full">
                            <Label htmlFor="recovery-codes-confirm-input">
                                {translate("Type one of the codes above to confirm you have saved them")}
                            </Label>
                            <Input
                                id="recovery-codes-confirm-input"
                                value={confirmInput}
                                onChange={(e) => setConfirmInput(e.target.value)}
                                autoComplete="off"
                            />
                        </div>
                    </Fragment>
                ) : null}
                <DialogFooter>
                    <Button
                        id={"recovery-codes-done"}
                        variant={"ghost"}
                        color={"primary"}
                        onClick={handleClose}
                        disabled={!error && (!codes || !confirmMatches)}
                    >
                        {translate("Done")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default RecoveryCodesGenerateDialog;
