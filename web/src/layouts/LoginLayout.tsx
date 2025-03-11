import React, { ReactNode, useCallback, useEffect, useState } from "react";

import { Box, Container, Theme } from "@mui/material";
import Grid from "@mui/material/Grid2";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

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
    title?: string | null;
    titleTooltip?: string | null;
    subtitle?: string | null;
    subtitleTooltip?: string | null;
    userInfo?: UserInfo;
}

const LoginLayout = function (props: Props) {
    const { t: translate } = useTranslation();
    const { locale, setLocale } = useLanguageContext();

    const [localeList, setLocaleList] = useState<Language[]>([]);

    const styles = useStyles();

    const logo = getLogoOverride() ? (
        <img src="./static/media/logo.png" alt="Logo" className={styles.icon} />
    ) : (
        <UserSvg className={styles.icon} />
    );

    // handle the language selection
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
                className={styles.root}
                container
                spacing={0}
                alignItems="center"
                justifyContent="center"
            >
                <Container maxWidth="xs" className={styles.rootContainer}>
                    <Grid container>
                        <Grid size={{ xs: 12 }}>{logo}</Grid>
                        {props.title ? (
                            <Grid size={{ xs: 12 }}>
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
                        <Grid size={{ xs: 12 }} className={styles.body}>
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

const useStyles = makeStyles((theme: Theme) => ({
    root: {
        minHeight: "90vh",
        textAlign: "center",
    },
    rootContainer: {
        paddingLeft: 32,
        paddingRight: 32,
    },
    title: {},
    subtitle: {},
    icon: {
        margin: theme.spacing(),
        width: "64px",
        fill: theme.custom.icon,
    },
    body: {
        marginTop: theme.spacing(),
        paddingTop: theme.spacing(),
        paddingBottom: theme.spacing(),
    },
}));

export default LoginLayout;
