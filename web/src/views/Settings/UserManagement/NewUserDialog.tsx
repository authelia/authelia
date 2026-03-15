import { useEffect, useState } from "react";

import { Button, Dialog, DialogContent, DialogTitle, FormControl, Grid, TextField, useTheme } from "@mui/material";
import { Path, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import {
    CreateUserRequest,
    getAttributeMetadata,
    isAttributeRequired,
    validateAttributeValue,
} from "@models/UserManagement";
import { UserAttributeMetadataBody, postNewUser } from "@services/UserManagement";
import { generateRandomPassword } from "@utils/GeneratePassword";
import { useUserManagementAttributeMetadataGET } from "@hooks/UserManagement.ts";
import ScaleLoader from "react-spinners/ScaleLoader";
import UserFormField from "@components/UserInputField.tsx";

interface Props {
    open: boolean;
    onClose: () => void;
}

const NewUserDialog = ({ onClose, open }: Props) => {
    const { t: translate } = useTranslation("settings");
    const theme = useTheme();
    const { createErrorNotification, createSuccessNotification } = useNotifications();
    const [metadata, refetch, loading, error] = useUserManagementAttributeMetadataGET();

    useEffect(() => {
        if (open) {
            refetch();
        }
    }, [open, refetch]);

    const {
        formState: { errors, isDirty },
        handleSubmit,
        register,
        reset,
        setValue,
        control,
    } = useForm<CreateUserRequest>({
        defaultValues: {
            password: "",
            username: "",
        },
    });

    // Reset form when dialog closes
    useEffect(() => {
        if (!open) {
            reset();
        }
    }, [open, reset]);

    const onSubmit = async (data: CreateUserRequest) => {
        try {
            const response = await postNewUser(data);
            console.log("Response:", response);
            createSuccessNotification(translate("User created successfully."));
            reset();
            onClose();
        } catch (e) {
            console.log(e);
            createErrorNotification(translate("Error creating user"));
        }
    };

    const handleClose = () => {
        if (isDirty) {
            // Show confirmation dialog if needed
            // For now, just close
        }
        onClose();
    };

    const generatePassword = () => {
        const newPassword = generateRandomPassword(16);
        setValue("password", newPassword, { shouldDirty: true });
    };

    const fieldConfig = {
        basic: [
            "username",
            "display_name",
            "given_name",
            "family_name",
            "password",
            "mail",
        ],
        additional: [
            "middle_name",
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
        ],
    };

    // Fields that are excluded from CreateUserRequest
    const excludedFields = [
        "last_logged_in",
        "last_password_change",
        "user_created_at",
        "method",
        "has_totp",
        "has_webauthn",
        "has_duo",
    ];

    const getOrderedFields = (fieldNames: string[]) => {
        if (!metadata) return [];

        return fieldNames
            .filter(fieldName =>
                metadata.supported_attributes[fieldName] &&
                !excludedFields.includes(fieldName)
            )
            .map(fieldName => ({
                name: fieldName as Path<CreateUserRequest>,
                meta: metadata.supported_attributes[fieldName],
                label: translate(`user_management.attributes.${fieldName}.label`, { defaultValue: fieldName }),
                description: translate(`user_management.attributes.${fieldName}.description`, { defaultValue: "" }),
                required: metadata.required_attributes.includes(fieldName),
            }));
    };

    const basicFields = getOrderedFields(fieldConfig.basic);
    const additionalFields = getOrderedFields(fieldConfig.additional);

    const [showAdditional, setShowAdditional] = useState(false);


    return (
        <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
            <DialogTitle>{translate("New {{item}}", { item: translate("User") })}</DialogTitle>

            <DialogContent>
                {loading && <ScaleLoader color={theme.custom.loadingBar} speedMultiplier={1.5} />}

                {error && <div>Error loading content: {error.message}</div>}

                {!loading && !error && metadata && (
                    <form onSubmit={handleSubmit(onSubmit)}>
                        <FormControl variant="standard">
                            <Grid container spacing={2}>
                                {basicFields.map((field) => (
                                    <Grid key={field.name} size={12} sx={{ pt: 1.5 }}>
                                        <UserFormField
                                            field={field}
                                            register={register}
                                            control={control}
                                            errors={errors}
                                            setValue={setValue}
                                            onGeneratePassword={field.name === "password" ? generatePassword : undefined}
                                        />
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
                                                : translate("Show Additional Fields")
                                            }
                                        </Button>
                                    </Grid>
                                )}

                                {showAdditional && additionalFields.map((field) => (
                                    <Grid key={field.name} size={12} sx={{ pt: 1.5 }}>
                                        <UserFormField
                                            field={field}
                                            register={register}
                                            control={control}
                                            errors={errors}
                                            setValue={setValue}
                                            onGeneratePassword={field.name === "password" ? generatePassword : undefined}
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
                                    <Button variant="contained" color="secondary" onClick={handleClose}>
                                        {translate("Cancel")}
                                    </Button>
                                </Grid>
                            </Grid>
                        </FormControl>
                    </form>
                )}
            </DialogContent>
        </Dialog>
    );
};

// Helper function for input types - simplified
const getInputType = (fieldType: string): string => {
    switch (fieldType) {
        case "email":
            return "email";
        case "password":
            return "password";
        case "url":
            return "url";
        case "tel":
            return "tel";
        case "date":
            return "date";
        default:
            return "text";
    }
};
export default NewUserDialog;
