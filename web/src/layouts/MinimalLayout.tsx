import { ReactNode, useEffect } from "react";

import { Box, Container } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import UserSvg from "@assets/images/user.svg?react";
import AppBarLoginPortal from "@components/AppBarLoginPortal";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { EncodedName } from "@constants/constants";
import { UserInfo } from "@models/UserInfo";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: null | string;
    userInfo?: UserInfo;
}

const MinimalLayout = function (props: Props) {
    const { t: translate } = useTranslation();

    const logo = getLogoOverride() ? (
        <Box
            component={"img"}
            src="./static/media/logo.png"
            alt="Logo"
            sx={{ fill: (theme) => theme.custom.icon, margin: (theme) => theme.spacing(), width: "64px" }}
        />
    ) : (
        <Box
            component={UserSvg}
            sx={{ fill: (theme) => theme.custom.icon, margin: (theme) => theme.spacing(), width: "64px" }}
        />
    );

    useEffect(() => {
        document.title = translate("Login - {{authelia}}", { authelia: atob(String.fromCodePoint(...EncodedName)) });
    }, [translate]);

    return (
        <Box>
            <AppBarLoginPortal userInfo={props.userInfo} />
            <Grid
                id={props.id}
                container
                spacing={0}
                alignItems="center"
                justifyContent="center"
                sx={{ minHeight: "90vh", textAlign: "center" }}
            >
                <Container maxWidth="xs" sx={{ paddingLeft: "32px", paddingRight: "32px" }}>
                    <Grid container>
                        <Grid size={{ xs: 12 }}>{logo}</Grid>
                        {props.title ? (
                            <Grid size={{ xs: 12 }}>
                                <TypographyWithTooltip variant="h5" value={props.title} />
                            </Grid>
                        ) : null}
                        <Grid
                            size={{ xs: 12 }}
                            sx={{
                                marginTop: (theme) => theme.spacing(),
                                paddingBottom: (theme) => theme.spacing(),
                                paddingTop: (theme) => theme.spacing(),
                            }}
                        >
                            {props.children}
                        </Grid>
                    </Grid>
                </Container>
                <PrivacyPolicyDrawer />
            </Grid>
        </Box>
    );
};

export default MinimalLayout;
