import { ReactNode } from "react";

import { useTranslation } from "react-i18next";

import FingerTouchIcon from "@components/FingerTouchIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import TimerIcon from "@components/TimerIcon";
import { Button } from "@components/UI/Button";
import { Dialog, DialogContent, DialogFooter } from "@components/UI/Dialog";
import { SecondFactorMethod } from "@models/Methods";

export interface Props {
    open: boolean;
    methods: Set<SecondFactorMethod>;
    webauthn: boolean;

    onClose: () => void;
    onClick: (_method: SecondFactorMethod) => void;
}

const MethodSelectionDialog = function (props: Props) {
    const { t: translate } = useTranslation();
    const pieChartIcon = (
        <TimerIcon width={32} height={32} period={15} color="var(--primary)" backgroundColor={"white"} />
    );

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) props.onClose();
            }}
        >
            <DialogContent className="sm:max-w-xl text-center" showCloseButton={false}>
                <div className="grid grid-cols-1 justify-center gap-3" id="methods-dialog">
                    {props.methods.has(SecondFactorMethod.TOTP) ? (
                        <MethodItem
                            id="one-time-password-option"
                            method={translate("Time-based One-Time Password")}
                            icon={pieChartIcon}
                            onClick={() => props.onClick(SecondFactorMethod.TOTP)}
                        />
                    ) : null}
                    {props.methods.has(SecondFactorMethod.WebAuthn) && props.webauthn ? (
                        <MethodItem
                            id="webauthn-option"
                            method={translate("Security Key - WebAuthn")}
                            icon={<FingerTouchIcon size={32} />}
                            onClick={() => props.onClick(SecondFactorMethod.WebAuthn)}
                        />
                    ) : null}
                    {props.methods.has(SecondFactorMethod.MobilePush) ? (
                        <MethodItem
                            id="push-notification-option"
                            method={translate("Push Notification")}
                            icon={<PushNotificationIcon width={32} height={32} />}
                            onClick={() => props.onClick(SecondFactorMethod.MobilePush)}
                        />
                    ) : null}
                </div>
                <DialogFooter>
                    <Button variant="default" onClick={props.onClose}>
                        {translate("Close")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

interface MethodItemProps {
    readonly id: string;
    readonly method: string;
    readonly icon: ReactNode;

    readonly onClick: () => void;
}

function MethodItem(props: MethodItemProps) {
    return (
        <div className="method-option w-full" id={props.id}>
            <Button className="block h-auto w-full py-8 [&_svg]:!size-auto" variant="default" onClick={props.onClick}>
                <div className="mb-2 flex justify-center fill-white">{props.icon}</div>
                <div>
                    <p className="text-sm">{props.method}</p>
                </div>
            </Button>
        </div>
    );
}

export default MethodSelectionDialog;
