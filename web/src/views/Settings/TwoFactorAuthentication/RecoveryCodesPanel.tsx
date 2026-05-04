// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Alert, AlertDescription } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { Card, CardContent } from "@components/UI/Card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/UI/Dialog";
import { Spinner } from "@components/UI/Spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { useRecoveryCodesStatus } from "@hooks/UserRecoveryCodes";
import { UserInfo } from "@models/UserInfo";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import RecoveryCodesGenerateDialog from "@views/Settings/TwoFactorAuthentication/RecoveryCodesGenerateDialog";

interface Props {
    info?: UserInfo;
}

const RecoveryCodesPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [status, fetchStatus, loading] = useRecoveryCodesStatus();

    const [elevation, setElevation] = useState<UserSessionElevation>();
    const [dialogSFOpening, setDialogSFOpening] = useState(false);
    const [dialogIVOpening, setDialogIVOpening] = useState(false);
    const [dialogConfirmOpen, setDialogConfirmOpen] = useState(false);
    const [dialogGenerateOpen, setDialogGenerateOpen] = useState(false);
    const [dialogGenerateOpening, setDialogGenerateOpening] = useState(false);

    useEffect(() => {
        fetchStatus();
    }, [fetchStatus]);

    const resetDialogs = useCallback(() => {
        setDialogSFOpening(false);
        setDialogIVOpening(false);
        setDialogGenerateOpening(false);
        setDialogConfirmOpen(false);
        setDialogGenerateOpen(false);
        setElevation(undefined);
    }, []);

    const handleElevationRefresh = useCallback(async () => {
        const result = await getUserSessionElevation();
        setElevation(result);

        return result;
    }, []);

    const openGenerate = useCallback(() => {
        setDialogGenerateOpen(true);
    }, []);

    const handleSFDialogClosed = (ok: boolean, changed: boolean) => {
        if (!ok) {
            resetDialogs();

            return;
        }

        if (changed) {
            handleElevationRefresh()
                .catch(console.error)
                .then((refreshed) => {
                    if (!refreshed) return;

                    if (refreshed.elevated || refreshed.skip_second_factor) {
                        setElevation(undefined);
                        openGenerate();
                    } else {
                        setDialogIVOpening(true);
                    }
                });
        } else {
            const isElevated = elevation && (elevation.elevated || elevation.skip_second_factor);

            if (isElevated) {
                setElevation(undefined);
                openGenerate();
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
                resetDialogs();

                return;
            }

            setElevation(undefined);
            openGenerate();
        },
        [resetDialogs, openGenerate],
    );

    const handleIVDialogOpened = useCallback(() => {
        setDialogIVOpening(false);
    }, []);

    const startGenerationFlow = useCallback(() => {
        setDialogGenerateOpening(true);
        handleElevationRefresh().catch(console.error);
        setDialogSFOpening(true);
    }, [handleElevationRefresh]);

    const handleGenerateClicked = () => {
        if (status && status.codes_total > 0) {
            setDialogConfirmOpen(true);

            return;
        }

        startGenerationFlow();
    };

    const handleConfirmRegenerate = () => {
        setDialogConfirmOpen(false);
        startGenerationFlow();
    };

    const handleGenerateDialogClosed = () => {
        setDialogGenerateOpen(false);
        setDialogGenerateOpening(false);
        fetchStatus();
    };

    const hasCodes = status !== undefined && status.codes_total > 0;
    const codesUsed = status ? status.codes_total - status.codes_remaining : 0;
    const lowCodes = props.info?.low_recovery_codes === true;

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
            <Dialog
                open={dialogConfirmOpen}
                onOpenChange={(open) => {
                    if (!open) setDialogConfirmOpen(false);
                }}
            >
                <DialogContent showCloseButton={false} aria-labelledby="recovery-codes-confirm-dialog-title">
                    <DialogHeader>
                        <DialogTitle id="recovery-codes-confirm-dialog-title">
                            {translate("Regenerate recovery codes?")}
                        </DialogTitle>
                    </DialogHeader>
                    <DialogDescription>
                        {translate(
                            "Regenerating will invalidate all of your existing recovery codes. Make sure you save the new codes; they will only be shown once.",
                        )}
                    </DialogDescription>
                    <DialogFooter>
                        <Button
                            id={"recovery-codes-confirm-cancel"}
                            variant={"ghost"}
                            onClick={() => setDialogConfirmOpen(false)}
                        >
                            {translate("Cancel")}
                        </Button>
                        <Button
                            id={"recovery-codes-confirm-accept"}
                            color={"primary"}
                            onClick={handleConfirmRegenerate}
                        >
                            {translate("Regenerate")}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
            <RecoveryCodesGenerateDialog open={dialogGenerateOpen} setClosed={handleGenerateDialogClosed} />
            <Card id={"recovery-codes-panel"} data-loading={loading ? "true" : "false"}>
                <CardContent className="grid grid-cols-12 gap-4 p-4">
                    <div className="col-span-12">
                        <h5 className="text-xl font-semibold">{translate("Recovery Codes")}</h5>
                        <p className="text-sm text-muted-foreground">
                            {translate(
                                "Single-use codes you keep somewhere safe and can enter at sign-in if you lose your second factor device.",
                            )}
                        </p>
                    </div>
                    {lowCodes ? (
                        <div className="col-span-12">
                            <Alert variant="warning">
                                <AlertDescription>
                                    {translate(
                                        "Recovery codes are running low. Regenerate to get a fresh batch before you run out.",
                                    )}
                                </AlertDescription>
                            </Alert>
                        </div>
                    ) : null}
                    <div className="col-span-12">
                        <TooltipProvider>
                            <Tooltip>
                                <TooltipTrigger
                                    render={
                                        <span>
                                            <Button
                                                id={"recovery-codes-add"}
                                                variant={"outline"}
                                                color={"primary"}
                                                onClick={handleGenerateClicked}
                                                disabled={loading || dialogGenerateOpening || dialogGenerateOpen}
                                            >
                                                {dialogGenerateOpening ? <Spinner size={20} /> : null}
                                                {hasCodes ? translate("Regenerate") : translate("Generate")}
                                            </Button>
                                        </span>
                                    }
                                />
                                <TooltipContent>
                                    {hasCodes
                                        ? translate("Generate a new batch (this invalidates the old codes)")
                                        : translate("Generate recovery codes for your account")}
                                </TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    </div>
                    <div className="col-span-12">
                        {hasCodes ? (
                            <Fragment>
                                <p className="text-sm">
                                    {translate("{{used}} of {{total}} codes used", {
                                        total: status!.codes_total,
                                        used: codesUsed,
                                    })}
                                </p>
                                {status!.generated_at ? (
                                    <div className="text-xs text-muted-foreground">
                                        {translate("Generated {{when}}", {
                                            when: status!.generated_at.toLocaleString(),
                                        })}
                                    </div>
                                ) : null}
                                <div className="text-xs text-muted-foreground">
                                    {status!.last_used_at
                                        ? translate("Last used {{when}}", {
                                              when: status!.last_used_at.toLocaleString(),
                                          })
                                        : translate("Never used")}
                                </div>
                            </Fragment>
                        ) : (
                            <p className="text-sm font-medium">
                                {translate("No recovery codes have been generated yet.")}
                            </p>
                        )}
                    </div>
                </CardContent>
            </Card>
        </Fragment>
    );
};

export default RecoveryCodesPanel;
