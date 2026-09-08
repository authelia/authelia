import { useTranslation } from "react-i18next";

import { Card, CardContent } from "@components/UI/Card";

const SettingsView = function () {
    const { t: translate } = useTranslation("settings");

    return (
        <Card>
            <CardContent className="p-6">
                <h4 className="text-2xl font-semibold text-center mb-2">{translate("User Settings")}</h4>
                <p className="text-center my-2">
                    {translate(
                        "This is the user settings area at the present time it's very minimal but will include new features in the near future",
                    )}
                </p>
                <p className="text-center my-2">
                    {translate("To view the currently available options select the menu icon at the top left")}
                </p>
            </CardContent>
        </Card>
    );
};

export default SettingsView;
