import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props {
    open: boolean;
    onConfirm: () => void;
    onCancel: () => void;
}

const VerifyExitDialog = (props: Props) => {
    const { t: translate } = useTranslation("settings");
    return (
        <Dialog open={props.open} onClose={props.onCancel}>
            <DialogTitle>{translate("Unsaved Changes")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {translate("You have unsaved changes. Are you sure you want to exit without saving")}?
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={props.onCancel}>Cancel</Button>
                <Button onClick={props.onConfirm} color="error">
                    {translate("Exit Without Saving")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default VerifyExitDialog;
