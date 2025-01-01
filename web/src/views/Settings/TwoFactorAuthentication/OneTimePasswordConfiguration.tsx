import React from "react";

import { UserInfoTOTPConfiguration } from "@models/TOTPConfiguration";
import OneTimePasswordCredentialItem from "@views/Settings/TwoFactorAuthentication/OneTimePasswordCredentialItem";

interface Props {
    config: UserInfoTOTPConfiguration;
    handleRefresh: () => void;
    handleInformation: (event: React.MouseEvent<HTMLElement>) => void;
    handleDelete: (event: React.MouseEvent<HTMLElement>) => void;
}

const OneTimePasswordConfiguration = function (props: Props) {
    return (
        <OneTimePasswordCredentialItem
            config={props.config}
            handleInformation={props.handleInformation}
            handleDelete={props.handleDelete}
        />
    );
};

export default OneTimePasswordConfiguration;
