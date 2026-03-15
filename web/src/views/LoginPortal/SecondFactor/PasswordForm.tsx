import { KeyboardEvent, useCallback, useEffect, useRef, useState } from "react";

import { Eye, EyeOff } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Alert, AlertTitle } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useFlow } from "@hooks/Flow";
import { useQueryParam } from "@hooks/QueryParam";
import { IsCapsLockModified } from "@services/CapsLock";
import { postSecondFactor } from "@services/Password";
import { cn } from "@utils/Styles";

export interface Props {
    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
}

const PasswordForm = function (props: Props) {
    const { createErrorNotification } = useNotifications();
    const { t: translate } = useTranslation(["portal", "settings"]);

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();

    const [loading, setLoading] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [passwordCapsLockPartial, setPasswordCapsLockPartial] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [showPassword, setShowPassword] = useState(false);

    const passwordRef = useRef<HTMLInputElement | null>(null);

    const focusPassword = useCallback(() => {
        if (passwordRef.current === null) return;

        passwordRef.current.focus();
    }, [passwordRef]);

    useEffect(() => {
        const timeout = setTimeout(() => focusPassword(), 10);
        return () => clearTimeout(timeout);
    }, [focusPassword]);

    const handleSignIn = useCallback(async () => {
        if (password === "") {
            setPasswordError(true);

            return;
        }

        setLoading(true);

        try {
            const res = await postSecondFactor(password, redirectionURL, flowID, flow, subflow);
            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Incorrect password"));
            setPassword("");
            setLoading(false);
            focusPassword();
        }
    }, [createErrorNotification, focusPassword, password, props, redirectionURL, translate, flowID, flow, subflow]);

    const handlePasswordKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            if (event.key === "Enter") {
                if (!password.length) {
                    focusPassword();
                }
                handleSignIn().catch(console.error);
                event.preventDefault();
            }
        },
        [focusPassword, handleSignIn, password.length],
    );

    const handlePasswordKeyUp = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            if (password.length <= 1) {
                setPasswordCapsLock(false);
                setPasswordCapsLockPartial(false);

                if (password.length === 0) {
                    return;
                }
            }

            const modified = IsCapsLockModified(event);

            if (modified === null) return;

            if (modified) {
                setPasswordCapsLock(true);
            } else {
                setPasswordCapsLockPartial(true);
            }
        },
        [password.length],
    );

    return (
        <form id={"form-password"} onSubmit={(e) => e.preventDefault()}>
            <div className="grid grid-cols-1 gap-4">
                <div className="w-full">
                    <Label htmlFor="password-textfield">{translate("Password")} *</Label>
                    <div className="relative">
                        <Input
                            ref={passwordRef}
                            id="password-textfield"
                            required
                            disabled={loading}
                            value={password}
                            className={cn("pr-10", passwordError && "border-destructive")}
                            onChange={(v) => setPassword(v.target.value)}
                            onFocus={() => setPasswordError(false)}
                            type={showPassword ? "text" : "password"}
                            autoComplete="current-password"
                            onKeyDown={handlePasswordKeyDown}
                            onKeyUp={handlePasswordKeyUp}
                        />
                        <button
                            type="button"
                            className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
                            aria-label={translate("Toggle password visibility")}
                            onMouseDown={() => setShowPassword(true)}
                            onMouseUp={() => setShowPassword(false)}
                            onMouseLeave={() => setShowPassword(false)}
                            onTouchStart={() => setShowPassword(true)}
                            onTouchEnd={() => setShowPassword(false)}
                            onTouchCancel={() => setShowPassword(false)}
                            onKeyDown={(e) => {
                                if (e.key === " ") {
                                    setShowPassword(true);
                                    e.preventDefault();
                                }
                            }}
                            onKeyUp={(e) => {
                                if (e.key === " ") {
                                    setShowPassword(false);
                                    e.preventDefault();
                                }
                            }}
                        >
                            {showPassword ? <Eye className="h-5 w-5" /> : <EyeOff className="h-5 w-5" />}
                        </button>
                    </div>
                </div>
                {passwordCapsLock ? (
                    <div className="w-full px-4">
                        <Alert variant="warning">
                            <AlertTitle>{translate("Warning")}</AlertTitle>
                            {passwordCapsLockPartial
                                ? translate("The password was partially entered with Caps Lock")
                                : translate("The password was entered with Caps Lock")}
                        </Alert>
                    </div>
                ) : null}
                <div className="w-full">
                    <Button
                        id="sign-in-button"
                        variant="default"
                        className="w-full"
                        disabled={loading}
                        onClick={handleSignIn}
                    >
                        {translate("Authenticate", { ns: "settings" })}
                        {loading ? <Spinner className="ml-2 h-5 w-5" /> : null}
                    </Button>
                </div>
            </div>
        </form>
    );
};

export default PasswordForm;
