import React from "react";

import { AppBar, Toolbar, Typography } from "@mui/material";
import { styled } from "@mui/material/styles";

import AppBarItemAccountSettings from "@components/AppBarItemAccountSettings";
import AppBarItemLanguage from "@components/AppBarItemLanguage";
import AppBarItemTemplateSwitcher from "@components/AppBarItemTemplateSwitcher";
import { Language } from "@models/LocaleInformation";
import { UserInfo } from "@models/UserInfo";

export interface Props {
    userInfo?: UserInfo;
    localeCurrent?: string;
    localeList?: Language[];
    onLocaleChange?: (locale: string) => void;
}

const StyledToolbar = styled(Toolbar)(({ theme }) => ({
    alignItems: "flex-start",
    paddingTop: theme.spacing(1),
    paddingBottom: theme.spacing(2),
    marginX: "auto",
}));

const AppBarLoginPortal = function (props: Props) {
    return (
        <AppBar position="static" color="transparent" elevation={0}>
            <Typography sx={{ flexGrow: 1 }} />
            <StyledToolbar variant={"regular"}>
                <Typography sx={{ flexGrow: 1 }} />
                <AppBarItemTemplateSwitcher />
                <AppBarItemLanguage
                    localeCurrent={props.localeCurrent}
                    localeList={props.localeList}
                    onChange={props.onLocaleChange}
                />
                <AppBarItemAccountSettings userInfo={props.userInfo} />
            </StyledToolbar>
        </AppBar>
    );
};

export default AppBarLoginPortal;
