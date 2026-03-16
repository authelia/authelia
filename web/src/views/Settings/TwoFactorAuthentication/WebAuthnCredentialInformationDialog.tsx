import { Fragment } from "react";

import { useTranslation } from "react-i18next";

import CopyButton from "@components/CopyButton";
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
import { WebAuthnCredential, toAttachmentName, toTransportName } from "@models/WebAuthn";

interface Props {
    open: boolean;
    credential?: WebAuthnCredential;
    handleClose: () => void;
}

const WebAuthnCredentialInformationDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) props.handleClose();
            }}
        >
            <DialogContent showCloseButton={false} aria-labelledby="webauthn-credential-info-dialog-title">
                <DialogHeader>
                    <DialogTitle id="webauthn-credential-info-dialog-title">
                        {translate("WebAuthn Credential Information")}
                    </DialogTitle>
                </DialogHeader>
                {props.credential ? (
                    <Fragment>
                        <DialogDescription className="mb-6">
                            {translate("Extended information for WebAuthn Credential", {
                                description: props.credential.description,
                            })}
                        </DialogDescription>
                        {props.credential.legacy ? (
                            <div className="mb-6 rounded-md border border-amber-500/50 bg-amber-500/10 p-3 text-sm text-amber-700 dark:text-amber-400">
                                {translate(
                                    "This is a legacy WebAuthn Credential if it's not operating normally you may need to delete it and register it again",
                                )}
                            </div>
                        ) : null}
                        <div className="grid grid-cols-12 gap-4">
                            <div className="col-span-12">
                                <Separator />
                            </div>
                            <PropertyText name={translate("Description")} value={props.credential.description} />
                            <PropertyText name={translate("Relying Party ID")} value={props.credential.rpid} />
                            <PropertyText
                                name={translate("Authenticator GUID")}
                                value={props.credential.aaguid ?? translate("Unknown")}
                            />
                            <PropertyText
                                name={translate("Attestation Type")}
                                value={props.credential.attestation_type}
                            />
                            <PropertyText
                                name={translate("Attachment")}
                                value={translate(toAttachmentName(props.credential.attachment))}
                            />
                            <PropertyText
                                name={translate("Discoverable")}
                                value={props.credential.discoverable ? translate("Yes") : translate("No")}
                            />
                            <PropertyText
                                name={translate("User Verified")}
                                value={props.credential.verified ? translate("Yes") : translate("No")}
                            />
                            {(() => {
                                let backupStateValue: string;

                                if (!props.credential.backup_eligible) {
                                    backupStateValue = translate("Not Eligible");
                                } else if (props.credential.backup_state) {
                                    backupStateValue = translate("Backed Up");
                                } else {
                                    backupStateValue = translate("Eligible");
                                }

                                return <PropertyText name={translate("Backup State")} value={backupStateValue} />;
                            })()}
                            <PropertyText
                                name={translate("Transports")}
                                value={
                                    props.credential.transports === null || props.credential.transports.length === 0
                                        ? translate("Unknown")
                                        : props.credential.transports
                                              .map((transport) => toTransportName(transport))
                                              .join(", ")
                                }
                            />
                            <PropertyText
                                name={translate("Clone Warning")}
                                value={props.credential.clone_warning ? translate("Yes") : translate("No")}
                            />
                            <PropertyText name={translate("Usage Count")} value={`${props.credential.sign_count}`} />
                            <PropertyText
                                name={translate("Added")}
                                value={translate("{{when, datetime}}", {
                                    formatParams: { when: FormatDateHumanReadable },
                                    when: new Date(props.credential.created_at),
                                })}
                            />
                            <PropertyText
                                name={translate("Last Used")}
                                value={
                                    props.credential.last_used_at
                                        ? translate("{{when, datetime}}", {
                                              formatParams: { when: FormatDateHumanReadable },
                                              when: new Date(props.credential.last_used_at),
                                          })
                                        : translate("Never")
                                }
                            />
                        </div>
                    </Fragment>
                ) : (
                    <DialogDescription className="mb-6">
                        {translate("The WebAuthn Credential information is not loaded")}
                    </DialogDescription>
                )}
                <DialogFooter>
                    {props.credential ? (
                        <Fragment>
                            <CopyButton
                                variant={"default"}
                                tooltip={translate("Click to copy the {{value}}", { value: "KID" })}
                                value={props.credential.kid.toString()}
                                fullWidth={false}
                                childrenCopied={translate("Copied")}
                            >
                                {translate("KID")}
                            </CopyButton>
                            <CopyButton
                                variant={"default"}
                                tooltip={translate("Click to copy the {{value}}", { value: translate("Public Key") })}
                                value={props.credential.public_key.toString()}
                                fullWidth={false}
                                childrenCopied={translate("Copied")}
                            >
                                {translate("Public Key")}
                            </CopyButton>
                        </Fragment>
                    ) : undefined}
                    <Button id={"dialog-close"} variant={"ghost"} color={"primary"} onClick={props.handleClose}>
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
}

function PropertyText(props: PropertyTextProps) {
    return (
        <div className="col-span-12">
            <span className="font-bold">{`${props.name}: `}</span>
            <span>{props.value}</span>
        </div>
    );
}

export default WebAuthnCredentialInformationDialog;
