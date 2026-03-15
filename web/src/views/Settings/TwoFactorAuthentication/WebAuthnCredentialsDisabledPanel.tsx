import { useTranslation } from "react-i18next";

import { Card, CardContent } from "@components/UI/Card";

const WebAuthnCredentialsDisabledPanel = function () {
    const { t: translate } = useTranslation("settings");

    return (
        <Card>
            <CardContent className="grid grid-cols-12 gap-4 p-4">
                <div className="col-span-12">
                    <h5 className="text-xl font-semibold">{translate("WebAuthn Credentials")}</h5>
                </div>
                <div className="col-span-12 text-center">
                    <h6 className="text-lg text-muted-foreground">
                        {translate(
                            "Your administrator has disabled WebAuthn preventing you from registering WebAuthn Credentials including Passkeys",
                        )}
                        .
                    </h6>
                </div>
                <div className="col-span-12 text-center">
                    <p className="text-sm">
                        <span>
                            {translate(
                                "WebAuthn Credentials are widely considered the most secure means of authentication, regardless of if they're used for Multi-Factor Authentication or Passwordless Authentication",
                            )}
                            .
                        </span>
                        <span>
                            {translate(
                                "The decision to disable WebAuthn Credentials when Multi-Factor Authentication is enabled significantly undermines security and is highly inadvisable",
                            )}
                            .
                        </span>
                    </p>
                </div>
            </CardContent>
        </Card>
    );
};

export default WebAuthnCredentialsDisabledPanel;
