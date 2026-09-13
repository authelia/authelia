// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import { signInWithRecoveryCode } from "@services/UserRecoveryCodes";

export interface Props {
    id: string;
    onSignInError: (_err: Error) => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
    onCancel: () => void;
}

const RecoveryCodeMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();

    const [code, setCode] = useState("");
    const [submitting, setSubmitting] = useState(false);

    const handleSubmit = useCallback(async () => {
        if (!code) return;

        setSubmitting(true);
        try {
            await signInWithRecoveryCode({
                code,
                flow,
                flowID,
                subflow,
                targetURL: redirectionURL ?? undefined,
                userCode,
            });

            props.onSignInSuccess(redirectionURL ?? undefined);
        } catch (err) {
            console.error(err);
            props.onSignInError(new Error(translate("The recovery code might be wrong")));
        } finally {
            setSubmitting(false);
        }
    }, [code, redirectionURL, flowID, flow, subflow, userCode, props, translate]);

    return (
        <div id={props.id} className="mt-4 grid grid-cols-1 gap-4">
            <div className="w-full">
                <Label htmlFor="recovery-code-input">{translate("Recovery code")}</Label>
                <Input
                    id="recovery-code-input"
                    autoFocus
                    placeholder="XXXXX-XXXXX"
                    value={code}
                    onChange={(e) => setCode(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === "Enter" && !submitting) {
                            handleSubmit().catch(console.error);
                        }
                    }}
                    autoComplete="off"
                    spellCheck={false}
                />
            </div>
            <div className="w-full">
                <Button
                    id="recovery-code-submit"
                    variant="default"
                    className="w-full"
                    onClick={() => handleSubmit().catch(console.error)}
                    disabled={!code || submitting}
                >
                    {translate("Sign in")}
                    {submitting ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                </Button>
            </div>
            <div className="w-full">
                <Button
                    id="recovery-code-cancel"
                    variant="ghost"
                    className="w-full"
                    onClick={props.onCancel}
                    disabled={submitting}
                >
                    {translate("Back to second factor")}
                </Button>
            </div>
        </div>
    );
};

export default RecoveryCodeMethod;
