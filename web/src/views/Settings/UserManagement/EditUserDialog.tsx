import { useEffect, useState } from "react";

import { Button, Dialog, DialogContent, DialogTitle, FormControl, Grid, TextField, useTheme } from "@mui/material";
import { Path, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import ScaleLoader from "react-spinners/ScaleLoader";

import UserFormField from "@components/UserInputField.tsx";
import { useAllGroupsGET } from "@hooks/GroupManagement.ts";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserManagementAttributeMetadataGET } from "@hooks/UserManagement.ts";
import { UserDetailsExtended } from "@models/UserManagement";
import { patchChangeUser } from "@services/UserManagement";
import VerifyExitDialog from "@views/Settings/Common/VerifyExitDialog";

interface Props {
    user: null | UserDetailsExtended;
    open: boolean;
    onClose: () => void;
}

const EditUserDialog = ({ onClose, open, user }: Props) => {
    const { t: translate } = useTranslation("settings");
    const theme = useTheme();
    const { createErrorNotification, createSuccessNotification } = useNotifications();
    const [metadata, refetch, loading, error] = useUserManagementAttributeMetadataGET();
    const [groups, groupsRefetch, groupsLoading, groupsError] = useAllGroupsGET();
    const [verifyExitDialogOpen, setVerifyExitDialogOpen] = useState(false);

    useEffect(() => {
        if (open) {
            refetch();
            groupsRefetch();
        }
    }, [open, refetch, groupsRefetch]);
    const {
        control,
        formState: { dirtyFields, errors, isDirty },
        handleSubmit,
        register,
        reset,
        setValue,
    } = useForm<UserDetailsExtended>({
        defaultValues: user || {},
    });

    useEffect(() => {
        if (user) {
            reset(user);
        }
    }, [user, reset]);

    useEffect(() => {
        if (!open && user) {
            reset(user);
        }
    }, [open, user, reset]);

    const onSubmit = async (data: UserDetailsExtended) => {
        if (!user) return;

        try {
            const updateMask: string[] = [];
            const changedData: Partial<UserDetailsExtended> = {};
            const addressFields = ["street_address", "locality", "region", "postal_code", "country"];

            const { extra: extraFieldNames } = categorizeFields();

            (Object.keys(dirtyFields) as Array<keyof UserDetailsExtended>).forEach((key) => {
                if (addressFields.includes(key)) {
                    updateMask.push(`address.${key}`);
                    changedData[key] = data[key] as any;
                } else if (extraFieldNames.includes(key)) {
                    updateMask.push(`extra.${key}`);
                    if (!changedData.extra) {
                        changedData.extra = {};
                    }
                    changedData.extra[key] = data[key] as any;
                } else if (key === "extra" && dirtyFields.extra) {
                    Object.keys(dirtyFields.extra || {}).forEach((extraKey) => {
                        updateMask.push(`extra.${extraKey}`);
                    });
                    changedData[key] = data[key] as any;
                } else {
                    updateMask.push(key);
                    changedData[key] = data[key] as any;
                }
            });

            if (updateMask.length === 0) {
                createSuccessNotification(translate("No changes to save"));
                handleClose();
                return;
            }

            await patchChangeUser(user.username, changedData, updateMask);
            createSuccessNotification(translate("User modified successfully."));
            reset(data);
            onClose();
        } catch (e) {
            console.log(e);
            createErrorNotification(translate("Error modifying user"));
        }
    };

    const handleSafeClose = () => {
        if (isDirty) {
            setVerifyExitDialogOpen(true);
        } else {
            handleClose();
        }
    };

    const handleClose = () => {
        setVerifyExitDialogOpen(false);
        onClose();
    };

    const handleConfirmExit = () => {
        if (user) {
            reset(user);
        }
        handleClose();
    };

    const handleCancelExit = () => {
        setVerifyExitDialogOpen(false);
    };

    const basicFields = ["username", "groups", "mail"];

    const standardOptionalFields = [
        "display_name",
        "given_name",
        "middle_name",
        "family_name",
        "nickname",
        "phone_number",
        "phone_extension",
        "birthdate",
        "gender",
        "website",
        "profile",
        "picture",
        "zoneinfo",
        "locale",
        "street_address",
        "locality",
        "region",
        "postal_code",
        "country",
    ];

    // Fields that are excluded from editing
    const excludedFields = [
        "password",
        "last_logged_in",
        "last_password_change",
        "user_created_at",
        "method",
        "has_totp",
        "has_webauthn",
        "has_duo",
    ];

    const categorizeFields = () => {
        if (!metadata) return { basic: [], extra: [], optional: [], required: [] };

        const allSupportedFields = Object.keys(metadata.supported_attributes);

        const basic: string[] = [];
        const required: string[] = [];
        const optional: string[] = [];
        const extra: string[] = [];

        allSupportedFields.forEach((fieldName) => {
            if (excludedFields.includes(fieldName)) return;

            const isRequiredField = metadata.required_attributes.includes(fieldName);
            const isBasicField = basicFields.includes(fieldName);
            const isStandardOptionalField = standardOptionalFields.includes(fieldName);

            const isExtra = !isBasicField && !isStandardOptionalField;

            if (isBasicField) {
                basic.push(fieldName);
            } else if (isRequiredField && !isExtra) {
                required.push(fieldName);
            } else if (isExtra) {
                extra.push(fieldName);
            } else {
                optional.push(fieldName);
            }
        });

        basic.sort((a, b) => basicFields.indexOf(a) - basicFields.indexOf(b));

        required.sort((a, b) => standardOptionalFields.indexOf(a) - standardOptionalFields.indexOf(b));

        optional.sort((a, b) => standardOptionalFields.indexOf(a) - standardOptionalFields.indexOf(b));

        return { basic, extra, optional, required };
    };

    const buildFieldConfig = (fieldName: string) => {
        if (!metadata) return null;

        return {
            description: translate(`user_management.attributes.${fieldName}.description`, { defaultValue: "" }),
            disabled: fieldName === "username",
            label: translate(`user_management.attributes.${fieldName}.label`, { defaultValue: fieldName }),
            meta: metadata.supported_attributes[fieldName],
            name: fieldName as Path<UserDetailsExtended>,
            required: metadata.required_attributes.includes(fieldName),
        };
    };

    const { basic, extra, optional, required } = categorizeFields();

    const shownByDefaultFields = [...basic.map(buildFieldConfig), ...required.map(buildFieldConfig)].filter(Boolean);

    const additionalFields = [...optional.map(buildFieldConfig), ...extra.map(buildFieldConfig)].filter(Boolean);

    const [showAdditional, setShowAdditional] = useState(false);

    return (
        <>
            <Dialog open={open} onClose={handleSafeClose} maxWidth="sm" fullWidth>
                <DialogTitle>
                    {translate("Edit {{item}}:", { item: translate("User") })} {user?.username}
                </DialogTitle>

                <DialogContent>
                    {loading ||
                        (groupsLoading && <ScaleLoader color={theme.custom.loadingBar} speedMultiplier={1.5} />)}

                    {error && <div>Error loading users: {error.message}</div>}
                    {groupsError && <div>Error loading groups: {groupsError.message}</div>}

                    {!loading && !groupsLoading && !error && !groupsError && groups && metadata && (
                        <form onSubmit={handleSubmit(onSubmit)}>
                            <FormControl variant="standard">
                                <Grid container spacing={2}>
                                    {shownByDefaultFields.map((field) => (
                                        <Grid key={field!.name} size={12} sx={{ pt: 1.5 }}>
                                            {field!.disabled ? (
                                                <TextField
                                                    fullWidth
                                                    disabled
                                                    type="text"
                                                    color="info"
                                                    label={field!.label}
                                                    helperText={field!.description}
                                                    value={user?.[field!.name as keyof UserDetailsExtended] || ""}
                                                />
                                            ) : (
                                                <UserFormField
                                                    field={field!}
                                                    register={register}
                                                    control={control}
                                                    errors={errors}
                                                    setValue={setValue}
                                                    options={groups}
                                                />
                                            )}
                                        </Grid>
                                    ))}

                                    {additionalFields.length > 0 && (
                                        <Grid size={12} sx={{ pt: 2 }}>
                                            <Button
                                                onClick={() => setShowAdditional(!showAdditional)}
                                                size="small"
                                                variant="text"
                                                color="info"
                                            >
                                                {showAdditional
                                                    ? translate("Hide Additional Fields")
                                                    : translate("Show Additional Fields")}
                                            </Button>
                                        </Grid>
                                    )}

                                    {showAdditional &&
                                        additionalFields.map((field) => (
                                            <Grid key={field!.name} size={12} sx={{ pt: 1.5 }}>
                                                <UserFormField
                                                    field={field!}
                                                    register={register}
                                                    control={control}
                                                    errors={errors}
                                                    setValue={setValue}
                                                />
                                            </Grid>
                                        ))}

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
                                        <Button variant="contained" color="secondary" onClick={handleSafeClose}>
                                            {translate("Cancel")}
                                        </Button>
                                    </Grid>
                                </Grid>
                            </FormControl>
                        </form>
                    )}
                </DialogContent>
            </Dialog>
            <VerifyExitDialog open={verifyExitDialogOpen} onConfirm={handleConfirmExit} onCancel={handleCancelExit} />
        </>
    );
};

export default EditUserDialog;
