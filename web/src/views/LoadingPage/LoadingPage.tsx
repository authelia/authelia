import React from "react";

import { useTranslation } from "react-i18next";

import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";

function LoadingPage() {
    const { t: translate } = useTranslation();

    return <BaseLoadingPage message={translate("Loading")} />;
}

export default LoadingPage;
