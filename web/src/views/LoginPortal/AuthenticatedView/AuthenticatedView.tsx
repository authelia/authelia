import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import MinimalLayout from "@layouts/MinimalLayout";
import { UserInfo } from "@models/UserInfo";
import Authenticated from "@views/LoginPortal/Authenticated";

export interface Props {
    userInfo: UserInfo;
}

const AuthenticatedView = function (props: Props) {
    const { t: translate } = useTranslation();

    return (
        <MinimalLayout
            id={"authenticated-stage"}
            title={`${translate("Hi")} ${props.userInfo.display_name}`}
            userInfo={props.userInfo}
        >
            <div className="flex flex-col items-center justify-center">
                <div className="w-full">
                    <LogoutButton />
                </div>
                <div className="my-4 w-full rounded-[10px] border border-[#d6d6d6] p-8">
                    <Authenticated />
                </div>
            </div>
        </MinimalLayout>
    );
};

export default AuthenticatedView;
