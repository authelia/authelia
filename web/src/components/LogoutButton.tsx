import React from "react";

import { Button } from "@mui/material";
import { useTranslation } from "react-i18next";

import { LogoutRoute as SignOutRoute } from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";

export interface Props {}

const LogoutButton = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useRouterNavigate();

    const handleLogoutClick = () => {
        navigate(SignOutRoute);
    };

    return (
        <Button id={"logout-button"} color={"secondary"} onClick={handleLogoutClick}>
            {translate("Logout")}
        </Button>
    );
};

export default LogoutButton;
