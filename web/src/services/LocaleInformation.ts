import { DefaultLanguage, Language, LocaleInformation } from "@models/LocaleInformation";
import { LocaleInformationPath } from "@services/Api";
import { Get } from "@services/Client";

interface LocaleInformationPayload {
    defaults: {
        language: DefaultLanguage;
        namespace: string;
    };
    namespaces: Array<string>;
    languages: Array<Language>;
}

export async function getLocaleInformation(): Promise<LocaleInformation> {
    const data = await Get<LocaleInformationPayload>(LocaleInformationPath);

    return { ...data };
}
