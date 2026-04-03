import Grid from "@mui/material/Grid";
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
            <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                <Grid size={{ xs: 12 }}>
                    <LogoutButton />
                </Grid>
                <Grid
                    size={{ xs: 12 }}
                    sx={{
                        border: "1px solid #d6d6d6",
                        borderRadius: "10px",
                        marginBottom: (theme) => theme.spacing(2),
                        marginTop: (theme) => theme.spacing(2),
                        padding: (theme) => theme.spacing(4),
                    }}
                >
                    <Authenticated />
                </Grid>
            </Grid>
        </MinimalLayout>
    );
};

export default AuthenticatedView;
