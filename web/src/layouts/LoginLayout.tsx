import React, { ReactNode, useCallback, useEffect, useState } from "react";

import { Box, Breakpoint, Container } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import UserSvg from "@assets/images/user.svg?react";
import AppBarLoginPortal from "@components/AppBarLoginPortal";
import Brand from "@components/Brand";
import PortalTemplateEffectHost from "@components/PortalTemplateEffectHost";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { EncodedName } from "@constants/constants";
import { useLanguageContext } from "@contexts/LanguageContext";
import { usePortalTemplate } from "@contexts/PortalTemplateContext";
import { usePortalStyles } from "@layouts/usePortalStyles";
import { Language } from "@models/LocaleInformation";
import { UserInfo } from "@models/UserInfo";
import { getLocaleInformation } from "@services/LocaleInformation";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string | null;
    titleTooltip?: string | null;
    subtitle?: string | null;
    subtitleTooltip?: string | null;
    userInfo?: UserInfo;
    maxWidth?: false | Breakpoint;
}

const LoginLayout = function (props: Props) {
    const { t: translate } = useTranslation();
    const { locale, setLocale } = useLanguageContext();

    const [localeList, setLocaleList] = useState<Language[]>([]);

    const { definition } = usePortalTemplate();
    const classes = usePortalStyles(definition);

    const logo = getLogoOverride() ? (
        <Box component={"img"} src="./static/media/logo.png" alt="Logo" className={classes.icon} />
    ) : (
        <UserSvg className={classes.icon} />
    );

    const handleChangeLanguage = (locale: string) => {
        setLocale(locale);
    };

    const fetchLocaleInformation = useCallback(async () => {
        try {
            const data = await getLocaleInformation();
            setLocaleList(data.languages);

            return data;
        } catch (err) {
            console.error("could not get locale list:", err);
        }
    }, []);

    useEffect(() => {
        fetchLocaleInformation().then();
    }, [fetchLocaleInformation]);

    useEffect(() => {
        document.title = translate("Login - {{authelia}}", { authelia: atob(String.fromCharCode(...EncodedName)) });
    }, [translate]);

    const rootEnhancer =
        `${classes.rootContainer} ${classes.typography} ${classes.links} ${classes.formElements} ${classes.buttons} ${classes.status}`.trim();

    const templateMaxWidth = definition.style.layout?.maxWidth;
    const inferredMaxWidth: false | Breakpoint = (() => {
        if (props.maxWidth !== undefined) {
            return props.maxWidth;
        }
        if (templateMaxWidth === undefined) {
            return "xs";
        }
        if (templateMaxWidth === false) {
            return false;
        }
        return templateMaxWidth as Breakpoint;
    })();

    return (
        <Box className={classes.page} data-portal-role="page">
            <PortalTemplateEffectHost className={classes.effectHost} />
            <AppBarLoginPortal
                userInfo={props.userInfo}
                onLocaleChange={handleChangeLanguage}
                localeList={localeList}
                localeCurrent={locale}
            />
            <Grid
                id={props.id}
                className={classes.root}
                container
                spacing={0}
                alignItems="center"
                justifyContent="center"
                data-portal-role="root"
            >
                <Container maxWidth={inferredMaxWidth} className={rootEnhancer} data-portal-role="card">
                    <Grid container className={classes.typography} data-portal-role="content">
                        <Grid size={{ xs: 12 }}>{logo}</Grid>
                        {props.title ? (
                            <Grid size={{ xs: 12 }} maxWidth={"xs"}>
                                <TypographyWithTooltip
                                    variant={"h5"}
                                    value={props.title}
                                    tooltip={props.titleTooltip !== null ? props.titleTooltip : undefined}
                                />
                            </Grid>
                        ) : null}
                        {props.subtitle ? (
                            <Grid size={{ xs: 12 }}>
                                <TypographyWithTooltip
                                    variant={"h6"}
                                    value={props.subtitle}
                                    tooltip={props.subtitleTooltip !== null ? props.subtitleTooltip : undefined}
                                />
                            </Grid>
                        ) : null}
                        <Grid size={{ xs: 12 }} className={classes.body}>
                            {props.children}
                        </Grid>
                        <Grid size={{ xs: 12 }} className={classes.links}>
                            <Brand />
                        </Grid>
                    </Grid>
                </Container>
                <PrivacyPolicyDrawer />
            </Grid>
        </Box>
    );
};

export default LoginLayout;
