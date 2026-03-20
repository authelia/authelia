import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

import { RedirectionURL } from "@constants/SearchParams";
import { useQueryParam } from "@hooks/QueryParam";

interface Props {
    open: boolean;
    setClosed: () => void;
}

const RedirectAfterEnrollmentDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const redirectionURL = useQueryParam(RedirectionURL);

    const targetURL = props.open ? redirectionURL : null;

    if (!targetURL) {
        return null;
    }

    return (
        <Dialog open={props.open} onClose={props.setClosed}>
            <DialogTitle>{translate("Multi-Factor Authentication Registered")}</DialogTitle>
            <DialogContent>
                <DialogContentText>{targetURL}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={props.setClosed}>{translate("Close")}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default RedirectAfterEnrollmentDialog;
