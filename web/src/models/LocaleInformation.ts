export interface LocaleInformation {
    defaults: {
        language: DefaultLanguage;
        namespace: string;
    };
    namespaces: string[];
    languages: Language[];
}

export interface DefaultLanguage {
    display: string;
    locale: string;
    parent?: string;
}

export interface Language {
    display: string;
    fallbacks: string[];
    locale: string;
    namespaces: string[];
    parent?: string;
}

export interface ChildLocale {
    display: string;
    locale: string;
}

export interface Locale extends ChildLocale {
    children: ChildLocale[];
}
