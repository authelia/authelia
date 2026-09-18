import { useEffect } from "react";

import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@components/UI/Dialog";
import { Field, FieldError, FieldLabel } from "@components/UI/Field";
import { Input } from "@components/UI/Input";
import { REGEX } from "@constants/Regex";
import { useNotifications } from "@contexts/NotificationsContext";
import { NewGroupRequest, postNewGroup } from "@services/GroupManagement";

interface Props {
    open: boolean;
    onClose: () => void;
}

const NewGroupDialog = ({ onClose, open }: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const {
        formState: { errors, isDirty },
        handleSubmit,
        register,
        reset,
    } = useForm<NewGroupRequest>({
        defaultValues: {
            name: "",
        },
    });

    useEffect(() => {
        if (!open) {
            reset();
        }
    }, [open, reset]);

    const onSubmit = async (data: NewGroupRequest) => {
        try {
            await postNewGroup(data);
            createSuccessNotification(translate("Group created successfully."));
            reset();
            onClose();
        } catch {
            createErrorNotification(translate("Error creating group"));
        }
    };

    const handleClose = () => {
        onClose();
    };

    return (
        <Dialog
            open={open}
            onOpenChange={(next) => {
                if (!next) handleClose();
            }}
        >
            <DialogContent id="new-group-dialog" className="sm:max-w-md" showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("New {{item}}", { item: translate("Group") })}</DialogTitle>
                </DialogHeader>

                <form noValidate className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
                    <Field data-invalid={!!errors.name}>
                        <FieldLabel htmlFor="new-group-name">{translate("Group Name")}</FieldLabel>
                        <Input
                            id="new-group-name"
                            type="text"
                            required
                            error={!!errors.name}
                            {...register("name", {
                                pattern: {
                                    message: translate("Invalid group name"),
                                    value: REGEX.GROUP,
                                },
                                required: translate("Group name is required"),
                            })}
                        />
                        {errors.name ? <FieldError>{errors.name.message}</FieldError> : null}
                    </Field>

                    <DialogFooter>
                        <Button id="new-group-cancel" type="button" variant="ghost" onClick={handleClose}>
                            {translate("Cancel")}
                        </Button>
                        <Button id="new-group-submit" type="submit" color="primary" disabled={!isDirty}>
                            {translate("Save")}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
};

export default NewGroupDialog;
