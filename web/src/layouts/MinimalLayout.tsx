import React, { ReactNode, useEffect } from "react";

import { Box, Container } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import UserSvg from "@assets/images/user.svg?react";
import AppBarLoginPortal from "@components/AppBarLoginPortal";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { EncodedName } from "@constants/constants";
import { usePortalTemplate } from "@contexts/PortalTemplateContext";
import { usePortalStyles } from "@layouts/usePortalStyles";
import { UserInfo } from "@models/UserInfo";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string | null;
    userInfo?: UserInfo;
}

const MinimalLayout = function (props: Props) {
    const { t: translate } = useTranslation();

    const { definition } = usePortalTemplate();
    const classes = usePortalStyles(definition);

    const logo = getLogoOverride() ? (
        <Box component={"img"} src="./static/media/logo.png" alt="Logo" className={classes.icon} />
    ) : (
        <UserSvg className={classes.icon} />
    );

    useEffect(() => {
        document.title = translate("Login - {{authelia}}", { authelia: atob(String.fromCharCode(...EncodedName)) });
    }, [translate]);

    const rootEnhancer =
        `${classes.rootContainer} ${classes.typography} ${classes.links} ${classes.formElements} ${classes.buttons} ${classes.status}`.trim();

    return (
        <Box className={classes.page}>
            <AppBarLoginPortal userInfo={props.userInfo} />
            <Grid
                id={props.id}
                className={classes.root}
                container
                spacing={0}
                alignItems={"center"}
                justifyContent={"center"}
            >
                <Container maxWidth={"xs"} className={rootEnhancer}>
                    <Grid container className={classes.typography}>
                        <Grid size={{ xs: 12 }}>{logo}</Grid>
                        {props.title ? (
                            <Grid size={{ xs: 12 }}>
                                <TypographyWithTooltip variant={"h5"} value={props.title} />
                            </Grid>
                        ) : null}
                        <Grid size={{ xs: 12 }} className={classes.body}>
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
