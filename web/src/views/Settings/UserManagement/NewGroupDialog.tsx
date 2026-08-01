import { useEffect } from "react";

import { Button, Dialog, DialogContent, DialogTitle, FormControl, Grid, TextField } from "@mui/material";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

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
        } catch (e) {
            console.log(e);
            createErrorNotification(translate("Error creating group"));
        }
    };

    const handleClose = () => {
        onClose();
    };

    return (
        <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
            <DialogTitle>{translate("New {{item}}", { item: translate("Group") })}</DialogTitle>

            <DialogContent>
                <form onSubmit={handleSubmit(onSubmit)}>
                    <FormControl variant="standard" fullWidth>
                        <Grid container spacing={2}>
                            <Grid size={12} sx={{ pt: 2 }}>
                                <TextField
                                    fullWidth
                                    required
                                    type="text"
                                    color="info"
                                    label={translate("Group Name")}
                                    error={!!errors.name}
                                    helperText={errors.name?.message}
                                    {...register("name", {
                                        required: translate("Group name is required"),
                                    })}
                                />
                            </Grid>

                            <Grid size={12} sx={{ pt: 3 }}>
                                <Button
                                    type="submit"
                                    variant="contained"
                                    color="info"
                                    disabled={!isDirty}
                                    sx={{ mr: 2 }}
                                >
                                    {translate("Save")}
                                </Button>
                                <Button variant="contained" color="secondary" onClick={handleClose}>
                                    {translate("Cancel")}
                                </Button>
                            </Grid>
                        </Grid>
                    </FormControl>
                </form>
            </DialogContent>
        </Dialog>
    );
};

export default NewGroupDialog;
