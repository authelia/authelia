import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props {
    open: boolean;
    setClosed: () => void;
}

const RedirectAfterEnrollmentDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog open={props.open} onClose={props.setClosed}>
            <DialogTitle>{translate("Multi-Factor Authentication Registered")}</DialogTitle>
            <DialogContent />
            <DialogActions>
                <Button onClick={props.setClosed}>{translate("Close")}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default RedirectAfterEnrollmentDialog;
