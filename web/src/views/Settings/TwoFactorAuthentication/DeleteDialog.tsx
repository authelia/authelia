import { Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/UI/Dialog";

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
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleCancel();
            }}
        >
            <DialogContent showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{props.title}</DialogTitle>
                    <DialogDescription className="my-4">{props.text}</DialogDescription>
                </DialogHeader>
                <DialogFooter>
                    <Button id={"dialog-cancel"} variant={"outline"} onClick={handleCancel}>
                        {translate("Cancel")}
                    </Button>
                    <Button id={"dialog-delete"} variant={"destructive"} onClick={handleDelete}>
                        <Trash2 className="size-4" />
                        {translate("Remove")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default DeleteDialog;
