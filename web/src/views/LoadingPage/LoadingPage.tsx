import { useTranslation } from "react-i18next";

import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";

const LoadingPage = function () {
    const { t: translate } = useTranslation();

    return <BaseLoadingPage message={translate("Loading")} />;
};

export default LoadingPage;
