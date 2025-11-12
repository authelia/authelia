import { ReactNode, useCallback, useEffect, useState } from "react";

import { Box, Breakpoint, Container, Theme } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import UserSvg from "@assets/images/user.svg?react";
import AppBarLoginPortal from "@components/AppBarLoginPortal";
import Brand from "@components/Brand";
import PrivacyPolicyDrawer from "@components/PrivacyPolicyDrawer";
import TypographyWithTooltip from "@components/TypographyWithTooltip";
import { EncodedName } from "@constants/constants";
import { useLanguageContext } from "@contexts/LanguageContext";
import { Language } from "@models/LocaleInformation";
import { UserInfo } from "@models/UserInfo";
import { getLocaleInformation } from "@services/LocaleInformation";
import { getLogoOverride } from "@utils/Configuration";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: null | string;
    titleTooltip?: null | string;
    subtitle?: null | string;
    subtitleTooltip?: null | string;
    userInfo?: UserInfo;
    maxWidth?: Breakpoint | false;
}

const LoginLayout = function (props: Props) {
    const { t: translate } = useTranslation();
    const { locale, setLocale } = useLanguageContext();

    const [localeList, setLocaleList] = useState<Language[]>([]);

    const { classes } = useStyles();

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
        const fetchData = async () => {
            await fetchLocaleInformation();
        };
        void fetchData();
    }, [fetchLocaleInformation]);

    useEffect(() => {
        document.title = translate("Login - {{authelia}}", { authelia: atob(String.fromCodePoint(...EncodedName)) });
    }, [translate]);

    return (
        <Box>
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
            >
                <Container maxWidth={props.maxWidth ?? "xs"} className={classes.rootContainer}>
                    <Grid container>
                        <Grid size={{ xs: 12 }}>{logo}</Grid>
                        {props.title ? (
                            <Grid size={{ xs: 12 }} maxWidth={"xs"}>
                                <TypographyWithTooltip
                                    variant={"h5"}
                                    value={props.title}
                                    tooltip={props.titleTooltip ?? undefined}
                                />
                            </Grid>
                        ) : null}
                        {props.subtitle ? (
                            <Grid size={{ xs: 12 }}>
                                <TypographyWithTooltip
                                    variant={"h6"}
                                    value={props.subtitle}
                                    tooltip={props.subtitleTooltip ?? undefined}
                                />
                            </Grid>
                        ) : null}
                        <Grid size={{ xs: 12 }} className={classes.body}>
                            {props.children}
                        </Grid>
                        <Brand />
                    </Grid>
                </Container>
                <PrivacyPolicyDrawer />
            </Grid>
        </Box>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    body: {
        marginTop: theme.spacing(),
        paddingBottom: theme.spacing(),
        paddingTop: theme.spacing(),
    },
    icon: {
        fill: theme.custom.icon,
        margin: theme.spacing(),
        width: "64px",
    },
    root: {
        minHeight: "90vh",
        textAlign: "center",
    },
    rootContainer: {
        paddingLeft: 32,
        paddingRight: 32,
    },
    subtitle: {},
    title: {},
}));

export default LoginLayout;
