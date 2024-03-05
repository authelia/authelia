import { Button, Drawer, DrawerProps, Grid, Typography } from "@mui/material";
import { Trans, useTranslation } from "react-i18next";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";
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
                <Grid container item xs={12} paddingY={2}>
                    <Grid item xs={12}>
                        <Typography id="privacy-policy-drawer-title" variant="h6" component="h2">
                            {translate("Privacy Policy")}
                        </Typography>
                    </Grid>
                </Grid>
                <Grid item xs={12}>
                    <Typography id="privacy-policy-drawer-description">
                        <Trans
                            i18nKey="You must view and accept the Privacy Policy before using"
                            components={[<PrivacyPolicyLink />]}
                        />{" "}
                        Authelia.
                    </Typography>
                </Grid>
                <Grid item xs={12} paddingY={2}>
                    <Button
                        onClick={() => {
                            setAccepted(true);
                        }}
                    >
                        {translate("Accept")}
                    </Button>
                </Grid>
            </Grid>
        </Drawer>
    ) : null;
};

export default PrivacyPolicyDrawer;
