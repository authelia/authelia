import { Button } from "@mui/material";
import { useTranslation } from "react-i18next";

import { IndexRoute } from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";

export interface Props {}

const HomeButton = function () {
    const { t: translate } = useTranslation(["portal"]);

    const navigate = useRouterNavigate();

    const handleHomeClick = () => {
        navigate(IndexRoute, false, false, false);
    };

    return (
        <Button id={"home-button"} color={"secondary"} onClick={handleHomeClick}>
            {translate("Home")}
        </Button>
    );
};

export default HomeButton;
