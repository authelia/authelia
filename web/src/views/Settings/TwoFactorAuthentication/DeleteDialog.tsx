import React from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props {
    open: boolean;
    title: string;
    text: string;
    handleClose: (ok: boolean) => void;
}

const DeleteDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const handleCancel = () => {
        props.handleClose(false);
    };

    const handleDelete = () => {
        props.handleClose(true);
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{props.title}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <Typography my={2}>{props.text}</Typography>
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancel}>{translate("Cancel")}</Button>
                <Button variant={"outlined"} color={"error"} startIcon={<DeleteIcon />} onClick={handleDelete}>
                    {translate("Remove")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default DeleteDialog;
