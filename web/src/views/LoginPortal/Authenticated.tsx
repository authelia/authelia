import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const { t: translate } = useTranslation();

    return (
        <div id="authenticated-stage">
            <div className="mb-4 flex-[0_0_100%]">
                <SuccessIcon />
            </div>
            <p>{translate("Authenticated")}</p>
        </div>
    );
};

export default Authenticated;
