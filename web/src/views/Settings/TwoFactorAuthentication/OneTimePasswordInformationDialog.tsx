import { Fragment } from "react";

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
import { Separator } from "@components/UI/Separator";
import { FormatDateHumanReadable } from "@i18n/formats";
import { UserInfoTOTPConfiguration, toAlgorithmString } from "@models/TOTPConfiguration";

interface Props {
    config: null | undefined | UserInfoTOTPConfiguration;
    open: boolean;
    handleClose: () => void;
}

const OneTimePasswordInformationDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) props.handleClose();
            }}
        >
            <DialogContent showCloseButton={false} aria-labelledby="one-time-password-info-dialog-title">
                <DialogHeader>
                    <DialogTitle id="one-time-password-info-dialog-title">
                        {translate("One-Time Password Information")}
                    </DialogTitle>
                </DialogHeader>
                {props.config ? (
                    <Fragment>
                        <DialogDescription className="mb-6">
                            {translate("Extended information for One-Time Password")}
                        </DialogDescription>
                        <div className="grid grid-cols-12 gap-4">
                            <div className="col-span-12">
                                <Separator />
                            </div>
                            <PropertyText
                                name={translate("Algorithm")}
                                value={translate("{{algorithm}}", {
                                    algorithm: toAlgorithmString(props.config.algorithm),
                                })}
                            />
                            <PropertyText
                                name={translate("Digits")}
                                value={translate("{{digits}}", {
                                    digits: props.config.digits,
                                })}
                            />
                            <PropertyText
                                name={translate("Period")}
                                value={translate("{{seconds}}", {
                                    seconds: props.config.period,
                                })}
                            />
                            <PropertyText name={translate("Issuer")} value={props.config.issuer} />
                            <PropertyText
                                name={translate("Added")}
                                value={translate("{{when, datetime}}", {
                                    formatParams: { when: FormatDateHumanReadable },
                                    when: new Date(props.config.created_at),
                                })}
                            />
                            <PropertyText
                                name={translate("Last Used")}
                                value={
                                    props.config.last_used_at
                                        ? translate("{{when, datetime}}", {
                                              formatParams: { when: FormatDateHumanReadable },
                                              when: new Date(props.config.last_used_at),
                                          })
                                        : translate("Never")
                                }
                            />
                        </div>
                    </Fragment>
                ) : (
                    <DialogDescription className="mb-6">
                        {translate("The One-Time Password information is not loaded")}
                    </DialogDescription>
                )}
                <DialogFooter>
                    <Button id={"dialog-close"} variant={"outline"} onClick={props.handleClose}>
                        {translate("Close")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

interface PropertyTextProps {
    readonly name: string;
    readonly value: string;
    readonly xs?: number;
}

function PropertyText(props: PropertyTextProps) {
    return (
        <div className="col-span-12">
            <span className="font-bold">{`${props.name}: `}</span>
            <span>{props.value}</span>
        </div>
    );
}

export default OneTimePasswordInformationDialog;
