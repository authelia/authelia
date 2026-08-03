import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

export interface Props {
    open: boolean;

    onChoice: (_rememberMe: boolean) => void;
}

const PasskeyRememberMeDialog = function (props: Props) {
    const { t: translate } = useTranslation();

    return (
        <Dialog id="passkey-remember-me-dialog" open={props.open} onClose={() => props.onChoice(false)}>
            <DialogTitle>{translate("Remember me?")}</DialogTitle>
            <DialogContent>
                <DialogContentText my={2}>
                    {translate("Would you like to stay signed in on this device?")}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button id="dialog-remember-me-no" onClick={() => props.onChoice(false)}>
                    {translate("No")}
                </Button>
                <Button
                    id="dialog-remember-me-yes"
                    variant="contained"
                    color="primary"
                    onClick={() => props.onChoice(true)}
                >
                    {translate("Yes")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default PasskeyRememberMeDialog;
