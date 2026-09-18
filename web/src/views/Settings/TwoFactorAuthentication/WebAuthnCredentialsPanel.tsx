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
import { UserInfo } from "@models/UserInfo";
import { WebAuthnCredential } from "@models/WebAuthn";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import ReauthenticationDialog from "@views/Settings/Common/ReauthenticationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import WebAuthnCredentialDeleteDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialDeleteDialog";
import WebAuthnCredentialEditDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialEditDialog";
import WebAuthnCredentialInformationDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialInformationDialog";
import WebAuthnCredentialRegisterDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialRegisterDialog";
import WebAuthnCredentialsGrid from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsGrid";

interface Props {
    info?: UserInfo;
    credentials: undefined | WebAuthnCredential[];
    handleRefreshState: () => void;
}

type Action = "delete" | "edit" | "register";

const WebAuthnCredentialsPanel = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [dialogRegisterOpen, setDialogRegisterOpen] = useState(false);

    const [dialogInformationOpen, setDialogInformationOpen] = useState(false);
    const [indexInformation, setIndexInformation] = useState(-1);

    const [dialogEditOpen, setDialogEditOpen] = useState(false);
    const [indexEdit, setIndexEdit] = useState(-1);

    const [dialogDeleteOpen, setDialogDeleteOpen] = useState(false);
    const [indexDelete, setIndexDelete] = useState(-1);

    const handleElevated = useCallback((action: Action) => {
        switch (action) {
            case "register":
                setDialogRegisterOpen(true);
                break;
            case "edit":
                setDialogEditOpen(true);
                break;
            case "delete":
                setDialogDeleteOpen(true);
                break;
        }
    }, []);

    const handleCancelled = useCallback(() => {
        setDialogRegisterOpen(false);
        setDialogEditOpen(false);
        setIndexEdit(-1);
        setDialogDeleteOpen(false);
        setIndexDelete(-1);
    }, []);

    const {
        identityVerificationDialogProps,
        pending,
        reauthenticationDialogProps,
        reset,
        secondFactorDialogProps,
        start,
    } = useElevationFlow<Action>({ onCancelled: handleCancelled, onElevated: handleElevated });

    const handleResetState = useCallback(() => {
        reset();
        handleCancelled();
    }, [handleCancelled, reset]);

    const dialogRegisterOpening = pending === "register";

    const handleRegister = () => {
        start("register");
    };

    const handleInformation = (index: number) => {
        if (!props.credentials) return;

        if (props.credentials.length + 1 < index) return;

        setIndexInformation(index);
        setDialogInformationOpen(true);
    };

    const handleEdit = (index: number) => {
        if (!props.credentials) return;

        if (props.credentials.length + 1 < index) return;

        setIndexEdit(index);
        start("edit");
    };

    const handleDelete = (index: number) => {
        if (!props.credentials) return;

        if (props.credentials.length + 1 < index) return;

        setIndexDelete(index);
        start("delete");
    };

    return (
        <Fragment>
            <ReauthenticationDialog info={props.info} {...reauthenticationDialogProps} />
            <SecondFactorDialog info={props.info} {...secondFactorDialogProps} />
            <IdentityVerificationDialog {...identityVerificationDialogProps} />
            <WebAuthnCredentialRegisterDialog
                open={dialogRegisterOpen}
                setClosed={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <WebAuthnCredentialInformationDialog
                credential={
                    indexInformation === -1 || !props.credentials ? undefined : props.credentials[indexInformation]
                }
                open={dialogInformationOpen}
                handleClose={() => {
                    setDialogInformationOpen(false);
                }}
            />
            <WebAuthnCredentialEditDialog
                credential={indexEdit === -1 || !props.credentials ? undefined : props.credentials[indexEdit]}
                open={dialogEditOpen}
                handleClose={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <WebAuthnCredentialDeleteDialog
                open={dialogDeleteOpen}
                credential={indexDelete === -1 || !props.credentials ? undefined : props.credentials[indexDelete]}
                handleClose={() => {
                    handleResetState();
                    props.handleRefreshState();
                }}
            />
            <Card id={"webauthn-credentials-panel"} data-loading={props.credentials === undefined ? "true" : "false"}>
                <CardContent className="grid grid-cols-12 gap-4 p-4">
                    <div className="col-span-12">
                        <h5 className="text-xl font-semibold">{translate("WebAuthn Credentials")}</h5>
                    </div>
                    <div className="col-span-12 md:col-span-2 xs:col-span-4">
                        <TooltipProvider>
                            <Tooltip>
                                <TooltipTrigger
                                    render={
                                        <Button
                                            id={"webauthn-credential-add"}
                                            variant={"outline"}
                                            color={"primary"}
                                            onClick={handleRegister}
                                            disabled={dialogRegisterOpening || dialogRegisterOpen}
                                        >
                                            {dialogRegisterOpening ? <Spinner size={20} /> : null}
                                            {translate("Add")}
                                        </Button>
                                    }
                                />
                                <TooltipContent>
                                    {translate("Click to add a {{item}} to your account", {
                                        item: translate("WebAuthn Credential"),
                                    })}
                                </TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    </div>
                    <div className="col-span-12">
                        {props.credentials === undefined ? (
                            <Spinner size={20} />
                        ) : props.credentials.length === 0 ? (
                            <p className="text-sm text-muted-foreground">
                                {translate(
                                    "No WebAuthn Credentials have been registered if you'd like to register one click add",
                                )}
                            </p>
                        ) : (
                            <WebAuthnCredentialsGrid
                                credentials={props.credentials}
                                handleInformation={handleInformation}
                                handleEdit={handleEdit}
                                handleDelete={handleDelete}
                            />
                        )}
                    </div>
                </CardContent>
            </Card>
        </Fragment>
    );
};

export default WebAuthnCredentialsPanel;
