import React, { Fragment } from "react";

import { Divider, Link, LinkProps } from "@mui/material";
import { useTranslation } from "react-i18next";

import { getPrivacyPolicyURL } from "@utils/Configuration";

export interface Props extends LinkProps {
    showDivider?: boolean;
    text?: string;
}

const PrivacyPolicyLink = function (props: Props) {
    const hrefPrivacyPolicy = getPrivacyPolicyURL();

    const { t: translate } = useTranslation();

    return (
        <Fragment>
            {props.showDivider ? <Divider orientation="vertical" flexItem variant="middle" /> : null}
            <Link {...props} href={hrefPrivacyPolicy} target="_blank" rel="noopener" underline="hover">
                {props.text === undefined ? translate("Privacy Policy") : props.text}
            </Link>
        </Fragment>
    );
};

export default PrivacyPolicyLink;
