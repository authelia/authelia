// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useRef, useState } from "react";

import { AuthenticationResponseJSON } from "@simplewebauthn/browser";
import axios, { AxiosResponse } from "axios";
import { useTranslation } from "react-i18next";

import { useAbortSignal } from "@hooks/Abort";
import {
    AssertionResult,
    AssertionResultFailureString,
    PublicKeyCredentialRequestOptionsStatus,
    WebAuthnCeremony,
    WebAuthnTouchState,
} from "@models/WebAuthn";
import { ServiceResponse } from "@services/Api";
import { SignInResponse } from "@services/SignIn";
import { cancelWebAuthnCeremony, getWebAuthnResult, isWebAuthnSupported } from "@services/WebAuthn";

export interface WebAuthnAssertionOptions {
    getOptions: (_signal: AbortSignal) => Promise<PublicKeyCredentialRequestOptionsStatus>;
    postResponse: (
        _response: AuthenticationResponseJSON,
        _signal: AbortSignal,
    ) => Promise<AxiosResponse<ServiceResponse<SignInResponse>>>;
    onError: (_err: Error) => void;
    onSuccess: (_redirectURL: string | undefined) => void;
}

export interface WebAuthnAssertion {
    ceremony?: WebAuthnCeremony;
    start: () => Promise<void>;
    state: WebAuthnTouchState;
}

export function useWebAuthnAssertion(options: WebAuthnAssertionOptions): WebAuthnAssertion {
    const { t: translate } = useTranslation();
    const getSignal = useAbortSignal();

    const [state, setState] = useState(WebAuthnTouchState.WaitTouch);
    const [ceremony, setCeremony] = useState<undefined | WebAuthnCeremony>(undefined);

    const optionsRef = useRef(options);
    const translateRef = useRef(translate);
    const activeRef = useRef<AbortSignal | null>(null);

    useEffect(() => {
        optionsRef.current = options;
        translateRef.current = translate;
    }, [options, translate]);

    useEffect(() => {
        return () => {
            cancelWebAuthnCeremony();
        };
    }, []);

    const start = useCallback(async () => {
        if (activeRef.current && !activeRef.current.aborted) {
            return;
        }

        const fail = (message: string) => {
            setState(WebAuthnTouchState.Failure);
            optionsRef.current.onError(new Error(translateRef.current(message)));
        };

        if (!isWebAuthnSupported()) {
            setState(WebAuthnTouchState.Unsupported);
            optionsRef.current.onError(
                new Error(
                    translateRef.current(AssertionResultFailureString(AssertionResult.FailureWebAuthnNotSupported)),
                ),
            );

            return;
        }

        const signal = getSignal();

        activeRef.current = signal;

        try {
            setState(WebAuthnTouchState.WaitTouch);
            setCeremony(undefined);

            const optionsStatus = await optionsRef.current.getOptions(signal);

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                fail("Unable to start sign in with your passkey or security key");

                return;
            }

            setCeremony({ timeout: optionsStatus.options.timeout });

            const result = await getWebAuthnResult(optionsStatus.options);

            if (signal.aborted) return;

            if (result.result === AssertionResult.FailureAborted) {
                setState(WebAuthnTouchState.Failure);

                return;
            }

            if (result.result !== AssertionResult.Success) {
                fail(AssertionResultFailureString(result.result));

                return;
            }

            if (result.response == null) {
                fail("Your browser did not return a response from your passkey or security key");

                return;
            }

            setState(WebAuthnTouchState.InProgress);

            const response = await optionsRef.current.postResponse(result.response, signal);

            if (response.data.status === "OK" && response.status === 200) {
                optionsRef.current.onSuccess(response.data.data ? response.data.data.redirect : undefined);

                return;
            }

            fail("Your passkey or security key could not be verified");
        } catch (err) {
            if (axios.isCancel(err)) return;

            console.error(err);
            fail("Unable to start sign in with your passkey or security key");
        } finally {
            if (activeRef.current === signal) {
                activeRef.current = null;
            }
        }
    }, [getSignal]);

    return { ceremony, start, state };
}
