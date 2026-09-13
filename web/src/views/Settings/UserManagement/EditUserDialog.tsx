import { useEffect, useState } from "react";

import { Path, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { Alert, AlertDescription } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@components/UI/Dialog";
import { Field, FieldDescription, FieldLabel } from "@components/UI/Field";
import { Input } from "@components/UI/Input";
import { Spinner } from "@components/UI/Spinner";
import UserFormField from "@components/UserInputField";
import { useNotifications } from "@contexts/NotificationsContext";
import { useAllGroupsGET } from "@hooks/GroupManagement";
import { useUserManagementAttributeMetadataGET } from "@hooks/UserManagement";
import { UserDetailsExtended } from "@models/UserManagement";
import { patchChangeUser } from "@services/UserManagement";
import VerifyExitDialog from "@views/Settings/Common/VerifyExitDialog";

interface Props {
    user: null | UserDetailsExtended;
    open: boolean;
    onClose: () => void;
}

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

const addressFields = ["street_address", "locality", "region", "postal_code", "country"];

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

const EditUserDialog = ({ onClose, open, user }: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();
    const [metadata, refetch, loading, error] = useUserManagementAttributeMetadataGET();
    const [groups, groupsRefetch, groupsLoading, groupsError] = useAllGroupsGET();
    const [showAdditional, setShowAdditional] = useState(false);
    const [verifyExitDialogOpen, setVerifyExitDialogOpen] = useState(false);
    const [wasOpen, setWasOpen] = useState(open);

    // Check if groups field uses "groups" type (LDAP) or "text" type (file provider)
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
            reset({ ...user, address: user.address || {} });
        }
    }, [user, reset]);

    if (open !== wasOpen) {
        setWasOpen(open);
        if (!open && user) {
            reset(user);
            setShowAdditional(false);
            setVerifyExitDialogOpen(false);
        }
    }

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

    const { basic, extra, optional, required } = categorizeFields();

    const buildFieldConfig = (fieldName: string) => {
        if (!metadata) return null;

        const isAddressField = addressFields.includes(fieldName);
        // Extra attribute names from the backend already carry the "extra." prefix (e.g. "extra.test_flag"),
        // so they need no further prefixing here - only bare address subfields do.
        const fieldPath = isAddressField ? `address.${fieldName}` : fieldName;

        return {
            description: translate(`user_management.attributes.${fieldName}.description`, { defaultValue: "" }),
            disabled: fieldName === "username",
            label: translate(`user_management.attributes.${fieldName}.label`, { defaultValue: fieldName }),
            meta: metadata.supported_attributes[fieldName],
            name: fieldPath as Path<UserDetailsExtended>,
            required: metadata.required_attributes.includes(fieldName),
        };
    };

    const shownByDefaultFields = [...basic.map(buildFieldConfig), ...required.map(buildFieldConfig)].filter(Boolean);
    const additionalFields = [...optional.map(buildFieldConfig), ...extra.map(buildFieldConfig)].filter(Boolean);

    const onSubmit = async (data: UserDetailsExtended) => {
        if (!user) return;

        try {
            const updateMask: string[] = [];
            const changedData: Partial<UserDetailsExtended> = {};

            (Object.keys(dirtyFields) as Array<keyof UserDetailsExtended>).forEach((key) => {
                if (key === "extra" && dirtyFields.extra) {
                    Object.keys(dirtyFields.extra || {}).forEach((extraKey) => {
                        updateMask.push(`extra.${extraKey}`);
                    });
                    changedData.extra = data.extra;
                } else if (key === "address" && dirtyFields.address) {
                    Object.keys(dirtyFields.address || {}).forEach((addressKey) => {
                        updateMask.push(`address.${addressKey}`);
                    });
                    changedData.address = data.address;
                } else {
                    updateMask.push(key as string);
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
        } catch {
            createErrorNotification(translate("Error modifying user"));
        }
    };

    const handleClose = () => {
        setVerifyExitDialogOpen(false);
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
        if (user) {
            reset(user);
        }
        handleClose();
    };

    const handleCancelExit = () => {
        setVerifyExitDialogOpen(false);
    };

    return (
        <>
            <Dialog
                open={open}
                onOpenChange={(next) => {
                    if (!next) handleSafeClose();
                }}
            >
                <DialogContent id="edit-user-dialog" className="sm:max-w-lg" showCloseButton={false}>
                    <DialogHeader>
                        <DialogTitle>
                            {translate("Edit {{item}}:", { item: translate("User") })} {user?.username}
                        </DialogTitle>
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
                            {shownByDefaultFields.map((field) =>
                                field!.disabled ? (
                                    <Field key={field!.name}>
                                        <FieldLabel htmlFor={`edit-user-${field!.name}`}>{field!.label}</FieldLabel>
                                        <Input
                                            id={`edit-user-${field!.name}`}
                                            type="text"
                                            disabled
                                            value={(user?.[field!.name as keyof UserDetailsExtended] as string) || ""}
                                        />
                                        {field!.description ? (
                                            <FieldDescription>{field!.description}</FieldDescription>
                                        ) : null}
                                    </Field>
                                ) : (
                                    <UserFormField
                                        key={field!.name}
                                        idPrefix="edit-user"
                                        field={field!}
                                        register={register}
                                        control={control}
                                        errors={errors}
                                        setValue={setValue}
                                        options={groups}
                                    />
                                ),
                            )}

                            {additionalFields.length > 0 && (
                                <Button
                                    id="edit-user-toggle-additional"
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
                                        idPrefix="edit-user"
                                        field={field!}
                                        register={register}
                                        control={control}
                                        errors={errors}
                                        setValue={setValue}
                                    />
                                ))}

                            <DialogFooter>
                                <Button id="edit-user-cancel" type="button" variant="ghost" onClick={handleSafeClose}>
                                    {translate("Cancel")}
                                </Button>
                                <Button id="edit-user-submit" type="submit" color="primary" disabled={!isDirty}>
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

export default EditUserDialog;
