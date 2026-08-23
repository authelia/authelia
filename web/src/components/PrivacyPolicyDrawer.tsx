import { Trans, useTranslation } from "react-i18next";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";
import { Button } from "@components/UI/Button";
import { Sheet, SheetContent, SheetTitle } from "@components/UI/Sheet";
import { EncodedName } from "@constants/constants";
import { LocalStoragePrivacyPolicyAccepted } from "@constants/LocalStorage";
import { usePersistentStorageValue } from "@hooks/PersistentStorage";
import { getPrivacyPolicyEnabled, getPrivacyPolicyRequireAccept } from "@utils/Configuration";

const PrivacyPolicyDrawer = function () {
    const { t: translate } = useTranslation();

    const privacyEnabled = getPrivacyPolicyEnabled();
    const privacyRequireAccept = getPrivacyPolicyRequireAccept();
    const [accepted, setAccepted] = usePersistentStorageValue<boolean>(LocalStoragePrivacyPolicyAccepted, false);

    return privacyEnabled && privacyRequireAccept && !accepted ? (
        <Sheet open={!accepted}>
            <SheetContent
                side="bottom"
                showCloseButton={false}
                aria-labelledby="privacy-policy-drawer-title"
                aria-describedby="privacy-policy-drawer-description"
            >
                <div className="flex flex-col items-center justify-center text-center gap-4 py-4">
                    <SheetTitle id="privacy-policy-drawer-title">{translate("Privacy Policy")}</SheetTitle>
                    <p id="privacy-policy-drawer-description">
                        <Trans
                            i18nKey="You must view and accept the Privacy Policy before using {{authelia}}."
                            values={{ authelia: atob(String.fromCharCode(...EncodedName)) }}
                            components={{
                                policy: <PrivacyPolicyLink />,
                            }}
                        />
                    </p>
                    <Button
                        onClick={() => {
                            setAccepted(true);
                        }}
                    >
                        {translate("Accept")}
                    </Button>
                </div>
            </SheetContent>
        </Sheet>
    ) : null;
};

export default PrivacyPolicyDrawer;
