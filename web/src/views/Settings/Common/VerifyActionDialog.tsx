import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";

interface Props {
    open: boolean;
    title: string;
    message: string;
    cancelText: string;
    confirmText: string;
    onConfirm: () => void;
    onCancel: () => void;
}

const VerifyActionDialog = (props: Props) => {
    return (
        <Dialog open={props.open} onClose={props.onCancel}>
            <DialogTitle>{props.title}</DialogTitle>
            <DialogContent>
                <DialogContentText>{props.message}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={props.onCancel}>{props.cancelText}</Button>
                <Button onClick={props.onConfirm} color="error">
                    {props.confirmText}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default VerifyActionDialog;
