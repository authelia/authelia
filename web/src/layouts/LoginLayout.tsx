import { ReactNode, useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import UserSvg from "@assets/images/user.svg?react";
import AppBarLoginPortal from "@components/AppBarLoginPortal";
import Brand from "@components/Brand";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { EncodedName } from "@constants/constants";
import { useLanguageContext } from "@contexts/LanguageContext";
import { Language } from "@models/LocaleInformation";
import { UserInfo } from "@models/UserInfo";
import { getLocaleInformation } from "@services/LocaleInformation";
import { getLogoOverride } from "@utils/Configuration";
import { cn } from "@utils/Styles";

const maxWidthMap: Record<string, string> = {
    lg: "max-w-4xl",
    md: "max-w-2xl",
    sm: "max-w-xl",
    xl: "max-w-6xl",
    xs: "max-w-[444px]",
};

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: null | string;
    titleTooltip?: null | string;
    subtitle?: null | string;
    subtitleTooltip?: null | string;
    userInfo?: UserInfo;
    maxWidth?: false | string;
}

const LoginLayout = function (props: Props) {
    const { t: translate } = useTranslation();
    const { locale, setLocale } = useLanguageContext();

    const [localeList, setLocaleList] = useState<Language[]>([]);

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

    const handleChangeLanguage = (locale: string) => {
        setLocale(locale);
    };

    const fetchLocaleInformation = useCallback(async () => {
        try {
            const data = await getLocaleInformation();
            setLocaleList(data.languages);

            return data;
        } catch (err) {
            console.error("could not get locale list:", err);
        }
    }, []);

    useEffect(() => {
        const fetchData = async () => {
            await fetchLocaleInformation();
        };
        void fetchData();
    }, [fetchLocaleInformation]);

    useEffect(() => {
        document.title = translate("Login - {{authelia}}", { authelia: atob(String.fromCodePoint(...EncodedName)) });
    }, [translate]);

    const containerMaxWidth = props.maxWidth === false ? "" : (maxWidthMap[props.maxWidth ?? "xs"] ?? "max-w-xs");

    return (
        <div>
            <AppBarLoginPortal
                userInfo={props.userInfo}
                onLocaleChange={handleChangeLanguage}
                localeList={localeList}
                localeCurrent={locale}
            />
            <div id={props.id} className="flex min-h-[90vh] items-center justify-center text-center">
                <div className={cn("mx-auto w-full px-8", containerMaxWidth)}>
                    <div className="flex flex-col items-stretch">
                        <div className="w-full">{logo}</div>
                        {props.title ? (
                            <div className="w-full">
                                <TypographyWithTooltip
                                    variant="h5"
                                    value={props.title}
                                    tooltip={props.titleTooltip ?? undefined}
                                />
                            </div>
                        ) : null}
                        {props.subtitle ? (
                            <div className="w-full">
                                <TypographyWithTooltip
                                    variant="h6"
                                    value={props.subtitle}
                                    tooltip={props.subtitleTooltip ?? undefined}
                                />
                            </div>
                        ) : null}
                        <div className="mt-2 py-2">{props.children}</div>
                        <Brand />
                    </div>
                </div>
                <PrivacyPolicyDrawer />
            </div>
        </div>
    );
};

export default LoginLayout;
