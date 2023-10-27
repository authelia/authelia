import React from "react";

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@mui/material";
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

    const handleRemove = () => {
        props.handleClose(true);
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{props.title}</DialogTitle>
            <DialogContent>
                <DialogContentText>{props.text}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancel}>{translate("Cancel")}</Button>
                <Button onClick={handleRemove} color={"error"}>
                    {translate("Remove")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default DeleteDialog;
