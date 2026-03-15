import { Fragment, ReactNode, useCallback, useEffect, useRef, useState } from "react";

import { PublicKeyCredentialCreationOptionsJSON } from "@simplewebauthn/browser";
import { useTranslation } from "react-i18next";

import InformationIcon from "@components/InformationIcon";
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
import { Step, StepLabel, Stepper } from "@components/UI/Stepper";
import WebAuthnRegisterIcon from "@components/WebAuthnRegisterIcon";
import { useNotifications } from "@contexts/NotificationsContext";
import { AttestationResult, AttestationResultFailureString, WebAuthnTouchState } from "@models/WebAuthn";
import {
    finishWebAuthnRegistration,
    getWebAuthnRegistrationOptions,
    startWebAuthnRegistration,
} from "@services/WebAuthn";

const steps = ["Description", "Verification"];

interface Props {
    open: boolean;
    setClosed: () => void;
}

const WebAuthnCredentialRegisterDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [state, setState] = useState(WebAuthnTouchState.WaitTouch);
    const [activeStep, setActiveStep] = useState(0);
    const [options, setOptions] = useState<null | PublicKeyCredentialCreationOptionsJSON>(null);
    const [timeout, setTimeout] = useState<null | number>(null);
    const [description, setDescription] = useState("");
    const [errorDescription, setErrorDescription] = useState(false);

    const nameRef = useRef<HTMLInputElement | null>(null);

    const resetStates = () => {
        setState(WebAuthnTouchState.WaitTouch);
        setOptions(null);
        setActiveStep(0);
        setTimeout(null);
        setDescription("");
        setErrorDescription(false);
    };

    const handleClose = useCallback(() => {
        resetStates();

        props.setClosed();
    }, [props]);

    const performCredentialCreation = useCallback(async () => {
        if (!props.open || options === null) {
            return;
        }

        setTimeout(options.timeout ? options.timeout : null);
        setActiveStep(1);

        try {
            setState(WebAuthnTouchState.WaitTouch);

            const result = await startWebAuthnRegistration(options);

            setTimeout(null);

            if (result.result === AttestationResult.Success) {
                if (result.response == null) {
                    createErrorNotification(
                        "Credential Creation Request succeeded but Registration Response is empty.",
                    );
                    return;
                }

                const response = await finishWebAuthnRegistration(result.response);

                switch (response.status) {
                    case AttestationResult.Success:
                        createSuccessNotification(
                            translate("Successfully {{action}} the {{item}}", {
                                action: translate("added"),
                                item: translate("WebAuthn Credential"),
                            }),
                        );
                        break;
                    case AttestationResult.Failure:
                        createErrorNotification(response.message);
                        break;
                }

                return;
            } else {
                createErrorNotification(translate(AttestationResultFailureString(result.result)));
                setState(WebAuthnTouchState.Failure);
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(
                translate("Failed to register your credential, the identity verification process might have timed out"),
            );
        } finally {
            handleClose();
        }
    }, [props.open, options, createSuccessNotification, translate, createErrorNotification, handleClose]);

    useEffect(() => {
        if (!props.open || state !== WebAuthnTouchState.Failure || activeStep !== 0) {
            return;
        }

        handleClose();
    }, [props, state, activeStep, handleClose]);

    useEffect(() => {
        (async function () {
            if (!props.open || activeStep !== 0 || options === null) {
                return;
            }

            await performCredentialCreation();
        })();
    }, [props.open, activeStep, options, performCredentialCreation]);

    const handleNext = useCallback(() => {
        if (!props.open) {
            return;
        }

        (async function () {
            if (description.length === 0 || description.length > 64) {
                setErrorDescription(true);
                createErrorNotification(
                    translate("The Description must be more than 1 character and less than 64 characters"),
                );

                return;
            }

            const res = await getWebAuthnRegistrationOptions(description);

            switch (res.status) {
                case 200:
                    if (res.options) {
                        setOptions(res.options);
                    } else {
                        createErrorNotification(
                            translate(
                                "Credential Creation Options Request succeeded but Credential Creation Options is empty",
                            ),
                        );
                    }

                    break;
                case 409:
                    setErrorDescription(true);
                    createErrorNotification(translate("A WebAuthn Credential with that Description already exists"));

                    break;
                default:
                    createErrorNotification(
                        translate("Error occurred obtaining the WebAuthn Credential creation options"),
                    );
            }

            await performCredentialCreation();
        })();
    }, [createErrorNotification, description, performCredentialCreation, props.open, translate]);

    const handleCredentialDescription = useCallback(
        (description: string) => {
            setDescription(description);

            if (errorDescription) {
                setErrorDescription(false);
            }
        },
        [errorDescription],
    );

    function renderStep(step: number) {
        switch (step) {
            case 0:
                return (
                    <div>
                        <div className="py-8">
                            <InformationIcon />
                        </div>
                        <p className="pb-8">{translate("Enter a description for this WebAuthn Credential")}</p>
                        <div className="grid grid-cols-12 gap-2">
                            <div className="col-span-12">
                                <Label htmlFor="webauthn-credential-description">{translate("Description")}</Label>
                                <Input
                                    ref={nameRef}
                                    id="webauthn-credential-description"
                                    required
                                    value={description}
                                    aria-invalid={errorDescription}
                                    className={errorDescription ? "border-destructive" : ""}
                                    onChange={(v) => handleCredentialDescription(v.target.value)}
                                    autoCapitalize="none"
                                    onKeyDown={(ev) => {
                                        if (ev.key === "Enter") {
                                            (async () => {
                                                handleNext();
                                            })();

                                            ev.preventDefault();
                                        }
                                    }}
                                />
                            </div>
                        </div>
                    </div>
                );
            case 1:
                return (
                    <Fragment>
                        <div className="py-8">{timeout ? <WebAuthnRegisterIcon timeout={timeout} /> : null}</div>
                        <p className="pb-8">{translate("Touch the token on your security key")}</p>
                    </Fragment>
                );
        }
    }

    const handleOnClose = () => {
        if (!props.open || activeStep === 1) {
            return;
        }

        handleClose();
    };

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleOnClose();
            }}
        >
            <DialogContent showCloseButton={false} className="sm:max-w-xs w-full">
                <DialogHeader>
                    <DialogTitle>
                        {translate("Register {{item}}", { item: translate("WebAuthn Credential") })}
                    </DialogTitle>
                    <DialogDescription className="mb-6">
                        {translate("This dialog handles registration of a {{item}}", {
                            item: translate("WebAuthn Credential"),
                        })}
                    </DialogDescription>
                </DialogHeader>
                <div className="flex flex-col items-center justify-center text-center">
                    <div className="w-full">
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
                    <div className="w-full">{renderStep(activeStep)}</div>
                </div>
                <DialogFooter>
                    <Button
                        id={"dialog-cancel"}
                        variant={activeStep === 1 && state !== WebAuthnTouchState.Failure ? "outline" : "destructive"}
                        disabled={activeStep === 1 && state !== WebAuthnTouchState.Failure}
                        onClick={handleClose}
                    >
                        {translate("Cancel")}
                    </Button>
                    {activeStep === 0 ? (
                        <Button
                            id={"dialog-next"}
                            variant={description.length > 0 ? "default" : "outline"}
                            className={description.length > 0 ? "bg-green-600 hover:bg-green-700 text-white" : ""}
                            disabled={activeStep !== 0}
                            onClick={async () => {
                                handleNext();
                            }}
                        >
                            {translate("Next")}
                        </Button>
                    ) : null}
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default WebAuthnCredentialRegisterDialog;
