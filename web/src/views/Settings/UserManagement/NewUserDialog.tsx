import { useEffect, useState } from "react";

import { Button, Dialog, DialogContent, DialogTitle, FormControl, Grid, TextField, useTheme } from "@mui/material";
import { Path, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import ScaleLoader from "react-spinners/ScaleLoader";

import UserFormField from "@components/UserInputField.tsx";
import { useAllGroupsGET } from "@hooks/GroupManagement.ts";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserManagementAttributeMetadataGET } from "@hooks/UserManagement.ts";
import {
    CreateUserRequest,
    getAttributeMetadata,
    isAttributeRequired,
    validateAttributeValue,
} from "@models/UserManagement";
import { UserAttributeMetadataBody, postNewUser } from "@services/UserManagement";
import { generateRandomPassword } from "@utils/GeneratePassword";

interface Props {
    open: boolean;
    onClose: () => void;
}

const NewUserDialog = ({ onClose, open }: Props) => {
    const { t: translate } = useTranslation("settings");
    const theme = useTheme();
    const { createErrorNotification, createSuccessNotification } = useNotifications();
    const [metadata, refetch, loading, error] = useUserManagementAttributeMetadataGET();
    const [groups, groupsRefetch, groupsLoading, groupsError] = useAllGroupsGET();

    useEffect(() => {
        if (open) {
            refetch();
            groupsRefetch();
        }
    }, [open, refetch, groupsRefetch]);

    const {
        control,
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

    useEffect(() => {
        if (!open) {
            reset();
        }
    }, [open, reset]);

    const onSubmit = async (data: CreateUserRequest) => {
        try {
            const { extra: extraFieldNames } = categorizeFields();
            const requestData: any = { ...data };

            const extraData: Record<string, any> = {};
            (Object.keys(data) as Array<keyof CreateUserRequest>).forEach((key) => {
                if (extraFieldNames.includes(key as string)) {
                    extraData[key as string] = data[key];
                    delete requestData[key];
                }
            });

            if (Object.keys(extraData).length > 0) {
                requestData.extra = extraData;
            }

            const response = await postNewUser(requestData);
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
            // Show confirmation dialog
        }
        onClose();
    };

    const generatePassword = () => {
        const newPassword = generateRandomPassword(16);
        setValue("password", newPassword, { shouldDirty: true });
    };

    const basicFields = ["username", "mail", "password", "groups"];

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
            label: translate(`user_management.attributes.${fieldName}.label`, { defaultValue: fieldName }),
            meta: metadata.supported_attributes[fieldName],
            name: fieldName as Path<CreateUserRequest>,
            required: metadata.required_attributes.includes(fieldName),
        };
    };

    const { basic = [], extra = [], optional = [], required = [] } = categorizeFields();

    const shownByDefaultFields = [...basic.map(buildFieldConfig), ...required.map(buildFieldConfig)].filter(Boolean);

    const additionalFields = [...optional.map(buildFieldConfig), ...extra.map(buildFieldConfig)].filter(Boolean);

    const [showAdditional, setShowAdditional] = useState(false);

    return (
        <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
            <DialogTitle>{translate("New {{item}}", { item: translate("User") })}</DialogTitle>

            <DialogContent>
                {loading || (groupsLoading && <ScaleLoader color={theme.custom.loadingBar} speedMultiplier={1.5} />)}

                {error && <div>Error loading users: {error.message}</div>}
                {groupsError && <div>Error loading groups: {groupsError.message}</div>}

                {!loading && !groupsLoading && !error && !groupsError && groups && metadata && (
                    <form onSubmit={handleSubmit(onSubmit)}>
                        <FormControl variant="standard">
                            <Grid container spacing={2}>
                                {shownByDefaultFields.map((field) => (
                                    <Grid key={field!.name} size={12} sx={{ pt: 1.5 }}>
                                        <UserFormField
                                            field={field!}
                                            register={register}
                                            control={control}
                                            errors={errors}
                                            setValue={setValue}
                                            options={groups}
                                            onGeneratePassword={
                                                field!.name === "password" ? generatePassword : undefined
                                            }
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
                                                onGeneratePassword={
                                                    field!.name === "password" ? generatePassword : undefined
                                                }
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

export default NewUserDialog;
