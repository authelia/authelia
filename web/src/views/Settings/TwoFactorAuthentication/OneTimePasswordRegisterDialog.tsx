import { Fragment, ReactNode, useCallback, useEffect, useState } from "react";

import { XCircle } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { useTranslation } from "react-i18next";

import AppStoreBadges from "@components/AppStoreBadges";
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
import { RadioGroup, RadioGroupItem } from "@components/UI/RadioGroup";
import { Step, StepLabel, Stepper } from "@components/UI/Stepper";
import { Switch } from "@components/UI/Switch";
import { GoogleAuthenticator } from "@constants/constants";
import { useNotifications } from "@hooks/NotificationsContext";
import { toAlgorithmString } from "@models/TOTPConfiguration";
import { completeTOTPRegister, stopTOTPRegister } from "@services/OneTimePassword";
import { getTOTPSecret } from "@services/RegisterDevice";
import { getTOTPOptions } from "@services/UserInfoTOTPConfiguration";
import { cn } from "@utils/Styles";
import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

const steps = ["Start", "Register", "Confirm"];

interface Props {
    open: boolean;
    setClosed: () => void;
}

interface Options {
    algorithm: string;
    length: number;
    period: number;
}

interface AvailableOptions {
    algorithms: string[];
    lengths: number[];
    periods: number[];
}

const OneTimePasswordRegisterDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [selected, setSelected] = useState<Options>({ algorithm: "", length: 6, period: 30 });
    const [defaults, setDefaults] = useState<null | Options>(null);
    const [available, setAvailable] = useState<AvailableOptions>({
        algorithms: [],
        lengths: [],
        periods: [],
    });

    const [activeStep, setActiveStep] = useState(0);

    const [secretURL, setSecretURL] = useState<null | string>(null);
    const [secretValue, setSecretValue] = useState<null | string>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [showAdvanced, setShowAdvanced] = useState(false);
    const [hasErrored, setHasErrored] = useState(false);
    const [dialValue, setDialValue] = useState("");
    const [dialState, setDialState] = useState(State.Idle);
    const [showQRCode, setShowQRCode] = useState(true);
    const [success, setSuccess] = useState(false);

    const resetStates = useCallback(() => {
        if (defaults) {
            setSelected(defaults);
        }

        setSecretURL(null);
        setSecretValue(null);
        setIsLoading(false);
        setShowAdvanced(false);
        setHasErrored(false);
        setActiveStep(0);
        setDialValue("");
        setDialState(State.Idle);
        setShowQRCode(true);
        setSuccess(false);
    }, [defaults]);

    const handleClose = useCallback(() => {
        (async () => {
            props.setClosed();

            if (secretURL) {
                try {
                    await stopTOTPRegister();
                } catch (err) {
                    console.error(err);
                }
            }

            resetStates();
        })();
    }, [props, secretURL, resetStates]);

    const handleFinished = useCallback(() => {
        setSuccess(true);

        setTimeout(() => {
            createSuccessNotification(
                translate("Successfully {{action}} the {{item}}", {
                    action: translate("added"),
                    item: translate("One-Time Password"),
                }),
            );

            props.setClosed();
            resetStates();
        }, 750);
    }, [createSuccessNotification, props, resetStates, translate]);

    const handleOnClose = () => {
        if (!props.open) {
            return;
        }

        handleClose();
    };

    useEffect(() => {
        if (!props.open || activeStep !== 0 || defaults !== null) {
            return;
        }

        (async () => {
            const opts = await getTOTPOptions();

            const decoded = {
                algorithm: toAlgorithmString(opts.algorithm),
                length: opts.length,
                period: opts.period,
            };

            setAvailable({
                algorithms: opts.algorithms.map((algorithm) => toAlgorithmString(algorithm)),
                lengths: opts.lengths,
                periods: opts.periods,
            });

            setDefaults(decoded);
            setSelected(decoded);
        })();
    }, [props.open, activeStep, defaults, selected]);

    const handleSetStepPrevious = useCallback(() => {
        if (activeStep === 0) {
            return;
        }

        setShowAdvanced(false);
        setActiveStep((prevState) => {
            return prevState - 1;
        });
    }, [activeStep]);

    const handleSetStepNext = useCallback(() => {
        if (activeStep === steps.length - 1) {
            return;
        }

        setShowAdvanced(false);
        setActiveStep((prevState) => {
            return prevState + 1;
        });
    }, [activeStep]);

    useEffect(() => {
        if (!props.open || activeStep !== 1) {
            return;
        }

        (async () => {
            setIsLoading(true);

            try {
                const secret = await getTOTPSecret(selected.algorithm, selected.length, selected.period);
                setSecretURL(secret.otpauth_url);
                setSecretValue(secret.base32_secret);
            } catch (err) {
                console.error(err);
                if ((err as Error).message.includes("Request failed with status code 403")) {
                    createErrorNotification(
                        translate("You must use the code from the same device and browser that initiated the process"),
                    );
                } else {
                    createErrorNotification(
                        translate("Failed to register device, the provided code is expired or has already been used"),
                    );
                }
                setHasErrored(true);
            }

            setIsLoading(false);
        })();
    }, [activeStep, createErrorNotification, selected, props.open, translate]);

    useEffect(() => {
        if (!props.open || activeStep !== 2 || dialState === State.InProgress || dialValue.length !== selected.length) {
            return;
        }

        (async () => {
            setDialState(State.InProgress);

            try {
                const registerValue = dialValue;
                setDialValue("");

                await completeTOTPRegister(registerValue);

                handleFinished();
            } catch (err) {
                console.error(err);
                setDialState(State.Failure);
            }
        })();
    }, [activeStep, dialState, dialValue, dialValue.length, handleFinished, props.open, selected.length]);

    const handleChangeAlgorithm = (value: string) => {
        setSelected((prevState) => {
            return {
                ...prevState,
                algorithm: value,
            };
        });
    };

    const handleChangeLength = (value: string) => {
        setSelected((prevState) => {
            return {
                ...prevState,
                length: Number.parseInt(value),
            };
        });
    };

    const handleChangePeriod = (value: string) => {
        setSelected((prevState) => {
            return {
                ...prevState,
                period: Number.parseInt(value),
            };
        });
    };

    const toggleAdvanced = () => {
        setShowAdvanced((prevState) => !prevState);
    };

    const advanced =
        defaults !== null &&
        (available.algorithms.length !== 1 || available.lengths.length !== 1 || available.periods.length !== 1);

    const disableAdvanced =
        defaults === null ||
        (available.algorithms.length <= 1 && available.lengths.length <= 1 && available.periods.length <= 1);

    const hideAlgorithms = advanced && available.algorithms.length <= 1;
    const hideLengths = advanced && available.lengths.length <= 1;
    const hidePeriods = advanced && available.periods.length <= 1;

    function renderStep(step: number) {
        switch (step) {
            case 0:
                return (
                    <Fragment>
                        {defaults === null ? (
                            <div className="col-span-12 my-6">
                                <p>Loading...</p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-12">
                                <div className="col-span-12 my-6">
                                    <p>{translate("To begin select next")}</p>
                                </div>
                                <div className={cn("col-span-12", disableAdvanced && "hidden")}>
                                    <div className="flex items-center gap-2">
                                        <Switch
                                            id={"one-time-password-advanced"}
                                            checked={showAdvanced}
                                            onCheckedChange={toggleAdvanced}
                                            disabled={disableAdvanced}
                                        />
                                        <Label htmlFor={"one-time-password-advanced"}>{translate("Advanced")}</Label>
                                    </div>
                                </div>
                                <div
                                    className={cn(
                                        "col-span-12 flex flex-col items-center justify-center",
                                        (disableAdvanced || !showAdvanced) && "hidden",
                                    )}
                                >
                                    <div className="mt-4 w-full space-y-4">
                                        {hideAlgorithms ? null : (
                                            <div className="space-y-2">
                                                <Label id={"lbl-adv-algorithms"}>{translate("Algorithm")}</Label>
                                                <RadioGroup
                                                    aria-labelledby={"lbl-adv-algorithms"}
                                                    value={selected.algorithm}
                                                    className="flex flex-row gap-4"
                                                    onValueChange={handleChangeAlgorithm}
                                                >
                                                    {available.algorithms.map((algorithm) => (
                                                        <div key={algorithm} className="flex items-center gap-2">
                                                            <RadioGroupItem
                                                                id={`one-time-password-algorithm-${algorithm}`}
                                                                value={algorithm}
                                                            />
                                                            <Label htmlFor={`one-time-password-algorithm-${algorithm}`}>
                                                                {algorithm}
                                                            </Label>
                                                        </div>
                                                    ))}
                                                </RadioGroup>
                                            </div>
                                        )}
                                        {hideLengths ? null : (
                                            <div className="space-y-2">
                                                <Label id={"lbl-adv-lengths"}>{translate("Length")}</Label>
                                                <RadioGroup
                                                    aria-labelledby={"lbl-adv-lengths"}
                                                    value={selected.length.toString()}
                                                    className="flex flex-row gap-4"
                                                    onValueChange={handleChangeLength}
                                                >
                                                    {available.lengths.map((length) => (
                                                        <div
                                                            key={length.toString()}
                                                            className="flex items-center gap-2"
                                                        >
                                                            <RadioGroupItem
                                                                id={`one-time-password-length-${length.toString()}`}
                                                                value={length.toString()}
                                                            />
                                                            <Label
                                                                htmlFor={`one-time-password-length-${length.toString()}`}
                                                            >
                                                                {length.toString()}
                                                            </Label>
                                                        </div>
                                                    ))}
                                                </RadioGroup>
                                            </div>
                                        )}
                                        {hidePeriods ? null : (
                                            <div className="space-y-2">
                                                <Label id={"lbl-adv-periods"}>{translate("Seconds")}</Label>
                                                <RadioGroup
                                                    aria-labelledby={"lbl-adv-periods"}
                                                    value={selected.period.toString()}
                                                    className="flex flex-row gap-4"
                                                    onValueChange={handleChangePeriod}
                                                >
                                                    {available.periods.map((period) => (
                                                        <div
                                                            key={period.toString()}
                                                            className="flex items-center gap-2"
                                                        >
                                                            <RadioGroupItem
                                                                id={`one-time-password-period-${period.toString()}`}
                                                                value={period.toString()}
                                                            />
                                                            <Label
                                                                htmlFor={`one-time-password-period-${period.toString()}`}
                                                            >
                                                                {period.toString()}
                                                            </Label>
                                                        </div>
                                                    ))}
                                                </RadioGroup>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        )}
                    </Fragment>
                );
            case 1:
                return (
                    <Fragment>
                        <div className="col-span-12 my-4">
                            <div className="flex items-center gap-2">
                                <Switch
                                    id={"qr-toggle"}
                                    checked={showQRCode}
                                    onCheckedChange={(checked) => {
                                        setShowQRCode(checked);
                                    }}
                                />
                                <Label htmlFor={"qr-toggle"}>{translate("QR Code")}</Label>
                            </div>
                        </div>
                        <div className={cn("col-span-12", !showQRCode && "hidden")}>
                            <div className={cn("inline-block relative", (isLoading || hasErrored) && "blur-sm")}>
                                {secretURL ? (
                                    <a href={secretURL} className="hover:opacity-80">
                                        <QRCodeSVG
                                            value={secretURL}
                                            size={200}
                                            className="inline-block my-4 p-2 bg-white"
                                        />
                                        {isLoading && !hasErrored ? (
                                            <div
                                                className="absolute"
                                                style={{
                                                    color: "rgba(255, 255, 255, 0.5)",
                                                    left: "calc(128px - 64px)",
                                                    top: "calc(128px - 64px)",
                                                }}
                                            >
                                                <div className="size-32 animate-spin rounded-full border-4 border-current border-t-transparent" />
                                            </div>
                                        ) : null}
                                        {hasErrored ? (
                                            <XCircle
                                                className="absolute text-red-400"
                                                style={{
                                                    fontSize: "8rem",
                                                    height: "128px",
                                                    left: "calc(128px - 64px)",
                                                    top: "calc(128px - 64px)",
                                                    width: "128px",
                                                }}
                                            />
                                        ) : null}
                                    </a>
                                ) : null}
                            </div>
                        </div>
                        <div className={cn("col-span-12", showQRCode && "hidden")}>
                            <div className="flex flex-col items-center gap-4">
                                <div className="grid w-64 grid-cols-2 gap-4">
                                    <CopyButton
                                        tooltip={translate("Click to Copy")}
                                        value={secretURL}
                                        childrenCopied={translate("Copied")}
                                        fullWidth={true}
                                    >
                                        {translate("URI")}
                                    </CopyButton>
                                    <CopyButton
                                        tooltip={translate("Click to Copy")}
                                        value={secretValue}
                                        childrenCopied={translate("Copied")}
                                        fullWidth={true}
                                    >
                                        {translate("Secret")}
                                    </CopyButton>
                                </div>
                                <div className="flex flex-col items-start">
                                    <Label htmlFor="secret-url">{translate("Secret")}</Label>
                                    <textarea
                                        id={"secret-url"}
                                        className="my-2 w-64 resize-none overflow-hidden rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs outline-none"
                                        value={secretURL ?? ""}
                                        readOnly
                                        rows={secretURL ? Math.ceil(secretURL.length / 30) : 1}
                                    />
                                </div>
                            </div>
                        </div>
                        <div className="col-span-12 hidden md:block">
                            <div className="text-center">
                                <p className="text-xs">{translate("Need Google Authenticator?")}</p>
                                <AppStoreBadges
                                    iconSize={110}
                                    targetBlank
                                    googlePlayLink={GoogleAuthenticator.googlePlay}
                                    appleStoreLink={GoogleAuthenticator.appleStore}
                                />
                            </div>
                        </div>
                    </Fragment>
                );
            case 2:
                return (
                    <div className="col-span-12 py-8">
                        {success ? (
                            <div className="flex-[0_0_100%] mb-4">
                                <SuccessIcon />
                            </div>
                        ) : (
                            <OTPDial
                                passcode={dialValue}
                                state={dialState}
                                digits={selected.length}
                                period={selected.period}
                                onChange={setDialValue}
                            />
                        )}
                    </div>
                );
        }
    }

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleOnClose();
            }}
        >
            <DialogContent showCloseButton={false} className="w-full">
                <DialogHeader>
                    <DialogTitle>
                        {translate("Register {{item}}", { item: translate("One-Time Password") })}
                    </DialogTitle>
                    <DialogDescription className="mb-6">
                        {translate("This dialog handles registration of a {{item}}", {
                            item: translate("One-Time Password"),
                        })}
                    </DialogDescription>
                </DialogHeader>
                <div className="flex flex-col items-center justify-center text-center">
                    <div className="w-full px-6">
                        <Stepper activeStep={activeStep}>
                            {steps.map((label) => {
                                const stepProps: { completed?: boolean } = {};
                                const labelProps: {
                                    optional?: ReactNode;
                                } = {};
                                return (
                                    <Step key={label} {...stepProps}>
                                        <StepLabel {...labelProps}>{translate(label)}</StepLabel>
                                    </Step>
                                );
                            })}
                        </Stepper>
                    </div>
                    <div className="w-full">
                        <div className="flex flex-col items-center justify-center gap-2">{renderStep(activeStep)}</div>
                    </div>
                </div>
                <DialogFooter>
                    <Button
                        id={"dialog-previous"}
                        variant={"outline"}
                        onClick={handleSetStepPrevious}
                        disabled={activeStep === 0}
                    >
                        {translate("Previous")}
                    </Button>
                    <Button id={"dialog-cancel"} variant={"destructive"} onClick={handleClose}>
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id={"dialog-next"}
                        variant={"outline"}
                        onClick={handleSetStepNext}
                        disabled={activeStep === steps.length - 1}
                    >
                        {translate("Next")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default OneTimePasswordRegisterDialog;
