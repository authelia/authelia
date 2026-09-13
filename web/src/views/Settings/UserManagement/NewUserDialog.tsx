import { useEffect, useState } from "react";

import { Path, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { Alert, AlertDescription } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@components/UI/Dialog";
import { Spinner } from "@components/UI/Spinner";
import UserFormField from "@components/UserInputField";
import { useNotifications } from "@contexts/NotificationsContext";
import { useAllGroupsGET } from "@hooks/GroupManagement";
import { useUserManagementAttributeMetadataGET } from "@hooks/UserManagement";
import { CreateUserRequest } from "@models/UserManagement";
import { postNewUser } from "@services/UserManagement";
import { generateRandomPassword } from "@utils/GeneratePassword";
import VerifyExitDialog from "@views/Settings/Common/VerifyExitDialog";

interface Props {
    open: boolean;
    onClose: () => void;
}

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

const NewUserDialog = ({ onClose, open }: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();
    const [metadata, refetch, loading, error] = useUserManagementAttributeMetadataGET();
    const [groups, groupsRefetch, groupsLoading, groupsError] = useAllGroupsGET();
    const [showAdditional, setShowAdditional] = useState(false);
    const [verifyExitDialogOpen, setVerifyExitDialogOpen] = useState(false);

    const groupsFieldType = metadata?.supported_attributes?.groups?.type;
    const shouldFetchGroups = groupsFieldType === "groups";

    useEffect(() => {
        if (open) {
            refetch();
            if (shouldFetchGroups) {
                groupsRefetch();
            }
        }
    }, [open, refetch, groupsRefetch, shouldFetchGroups]);

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
            setShowAdditional(false);
            setVerifyExitDialogOpen(false);
        }
    }, [open, reset]);

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

            await postNewUser(requestData);
            createSuccessNotification(translate("User created successfully."));
            reset();
            onClose();
        } catch {
            createErrorNotification(translate("Error creating user"));
        }
    };

    const handleClose = () => {
        setVerifyExitDialogOpen(false);
        reset();
        onClose();
    };

    const handleSafeClose = () => {
        if (isDirty) {
            setVerifyExitDialogOpen(true);
        } else {
            handleClose();
        }
    };

    const handleConfirmExit = () => {
        handleClose();
    };

    const handleCancelExit = () => {
        setVerifyExitDialogOpen(false);
    };

    const generatePassword = () => {
        const newPassword = generateRandomPassword(16);
        setValue("password", newPassword, { shouldDirty: true });
    };

    const { basic = [], extra = [], optional = [], required = [] } = categorizeFields();

    const shownByDefaultFields = [...basic.map(buildFieldConfig), ...required.map(buildFieldConfig)].filter(Boolean);
    const additionalFields = [...optional.map(buildFieldConfig), ...extra.map(buildFieldConfig)].filter(Boolean);

    return (
        <>
            <Dialog
                open={open}
                onOpenChange={(next) => {
                    if (!next) handleSafeClose();
                }}
            >
                <DialogContent id="new-user-dialog" className="sm:max-w-lg" showCloseButton={false}>
                    <DialogHeader>
                        <DialogTitle>{translate("New {{item}}", { item: translate("User") })}</DialogTitle>
                    </DialogHeader>

                    {(loading || groupsLoading) && <Spinner size={20} />}

                    {error && (
                        <Alert variant="destructive">
                            <AlertDescription>
                                {translate("Error loading users")}: {error.message}
                            </AlertDescription>
                        </Alert>
                    )}
                    {groupsError && (
                        <Alert variant="destructive">
                            <AlertDescription>
                                {translate("Error loading groups")}: {groupsError.message}
                            </AlertDescription>
                        </Alert>
                    )}

                    {!loading && !groupsLoading && !error && !groupsError && metadata && (
                        <form noValidate className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
                            {shownByDefaultFields.map((field) => (
                                <UserFormField
                                    key={field!.name}
                                    idPrefix="new-user"
                                    field={field!}
                                    register={register}
                                    control={control}
                                    errors={errors}
                                    setValue={setValue}
                                    options={groups}
                                    onGeneratePassword={field!.name === "password" ? generatePassword : undefined}
                                />
                            ))}

                            {additionalFields.length > 0 && (
                                <Button
                                    id="new-user-toggle-additional"
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    className="w-fit"
                                    onClick={() => setShowAdditional(!showAdditional)}
                                >
                                    {showAdditional
                                        ? translate("Hide Additional Fields")
                                        : translate("Show Additional Fields")}
                                </Button>
                            )}

                            {showAdditional &&
                                additionalFields.map((field) => (
                                    <UserFormField
                                        key={field!.name}
                                        idPrefix="new-user"
                                        field={field!}
                                        register={register}
                                        control={control}
                                        errors={errors}
                                        setValue={setValue}
                                        onGeneratePassword={field!.name === "password" ? generatePassword : undefined}
                                    />
                                ))}

                            <DialogFooter>
                                <Button id="new-user-cancel" type="button" variant="ghost" onClick={handleSafeClose}>
                                    {translate("Cancel")}
                                </Button>
                                <Button id="new-user-submit" type="submit" color="primary" disabled={!isDirty}>
                                    {translate("Save")}
                                </Button>
                            </DialogFooter>
                        </form>
                    )}
                </DialogContent>
            </Dialog>
            <VerifyExitDialog open={verifyExitDialogOpen} onConfirm={handleConfirmExit} onCancel={handleCancelExit} />
        </>
    );
};

export default NewUserDialog;
