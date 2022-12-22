import { Button, Drawer, DrawerProps, Grid, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";
import { usePersistentStorageValue } from "@hooks/PersistentStorage";
import { getPrivacyPolicyEnabled, getPrivacyPolicyRequireAccept } from "@utils/Configuration";

const PrivacyPolicyDrawer = function (props: DrawerProps) {
    const privacyEnabled = getPrivacyPolicyEnabled();
    const privacyRequireAccept = getPrivacyPolicyRequireAccept();
    const [accepted, setAccepted] = usePersistentStorageValue<boolean>("privacy-policy-accepted", false);
    const { t: translate } = useTranslation();

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
                        {translate("You must view and accept the Privacy Policy before using")} Authelia.
                    </Typography>
                </Grid>
                <Grid item xs={12} paddingY={2}>
                    <PrivacyPolicyLink />
                </Grid>
                <Grid item xs={12} paddingBottom={2}>
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
