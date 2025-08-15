import React, { Fragment, useCallback, useEffect, useState } from "react";

import { ExpandLess, ExpandMore, Language as LanguageIcon } from "@mui/icons-material";
import { Box, Collapse, IconButton, ListItemText, Menu, MenuItem, Tooltip, Typography, useTheme } from "@mui/material";
import { useTranslation } from "react-i18next";

import { ChildLocale, Language, Locale } from "@models/LocaleInformation";

export interface Props {
    localeCurrent?: string;
    localeList?: Language[];
    onChange?: (lng: string) => void;
}

const Fallbacks: { [id: string]: string } = {
    sc: "Basa Sunda",
    ss: "Siswati",
    ty: "reo Tahiti",
    vec: "vèneto",
};

const AppBarItemLanguage = function (props: Props) {
    const { t: translate } = useTranslation();
    const theme = useTheme();

    const [elementLanguage, setElementLanguage] = useState<HTMLElement | null>(null);
    const open = Boolean(elementLanguage);
    const [expanded, setExpanded] = useState("");
    const [items, setItems] = useState<Locale[]>([]);
    const [current, setCurrent] = useState<ChildLocale | null>(null);

    const render = props.localeList !== undefined && props.localeCurrent !== undefined && props.onChange !== undefined;

    const handleMenuClick = (event: React.MouseEvent<HTMLElement>) => {
        setElementLanguage(event.currentTarget);
    };

    const closeMenu = useCallback(() => {
        setElementLanguage(null);
    }, []);

    const handleLanguageDisplayName = useCallback((locale: string, fallback: string) => {
        const browser = new Intl.DisplayNames(locale, { type: "language" }).of(locale);

        if (browser && browser !== locale && browser !== "") {
            return browser;
        }

        if (fallback !== "") {
            return fallback;
        }

        if (locale in Fallbacks) {
            return Fallbacks[locale];
        }

        console.error(
            `Error determining display value for locale ${locale} as it's unknown by both the browser and Golang, and does not have a unique fallback configured. Using the raw locale instead.`,
        );

        return browser || locale;
    }, []);

    const handleChange = useCallback(
        (language: ChildLocale) => {
            closeMenu();
            setCurrent(language);

            if (props.onChange) {
                props.onChange(language.locale);
            }
        },
        [closeMenu, props],
    );

    const handleCollapse = useCallback(
        (locale: string) => {
            if (locale === expanded) {
                setExpanded("");
            } else {
                setExpanded(locale);
            }
        },
        [expanded],
    );

    const filterParent = (locale: Language) => !locale.parent;
    const filterChildren = (parent: Language) => (locale: Language) =>
        locale.locale !== parent.locale && locale.parent === parent.locale;

    const generate = useCallback(() => {
        if (!props.localeList || !render) return;

        const locales = props.localeList;

        const items: Locale[] = locales.filter(filterParent).map((parent) => {
            const locale: Locale = {
                display: handleLanguageDisplayName(parent.locale, parent.display),
                locale: parent.locale,
                children: locales.filter(filterChildren(parent)).map((child) => {
                    return {
                        display: handleLanguageDisplayName(child.locale, child.display),
                        locale: child.locale,
                    };
                }),
            };

            if (locale.children.length === 1) {
                locale.locale = locale.children[0].locale;
            }

            return locale;
        });

        setItems(items);
    }, [props.localeList, render, handleLanguageDisplayName]);

    useEffect(() => {
        let localeNew: ChildLocale | null = null;

        for (const parent of items) {
            if (parent.locale === props.localeCurrent) {
                localeNew = parent;
                break;
            }

            for (const child of parent.children) {
                if (child.locale === props.localeCurrent) {
                    localeNew = child;
                    break;
                }
            }

            if (localeNew) {
                break;
            }
        }

        if (localeNew) {
            setCurrent(localeNew);
        }
    }, [items, props.localeCurrent]);

    useEffect(() => {
        generate();
    }, [generate]);

    return render ? (
        <Fragment>
            <Box sx={{ display: "flex", alignItems: "center", textAlign: "center" }}>
                <Tooltip title={translate("Language")}>
                    <IconButton
                        id={"language-button"}
                        key={"language-button"}
                        onClick={handleMenuClick}
                        sx={{ ml: 2 }}
                        aria-controls={open ? "language-menu" : undefined}
                        aria-expanded={open ? "true" : undefined}
                        aria-haspopup="true"
                    >
                        <LanguageIcon />
                        <Typography sx={{ paddingLeft: theme.spacing(1) }}>{current?.display}</Typography>
                    </IconButton>
                </Tooltip>
            </Box>
            <Menu
                anchorEl={elementLanguage}
                id="language-menu"
                open={open}
                onClose={closeMenu}
                slotProps={{
                    list: {
                        "aria-labelledby": "language-button",
                    },
                    paper: {
                        elevation: 0,
                        sx: {
                            maxHeight: { xs: "80vh", sm: "70vh", md: "50vh", lg: "40vh" },
                            filter: "drop-shadow(0px 2px 8px rgba(0,0,0,0.32))",
                            "&::before": {
                                content: '""',
                                position: "relative",
                                top: 0,
                                right: 14,
                                width: 10,
                                height: 10,
                                bgcolor: "background.paper",
                                transform: "translateY(-50%) rotate(45deg)",
                                zIndex: 0,
                            },
                        },
                    },
                }}
            >
                {items
                    .map((language) => {
                        const menuItems = [
                            <MenuItem
                                key={language.locale}
                                id={`language-${language.locale}`}
                                value={language.locale}
                                selected={props.localeCurrent === language.locale}
                            >
                                <ListItemText
                                    onClick={
                                        language.children.length <= 1
                                            ? () => handleChange(language)
                                            : () => handleCollapse(language.locale)
                                    }
                                >
                                    {language.display} ({language.locale})
                                </ListItemText>
                                {language.children.length <= 1 ? null : expanded === language.locale ? (
                                    <ExpandLess onClick={() => handleCollapse(language.locale)} />
                                ) : (
                                    <ExpandMore onClick={() => handleCollapse(language.locale)} />
                                )}
                            </MenuItem>,
                        ];

                        if (language.children.length > 1) {
                            menuItems.push(
                                <Collapse
                                    key={`${language.locale}-collapse`}
                                    in={expanded === language.locale}
                                    timeout="auto"
                                    onClick={() => handleCollapse(language.locale)}
                                >
                                    {language.children.map((child) => (
                                        <MenuItem
                                            id={`language-${language.locale}-child-${child.locale}`}
                                            key={`${language.locale}-child-${child.locale}`}
                                            onClick={() => handleChange(child)}
                                            value={child.locale}
                                            selected={props.localeCurrent === child.locale}
                                        >
                                            <ListItemText>
                                                &nbsp;&nbsp;{child.display} ({child.locale})
                                            </ListItemText>
                                        </MenuItem>
                                    ))}
                                </Collapse>,
                            );
                        }

                        return menuItems;
                    })
                    .flat()}
            </Menu>
        </Fragment>
    ) : null;
};

export default AppBarItemLanguage;
