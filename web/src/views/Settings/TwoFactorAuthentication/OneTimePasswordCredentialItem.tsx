import { Fragment, MouseEvent, useState } from "react";

import { QrCode2 } from "@mui/icons-material";
import { useTranslation } from "react-i18next";

import { UserInfoTOTPConfiguration } from "@models/TOTPConfiguration";
import CredentialItem from "@views/Settings/TwoFactorAuthentication/CredentialItem";
import OneTimePasswordInformationDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordInformationDialog";

interface Props {
    config: UserInfoTOTPConfiguration;
    handleInformation: (_event: MouseEvent<HTMLElement>) => void;
    handleDelete: (_event: MouseEvent<HTMLElement>) => void;
}

const OneTimePasswordCredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const [showDialogDetails, setShowDialogDetails] = useState<boolean>(false);

    return (
        <Fragment>
            <OneTimePasswordInformationDialog
                config={props.config}
                open={showDialogDetails}
                handleClose={() => {
                    setShowDialogDetails(false);
                }}
            />
            <CredentialItem
                id={"one-time-password"}
                icon={<QrCode2 fontSize="large" />}
                description={props.config.issuer}
                qualifier={``}
                created_at={props.config.created_at}
                last_used_at={props.config.last_used_at}
                handleDelete={props.handleDelete}
                handleInformation={props.handleInformation}
                tooltipInformation={translate("Display extended information for this One-Time Password")}
                tooltipDelete={translate("Remove this {{item}}", { item: translate("One-Time Password") })}
            />
        </Fragment>
    );
};

export default OneTimePasswordCredentialItem;
