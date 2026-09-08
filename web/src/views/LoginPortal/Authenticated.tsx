import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const { t: translate } = useTranslation();

    return (
        <div id="authenticated-stage" className="flex flex-col items-center">
            <div className="mb-4">
                <SuccessIcon />
            </div>
            <p>{translate("Authenticated")}</p>
        </div>
    );
};

export default Authenticated;
