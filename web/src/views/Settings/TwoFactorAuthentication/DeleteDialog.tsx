import { Delete } from "@mui/icons-material";
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props {
    open: boolean;
    title: string;
    text: string;
    onConfirm: () => void;
    onCancel: () => void;
}

const DeleteDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const handleCancel = () => {
        props.onCancel();
    };

    const handleDelete = () => {
        props.onConfirm();
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{props.title}</DialogTitle>
            <DialogContent>
                <DialogContentText my={2}>{props.text}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button id={"dialog-cancel"} onClick={handleCancel}>
                    {translate("Cancel")}
                </Button>
                <Button
                    id={"dialog-delete"}
                    variant={"outlined"}
                    color={"error"}
                    startIcon={<Delete />}
                    onClick={handleDelete}
                >
                    {translate("Remove")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default DeleteDialog;
