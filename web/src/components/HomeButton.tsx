import { House } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
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
        <Button id={"home-button"} variant={"outline"} color={"default"} onClick={handleHomeClick}>
            <House />
            {translate("Home")}
        </Button>
    );
};

export default HomeButton;
