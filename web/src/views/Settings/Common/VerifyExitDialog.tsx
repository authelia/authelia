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
    onConfirm: () => void;
    onCancel: () => void;
}

const VerifyExitDialog = (props: Props) => {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) props.onCancel();
            }}
        >
            <DialogContent id="verify-exit-dialog" className="sm:max-w-md" showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Unsaved Changes")}</DialogTitle>
                    <DialogDescription>
                        {translate("You have unsaved changes. Are you sure you want to exit without saving?")}
                    </DialogDescription>
                </DialogHeader>
                <DialogFooter>
                    <Button id="verify-exit-cancel" variant="ghost" onClick={props.onCancel}>
                        {translate("Cancel")}
                    </Button>
                    <Button id="verify-exit-confirm" color="destructive" onClick={props.onConfirm}>
                        {translate("Exit Without Saving")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default VerifyExitDialog;
