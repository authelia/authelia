import { Fragment, useCallback, useMemo, useState } from "react";

import { ChevronDown, ChevronUp, Languages } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Collapsible, CollapsibleContent } from "@components/UI/Collapsible";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/UI/DropdownMenu";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { ChildLocale, Language, Locale } from "@models/LocaleInformation";

export interface Props {
    localeCurrent?: string;
    localeList?: Language[];
    onChange?: (_lng: string) => void;
}

const Fallbacks: { [id: string]: string } = {
    sc: "Basa Sunda",
    ss: "Siswati",
    ty: "reo Tahiti",
    vec: "vèneto",
};

const AppBarItemLanguage = function (props: Props) {
    const { t: translate } = useTranslation();

    const [expanded, setExpanded] = useState("");

    const render = props.localeList !== undefined && props.localeCurrent !== undefined && props.onChange !== undefined;

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
            if (props.onChange) {
                props.onChange(language.locale);
            }
        },
        [props],
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

    const items = useMemo(() => {
        if (!props.localeList || !render) return [];

        const locales = props.localeList;

        return locales.filter(filterParent).map((parent) => {
            const locale: Locale = {
                children: locales.filter(filterChildren(parent)).map((child) => {
                    return {
                        display: handleLanguageDisplayName(child.locale, child.display),
                        locale: child.locale,
                    };
                }),
                display: handleLanguageDisplayName(parent.locale, parent.display),
                locale: parent.locale,
            };

            if (locale.children.length === 1) {
                locale.locale = locale.children[0].locale;
            }

            return locale;
        });
    }, [props.localeList, render, handleLanguageDisplayName]);

    const current = useMemo(() => {
        if (!items.length || !props.localeCurrent) return null;

        for (const parent of items) {
            if (parent.locale === props.localeCurrent) {
                return parent;
            }

            for (const child of parent.children) {
                if (child.locale === props.localeCurrent) {
                    return child;
                }
            }
        }

        return null;
    }, [items, props.localeCurrent]);

    return render ? (
        <Fragment>
            <div className="flex items-center text-center">
                <DropdownMenu>
                    <TooltipProvider>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <DropdownMenuTrigger asChild>
                                    <button
                                        id="language-button"
                                        className="ml-4 inline-flex items-center rounded-full p-2 hover:bg-accent focus:outline-none focus:ring-2 focus:ring-ring"
                                        aria-haspopup="true"
                                    >
                                        <Languages className="size-5" />
                                        <span className="pl-2">{current?.display}</span>
                                    </button>
                                </DropdownMenuTrigger>
                            </TooltipTrigger>
                            <TooltipContent>{translate("Language")}</TooltipContent>
                        </Tooltip>
                    </TooltipProvider>
                    <DropdownMenuContent
                        align="end"
                        className="max-h-[80vh] overflow-y-auto sm:max-h-[70vh] md:max-h-[50vh] lg:max-h-[40vh]"
                        aria-labelledby="language-button"
                    >
                        {items.flatMap((language) => {
                            const hasChildren = language.children.length > 1;
                            const isExpanded = expanded === language.locale;

                            const menuItems = [
                                <DropdownMenuItem
                                    key={language.locale}
                                    id={`language-${language.locale}`}
                                    className={props.localeCurrent === language.locale ? "bg-accent" : ""}
                                    onSelect={(e) => {
                                        if (language.children.length <= 1) {
                                            handleChange(language);
                                        } else {
                                            e.preventDefault();
                                            handleCollapse(language.locale);
                                        }
                                    }}
                                >
                                    <span className="flex-1">
                                        {language.display} ({language.locale})
                                    </span>
                                    {hasChildren ? (
                                        isExpanded ? (
                                            <ChevronUp className="size-4" />
                                        ) : (
                                            <ChevronDown className="size-4" />
                                        )
                                    ) : null}
                                </DropdownMenuItem>,
                            ];

                            if (language.children.length > 1) {
                                menuItems.push(
                                    <Collapsible
                                        key={`${language.locale}-collapse`}
                                        open={isExpanded}
                                        onOpenChange={() => handleCollapse(language.locale)}
                                    >
                                        <CollapsibleContent>
                                            {language.children.map((child) => (
                                                <DropdownMenuItem
                                                    id={`language-${language.locale}-child-${child.locale}`}
                                                    key={`${language.locale}-child-${child.locale}`}
                                                    className={
                                                        props.localeCurrent === child.locale ? "bg-accent pl-6" : "pl-6"
                                                    }
                                                    onSelect={() => handleChange(child)}
                                                >
                                                    {child.display} ({child.locale})
                                                </DropdownMenuItem>
                                            ))}
                                        </CollapsibleContent>
                                    </Collapsible>,
                                );
                            }

                            return menuItems;
                        })}
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
        </Fragment>
    ) : null;
};

export default AppBarItemLanguage;
