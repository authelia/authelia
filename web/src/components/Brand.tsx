import { Fragment } from "react";

import { useTranslation } from "react-i18next";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";
import { Separator } from "@components/UI/Separator";
import { EncodedName, EncodedURL } from "@constants/constants";
import { getPrivacyPolicyEnabled } from "@utils/Configuration";

export interface Props {}

const Brand = function () {
    const { t: translate } = useTranslation();

    const privacyEnabled = getPrivacyPolicyEnabled();

    return (
        <div className="flex items-center justify-center gap-2 w-full">
            <a
                href={atob(String.fromCodePoint(...EncodedURL))}
                target="_blank"
                rel="noopener noreferrer"
                className="text-[#9e9e9e] text-[0.7rem] whitespace-nowrap hover:underline"
            >
                {translate("Powered by {{authelia}}", { authelia: atob(String.fromCodePoint(...EncodedName)) })}
            </a>
            {privacyEnabled ? (
                <Fragment>
                    <Separator orientation="vertical" className="h-4" />
                    <PrivacyPolicyLink className="text-[#9e9e9e] text-[0.7rem] whitespace-nowrap hover:underline" />
                </Fragment>
            ) : null}
        </div>
    );
};

export default Brand;
