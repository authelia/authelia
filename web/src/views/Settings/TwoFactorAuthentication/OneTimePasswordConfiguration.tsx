import React from "react";

import { QrCode2 } from "@mui/icons-material";
import { useTranslation } from "react-i18next";

import { UserInfoTOTPConfiguration, toAlgorithmString } from "@models/TOTPConfiguration";
import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";

interface Props {
    config: UserInfoTOTPConfiguration;
    handleRefresh: () => void;
    handleDelete: (event: React.MouseEvent<HTMLElement>) => void;
}

const OneTimePasswordConfiguration = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <CredentialItem
            id={"one-time-password"}
            icon={<QrCode2 fontSize="large" />}
            description={props.config.issuer}
            qualifier={
                " (" +
                translate("{{algorithm}}, {{digits}} digits, {{seconds}} seconds", {
                    algorithm: toAlgorithmString(props.config.algorithm),
                    digits: props.config.digits,
                    seconds: props.config.period,
                }) +
                ")"
            }
            created_at={props.config.created_at}
            last_used_at={props.config.last_used_at}
            handleDelete={props.handleDelete}
            tooltipDelete={translate("Remove this {{item}}", { item: translate("One-Time Password") })}
        />
    );
};

export default OneTimePasswordConfiguration;
