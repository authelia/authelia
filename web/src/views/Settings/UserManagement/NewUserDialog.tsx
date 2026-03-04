import { useEffect } from "react";

import { Button, Dialog, DialogContent, DialogTitle, FormControl, Grid, TextField } from "@mui/material";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import {
    CreateUserRequest,
    UserDetailsAddress,
    getFieldMetadata,
    isFieldRequired,
    validateFieldValue,
} from "@models/UserManagement";
import { UserFieldMetadataBody, postNewUser } from "@services/UserManagement";
import { generateRandomPassword } from "@utils/GeneratePassword";

interface Props {
    open: boolean;
    onClose: () => void;
    metadata: UserFieldMetadataBody; // Pass metadata as prop
}

const NewUserDialog = ({ metadata, onClose, open }: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const {
        formState: { errors, isDirty },
        handleSubmit,
        register,
        reset,
        setValue,
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
            await postNewUser(data);
            createSuccessNotification(translate("User created successfully."));
            onClose();
        } catch {
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

    return (
        <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
            <DialogTitle>{translate("New {{item}}", { item: translate("User") })}</DialogTitle>

            <DialogContent>
                <form onSubmit={handleSubmit(onSubmit)}>
                    <FormControl variant="standard">
                        <Grid container spacing={2}>
                            {metadata.supported_fields
                                .filter((fieldName): fieldName is keyof CreateUserRequest => {
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

                                    return !excludedFields.includes(fieldName);
                                })
                                .map((fieldName) => {
                                    const fieldMeta = getFieldMetadata(fieldName, metadata);
                                    const required = isFieldRequired(fieldName, metadata);
                                    const { baseType } = getInputType(fieldMeta?.type);

                                    if (["address", "groups"].includes(fieldName)) return null;

                                    return (
                                        <Grid key={fieldName} size={12} sx={{ pt: 1.5 }}>
                                            <TextField
                                                fullWidth
                                                {...register(fieldName, {
                                                    required: required
                                                        ? `${fieldMeta?.display_name || fieldName} is required`
                                                        : false,
                                                    validate: fieldMeta
                                                        ? (
                                                              value:
                                                                  | Record<string, any>
                                                                  | string
                                                                  | string[]
                                                                  | undefined
                                                                  | UserDetailsAddress,
                                                          ) => validateFieldValue(value, fieldMeta, fieldName) || true
                                                        : undefined,
                                                })}
                                                label={fieldMeta?.display_name || fieldName}
                                                type={
                                                    baseType === "email"
                                                        ? "email"
                                                        : baseType === "password"
                                                          ? "password"
                                                          : baseType === "url"
                                                            ? "url"
                                                            : "text"
                                                }
                                                error={!!errors[fieldName]}
                                                helperText={
                                                    errors[fieldName]?.message?.toString() || fieldMeta?.description
                                                }
                                                required={required}
                                                slotProps={{
                                                    htmlInput: {
                                                        maxLength: fieldMeta?.maxLength,
                                                    },
                                                }}
                                            />

                                            {fieldName === "password" && (
                                                <Button onClick={generatePassword} size="small" sx={{ mt: 0.75 }}>
                                                    {translate("Generate Password")}
                                                </Button>
                                            )}
                                        </Grid>
                                    );
                                })}

                            <Grid size={12} sx={{ pt: 3 }}>
                                <Button
                                    type="submit"
                                    variant="contained"
                                    color="primary"
                                    disabled={!isDirty}
                                    sx={{ mr: 2 }}
                                >
                                    {translate("Save")}
                                </Button>
                                <Button variant="outlined" color="secondary" onClick={handleClose}>
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

// Helper function for input types
const getInputType = (fieldType?: string, _isArray?: boolean) => {
    const parseFieldType = (typeString?: string) => {
        if (!typeString) return { baseType: undefined, isArray: false };

        if (typeString.endsWith("[]")) {
            return {
                baseType: typeString.slice(0, -2), // "email" from "email[]"
                fieldIsArray: true,
            };
        }

        if (typeString.includes("array")) {
            return {
                baseType: "string",
                fieldIsArray: true,
            };
        }
        return {
            baseType: typeString,
            fieldIsArray: false,
        };
    };

    const { baseType, fieldIsArray } = parseFieldType(fieldType);

    switch (baseType) {
        case "email":
            return { baseType, fieldIsArray };
        case "password":
            return { baseType, fieldIsArray };
        case "url":
            return { baseType, fieldIsArray };
        case "string":
            return { baseType, fieldIsArray };
        default:
            return { baseType: "string", fieldIsArray: false };
    }
};

export default NewUserDialog;
