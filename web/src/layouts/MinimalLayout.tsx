import { ReactNode, useEffect } from "react";

import { useTranslation } from "react-i18next";

import UserSvg from "@assets/images/user.svg?react";
import AppBarLoginPortal from "@components/AppBarLoginPortal";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { EncodedName } from "@constants/constants";
import { UserInfo } from "@models/UserInfo";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: null | string;
    userInfo?: UserInfo;
}

const MinimalLayout = function (props: Props) {
    const { t: translate } = useTranslation();

    const logo = getLogoOverride() ? (
        <img
            src="./static/media/logo.png"
            alt="Logo"
            className="mx-auto my-2 w-16"
            style={{ fill: "var(--custom-icon)" }}
        />
    ) : (
        <UserSvg className="mx-auto my-2 w-16" style={{ fill: "var(--custom-icon)" }} />
    );

    useEffect(() => {
        document.title = translate("Login - {{authelia}}", { authelia: atob(String.fromCodePoint(...EncodedName)) });
    }, [translate]);

    return (
        <div>
            <AppBarLoginPortal userInfo={props.userInfo} />
            <div id={props.id} className="flex min-h-[90vh] items-center justify-center text-center">
                <div className="mx-auto w-full max-w-[444px] px-8">
                    <div className="flex flex-col items-stretch">
                        <div className="w-full p-2">{logo}</div>
                        {props.title ? (
                            <div className="w-full">
                                <TypographyWithTooltip variant="h5" value={props.title} />
                            </div>
                        ) : null}
                        <div className="mt-2 py-2">{props.children}</div>
                    </div>
                </div>
                <PrivacyPolicyDrawer />
            </div>
        </div>
    );
};

export default MinimalLayout;
