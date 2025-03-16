import React from "react";

import { AppBar, Toolbar, Typography } from "@mui/material";
import { styled } from "@mui/styles";

import AppBarItemAccountSettings from "@components/AppBarItemAccountSettings";
import AppBarItemLanguage from "@components/AppBarItemLanguage";
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
}));

const AppBarLoginPortal = function (props: Props) {
    return (
        <AppBar position="static" color="transparent" elevation={0}>
            <Typography style={{ flexGrow: 1 }} />
            <StyledToolbar variant={"regular"}>
                <Typography style={{ flexGrow: 1 }} />
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
