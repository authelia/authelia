import { ReactNode, useCallback, useEffect, useState } from "react";

import { Box, Breakpoint, Container, Theme, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";
import { FiUser } from "react-icons/fi";

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
        <Box className={classes.iconContainer}>
            <FiUser className={classes.icon} />
        </Box>
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
                        <Grid size={{ xs: 12 }}>
                            <Typography variant="h5" className={classes.welcomeHeading}>
                                Welcome to Adgone
                            </Typography>
                        </Grid>
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
        marginTop: theme.spacing(2),
        paddingBottom: theme.spacing(1),
        paddingTop: theme.spacing(1),
    },
    icon: {
        color: theme.custom.icon,
        width: "72px",
        height: "72px",
        strokeWidth: "1.5",
    },
    iconContainer: {
        margin: theme.spacing(2),
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
    },
    root: {
        minHeight: "90vh",
        textAlign: "center",
        background: "transparent",
    },
    rootContainer: {
        paddingLeft: 24,
        paddingRight: 24,
        backgroundColor: "#1e2b39",
        borderRadius: "16px",
        boxShadow: "0 2px 8px rgba(0, 0, 0, 0.15)",
        border: "1px solid #2f3d4d",
        backdropFilter: "blur(10px)",
        paddingTop: theme.spacing(3),
        paddingBottom: theme.spacing(3),
        animation: "fadeIn 0.6s ease-out",
        maxWidth: "440px !important",
    },
    subtitle: {
        color: theme.palette.text.secondary,
        fontSize: "0.95rem",
    },
    title: {
        fontWeight: 600,
        fontSize: "1.75rem",
        color: theme.palette.text.primary,
        marginBottom: theme.spacing(1),
    },
    welcomeHeading: {
        color: "#FFFFFF",
        fontSize: "1.5rem",
        fontWeight: 700,
        marginBottom: theme.spacing(0.5),
        marginTop: theme.spacing(1),
    },
    welcomeSubtitle: {
        color: theme.palette.text.primary,
        fontSize: "0.875rem",
        fontWeight: 400,
        marginBottom: theme.spacing(1),
    },
}));

export default LoginLayout;
