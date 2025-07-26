import { Button, Drawer, DrawerProps, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { Trans, useTranslation } from "react-i18next";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";
import { EncodedName } from "@constants/constants";
import { LocalStoragePrivacyPolicyAccepted } from "@constants/LocalStorage";
import { usePersistentStorageValue } from "@hooks/PersistentStorage";
import { getPrivacyPolicyEnabled, getPrivacyPolicyRequireAccept } from "@utils/Configuration";

const PrivacyPolicyDrawer = function (props: DrawerProps) {
    const { t: translate } = useTranslation();

    const privacyEnabled = getPrivacyPolicyEnabled();
    const privacyRequireAccept = getPrivacyPolicyRequireAccept();
    const [accepted, setAccepted] = usePersistentStorageValue<boolean>(LocalStoragePrivacyPolicyAccepted, false);

    return privacyEnabled && privacyRequireAccept && !accepted ? (
        <Drawer {...props} anchor="bottom" open={!accepted}>
            <Grid
                container
                alignItems="center"
                justifyContent="center"
                textAlign="center"
                aria-labelledby="privacy-policy-drawer-title"
                aria-describedby="privacy-policy-drawer-description"
            >
                <Grid container size={{ xs: 12 }} paddingY={2}>
                    <Grid size={{ xs: 12 }}>
                        <Typography id="privacy-policy-drawer-title" variant="h6" component="h2">
                            {translate("Privacy Policy")}
                        </Typography>
                    </Grid>
                </Grid>
                <Grid size={{ xs: 12 }}>
                    <Typography id="privacy-policy-drawer-description">
                        <Trans
                            i18nKey="You must view and accept the Privacy Policy before using {{authelia}}."
                            values={{ authelia: atob(String.fromCharCode(...EncodedName)) }}
                            components={{
                                policy: <PrivacyPolicyLink />,
                            }}
                        />
                    </Typography>
                </Grid>
                <Grid size={{ xs: 12 }} paddingY={2}>
                    <Button
                        onClick={() => {
                            setAccepted(true);
                        }}
                        data-1p-ignore
                    >
                        {translate("Accept")}
                    </Button>
                </Grid>
            </Grid>
        </Drawer>
    ) : null;
};

export default PrivacyPolicyDrawer;
