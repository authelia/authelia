import { KeyboardEvent, Ref, useCallback, useId, useState } from "react";

import { AlertCircle, Eye, EyeOff } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Alert, AlertTitle } from "@components/UI/Alert";
import { Field, FieldError, FieldLabel } from "@components/UI/Field";
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "@components/UI/InputGroup";
import { IsCapsLockModified } from "@services/CapsLock";

export interface Props {
    disabled: boolean;
    error: boolean;
    failure?: null | string;
    onChange: (_value: string) => void;
    ref?: Ref<HTMLInputElement>;
    value: string;
}

function DecisionFormReauthentication({ disabled, error, failure, onChange, ref, value }: Props) {
    const { t: translate } = useTranslation(["consent", "portal"]);

    const failureId = useId();

    const [showPassword, setShowPassword] = useState(false);
    const [hasCapsLock, setHasCapsLock] = useState(false);
    const [isCapsLockPartial, setIsCapsLockPartial] = useState(false);

    const handleKeyUp = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            if (value.length <= 1) {
                setHasCapsLock(false);
                setIsCapsLockPartial(false);

                if (value.length === 0) {
                    return;
                }
            }

            const modified = IsCapsLockModified(event);

            if (modified === null) return;

            if (modified) {
                setHasCapsLock(true);
            } else {
                setIsCapsLockPartial(true);
            }
        },
        [value.length],
    );

    const handleToggleKeyDown = useCallback((event: KeyboardEvent<HTMLButtonElement>) => {
        if (event.key === " " || event.key === "Enter") {
            setShowPassword(true);
            event.preventDefault();
        }
    }, []);

    const handleToggleKeyUp = useCallback((event: KeyboardEvent<HTMLButtonElement>) => {
        if (event.key === " " || event.key === "Enter") {
            setShowPassword(false);
            event.preventDefault();
        }
    }, []);

    const invalid = error || !!failure;
    const showCapsLock = hasCapsLock && value.length !== 0;

    return (
        <Field id={"openid-consent-prompt-login"} data-invalid={invalid || undefined} className="text-left">
            {failure ? (
                <FieldError
                    id={failureId}
                    className="flex items-center gap-3 rounded-md border border-destructive px-4 py-3 font-medium"
                >
                    <AlertCircle className="size-4 shrink-0" />
                    {failure}
                </FieldError>
            ) : null}
            <FieldLabel htmlFor="password-textfield">{translate("Password", { ns: "portal" })}</FieldLabel>
            <InputGroup>
                <InputGroupInput
                    id={"password-textfield"}
                    name={"password"}
                    ref={ref}
                    error={invalid}
                    disabled={disabled}
                    value={value}
                    onChange={(event) => onChange(event.target.value)}
                    onKeyUp={handleKeyUp}
                    type={showPassword ? "text" : "password"}
                    autoComplete={"current-password"}
                    aria-required={"true"}
                    aria-describedby={failure ? failureId : undefined}
                />
                <InputGroupAddon align={"inline-end"}>
                    <InputGroupButton
                        size={"icon-sm"}
                        tabIndex={-1}
                        aria-label={translate("Toggle password visibility", { ns: "portal" })}
                        aria-pressed={showPassword}
                        onMouseDown={() => setShowPassword(true)}
                        onMouseUp={() => setShowPassword(false)}
                        onMouseLeave={() => setShowPassword(false)}
                        onTouchStart={() => setShowPassword(true)}
                        onTouchEnd={() => setShowPassword(false)}
                        onTouchCancel={() => setShowPassword(false)}
                        onKeyDown={handleToggleKeyDown}
                        onKeyUp={handleToggleKeyUp}
                    >
                        {showPassword ? <Eye /> : <EyeOff />}
                    </InputGroupButton>
                </InputGroupAddon>
            </InputGroup>
            {showCapsLock ? (
                <Alert variant="default">
                    <AlertTitle>{translate("Warning", { ns: "portal" })}</AlertTitle>
                    {isCapsLockPartial
                        ? translate("The password was partially entered with Caps Lock", { ns: "portal" })
                        : translate("The password was entered with Caps Lock", { ns: "portal" })}
                </Alert>
            ) : null}
        </Field>
    );
}

export default DecisionFormReauthentication;
