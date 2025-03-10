import React, { Fragment, useCallback, useEffect, useState } from "react";

import ExpandLess from "@mui/icons-material/ExpandLess";
import ExpandMore from "@mui/icons-material/ExpandMore";
import LanguageIcon from "@mui/icons-material/Language";
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

const ITEM_HEIGHT = 80;

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
            const language = {
                display: handleLanguageDisplayName(parent.locale, parent.display),
                locale: parent.locale,
                children: locales.filter(filterChildren(parent)).map((child) => {
                    const childLanguage = {
                        display: handleLanguageDisplayName(child.locale, child.display),
                        locale: child.locale,
                    };

                    if (props.localeCurrent === childLanguage.locale) {
                        setCurrent(childLanguage);
                    }

                    return childLanguage;
                }),
            };

            if (props.localeCurrent === language.locale) {
                setCurrent(language);
            }

            return language;
        });

        setItems(items);
    }, [props.localeList, props.localeCurrent, render, handleLanguageDisplayName]);

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
                        style: {
                            maxHeight: ITEM_HEIGHT * 4.5,
                        },
                        sx: {
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
                {items.map((language) => {
                    return (
                        <Fragment>
                            <MenuItem
                                key={language.locale}
                                value={language.locale}
                                selected={props.localeCurrent === language.locale}
                            >
                                <ListItemText
                                    key={`item-${language.locale}`}
                                    onClick={
                                        language.children.length <= 1
                                            ? () => handleChange(language)
                                            : () => handleCollapse(language.locale)
                                    }
                                >
                                    {language.display} ({language.locale})
                                </ListItemText>
                                {language.children.length <= 1 ? null : expanded === language.locale ? (
                                    <ExpandLess
                                        key={`expand-${language.locale}`}
                                        onClick={() => handleCollapse(language.locale)}
                                    />
                                ) : (
                                    <ExpandMore
                                        key={`expand-${language.locale}`}
                                        onClick={() => handleCollapse(language.locale)}
                                    />
                                )}
                            </MenuItem>
                            {language.children.length <= 1 ? null : (
                                <Collapse
                                    key={`collapse-${language.locale}`}
                                    in={expanded === language.locale}
                                    timeout="auto"
                                    onClick={() => handleCollapse(language.locale)}
                                >
                                    {language.children.map((child) => {
                                        return (
                                            <MenuItem
                                                key={`child-${child.locale}`}
                                                onClick={() => handleChange(child)}
                                                value={child.locale}
                                                selected={props.localeCurrent === language.locale}
                                            >
                                                <ListItemText key={`item-${child.locale}`}>
                                                    &nbsp;&nbsp;{child.display} ({child.locale})
                                                </ListItemText>
                                            </MenuItem>
                                        );
                                    })}
                                </Collapse>
                            )}
                        </Fragment>
                    );
                })}
            </Menu>
        </Fragment>
    ) : null;
};

export default AppBarItemLanguage;
