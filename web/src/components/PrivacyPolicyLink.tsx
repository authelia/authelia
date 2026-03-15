import { ComponentProps } from "react";

import { useTranslation } from "react-i18next";

import { getPrivacyPolicyURL } from "@utils/Configuration";

const PrivacyPolicyLink = function (props: ComponentProps<"a">) {
    const { t: translate } = useTranslation();

    const hrefPrivacyPolicy = getPrivacyPolicyURL();

    return (
        <a {...props} href={hrefPrivacyPolicy} target="_blank" rel="noopener noreferrer" className={props.className}>
            {translate("Privacy Policy")}
        </a>
    );
};

export default PrivacyPolicyLink;
