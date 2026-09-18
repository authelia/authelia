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
    message: string;
    cancelText: string;
    confirmText: string;
    onConfirm: () => void;
    onCancel: () => void;
}

const VerifyActionDialog = (props: Props) => {
    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) props.onCancel();
            }}
        >
            <DialogContent id="verify-action-dialog" className="sm:max-w-md" showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{props.title}</DialogTitle>
                    <DialogDescription>{props.message}</DialogDescription>
                </DialogHeader>
                <DialogFooter>
                    <Button id="verify-action-cancel" variant="ghost" onClick={props.onCancel}>
                        {props.cancelText}
                    </Button>
                    <Button id="verify-action-confirm" color="destructive" onClick={props.onConfirm}>
                        {props.confirmText}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default VerifyActionDialog;
