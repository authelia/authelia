export interface LocaleInformation {
    defaults: {
        language: DefaultLanguage;
        namespace: string;
    };
    namespaces: Array<string>;
    languages: Array<Language>;
}

export interface DefaultLanguage {
    display: string;
    locale: string;
    parent?: string;
}

export interface Language {
    display: string;
    fallbacks: Array<string>;
    locale: string;
    namespaces: Array<string>;
    parent?: string;
}
