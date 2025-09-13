import React from "react";

import { Delete } from "@mui/icons-material";
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

    const handleDelete = () => {
        props.handleClose(true);
    };

    return (
        <Dialog open={props.open} onClose={handleCancel}>
            <DialogTitle>{props.title}</DialogTitle>
            <DialogContent>
                <DialogContentText my={2}>{props.text}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button id={"dialog-cancel"} onClick={handleCancel} data-1p-ignore>
                    {translate("Cancel")}
                </Button>
                <Button
                    id={"dialog-delete"}
                    variant={"outlined"}
                    color={"error"}
                    startIcon={<Delete />}
                    onClick={handleDelete}
                    data-1p-ignore
                >
                    {translate("Remove")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default DeleteDialog;
