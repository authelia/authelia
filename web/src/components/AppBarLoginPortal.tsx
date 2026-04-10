import AppBarItemAccountSettings from "@components/AppBarItemAccountSettings";
import AppBarItemLanguage from "@components/AppBarItemLanguage";
import { Language } from "@models/LocaleInformation";
import { UserInfo } from "@models/UserInfo";

export interface Props {
    userInfo?: UserInfo;
    localeCurrent?: string;
    localeList?: Language[];
    onLocaleChange?: (_locale: string) => void;
}

const AppBarLoginPortal = function (props: Props) {
    return (
        <header className="bg-transparent">
            <div className="flex-grow" />
            <div className="mx-auto flex items-center px-4 pb-4 pt-2">
                <div className="flex-grow" />
                <AppBarItemLanguage
                    localeCurrent={props.localeCurrent}
                    localeList={props.localeList}
                    onChange={props.onLocaleChange}
                />
                <AppBarItemAccountSettings userInfo={props.userInfo} />
            </div>
        </header>
    );
};

export default AppBarLoginPortal;
