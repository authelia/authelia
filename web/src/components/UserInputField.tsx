import {
    Autocomplete,
    Box,
    Button,
    Checkbox,
    FormControl,
    FormHelperText,
    TextField as MuiTextField,
} from "@mui/material";
import { Control, Controller, FieldErrors, Path, UseFormRegister, UseFormSetValue } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { REGEX } from "@constants/Regex.ts";
import { CreateUserRequest, UserDetailsExtended } from "@root/models/UserManagement";
import { AttributeMetadata } from "@services/UserManagement";

interface UserFormFieldProps<T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest> {
    field: {
        name: Path<T>;
        meta: AttributeMetadata;
        label: string;
        description: string;
        required: boolean;
        disabled?: boolean;
    };
    register: UseFormRegister<T>;
    control: Control<T>;
    errors: FieldErrors<T>;
    setValue: UseFormSetValue<T>;
    options?: string[];
    onGeneratePassword?: () => void;
}

const UserFormField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    errors,
    field,
    onGeneratePassword,
    options,
    register,
}: UserFormFieldProps<T>) => {
    const error = errors[field.name as keyof FieldErrors<T>];

    switch (field.name as string) {
        case "username":
            return <UsernameField field={field} register={register} error={error} control={control} />;
        case "password":
            return (
                <PasswordField
                    field={field}
                    register={register}
                    error={error}
                    onGeneratePassword={onGeneratePassword}
                    control={control}
                />
            );

        case "groups":
            return renderByType(field, register, control, error, options);

        case "mail":
            if (field.meta.multiple) {
                return <MultiEmailField field={field} register={register} control={control} error={error} />;
            }
            return <EmailField field={field} register={register} error={error} />;

        case "birthdate":
            return <DateField field={field} register={register} control={control} error={error} />;

        default:
            return renderByType(field, register, control, error);
    }
};

const renderByType = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>(
    field: UserFormFieldProps<T>["field"],
    register: any,
    control: any,
    error: any,
    options?: string[],
    onGeneratePassword?: () => void,
) => {
    switch (field.meta.type) {
        case "email":
            if (field.meta.multiple) {
                return <MultiEmailField field={field} register={register} control={control} error={error} />;
            }
            return <EmailField field={field} register={register} error={error} />;

        case "tel":
            return <PhoneField field={field} register={register} error={error} />;

        case "url":
            return <UrlField field={field} register={register} error={error} />;

        case "date":
            return <DateField field={field} register={register} control={control} error={error} />;

        case "checkbox":
            return <CheckboxField field={field} register={register} control={control} error={error} />;

        case "number":
            return <NumberField field={field} register={register} error={error} />;

        case "groups":
            return <GroupsField field={field} control={control} error={error} options={options} register={register} />;

        case "password":
            return (
                <PasswordField
                    field={field}
                    register={register}
                    error={error}
                    onGeneratePassword={onGeneratePassword}
                    control={control}
                />
            );

        case "text":
        default:
            if (field.meta.multiple) {
                return <MultiValuedTextField field={field} register={register} error={error} control={control} />;
            }
            return <TextField field={field} register={register} error={error} />;
    }
};

interface FieldComponentProps<T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest> {
    field: UserFormFieldProps<T>["field"];
    register: any;
    error: any;
    control?: any;
    onGeneratePassword?: () => void;
}

const TextField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => (
    <MuiTextField
        fullWidth
        type="text"
        color="info"
        label={field.label}
        helperText={error?.message?.toString() || field.description}
        error={!!error}
        required={field.required}
        {...register(field.name, {
            required: field.required ? `${field.label} is required` : false,
        })}
    />
);

const MultiValuedTextField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
}: FieldComponentProps<T>) => {
    return (
        <Controller
            name={field.name}
            control={control}
            rules={{
                required: field.required ? `${field.label} is required` : false,
            }}
            render={({ field: { onChange, value } }) => (
                <Autocomplete
                    multiple
                    freeSolo
                    options={[]}
                    value={value || []}
                    onChange={(_, newValue) => onChange(newValue)}
                    renderInput={(params) => (
                        <MuiTextField
                            {...params}
                            label={`${field.label} (Press Enter to add multiple)`}
                            placeholder={field.label}
                            helperText={error?.message?.toString() || field.description}
                            error={!!error}
                            required={field.required}
                            color="info"
                        />
                    )}
                />
            )}
        />
    );
};

const PasswordField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
    onGeneratePassword,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");

    //TODO: refactor password field to have password visibility button
    return (
        <>
            <Controller
                name={field.name}
                control={control}
                rules={{
                    required: field.required ? `${field.label} is required` : false,
                }}
                render={({ field: controllerField }) => (
                    <MuiTextField
                        {...controllerField}
                        fullWidth
                        type="password"
                        color="info"
                        label={field.label}
                        helperText={error?.message?.toString() || field.description}
                        error={!!error}
                        required={field.required}
                    />
                )}
            />
            {onGeneratePassword && (
                <Button onClick={onGeneratePassword} color="info" size="small" sx={{ mt: 0.75 }}>
                    {translate("Generate Password")}
                </Button>
            )}
        </>
    );
};

const GroupsField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
    options = [],
    register,
}: FieldComponentProps<T> & { options?: string[] }) => {
    const { t: translate } = useTranslation("settings");
    return (
        <FormControl fullWidth error={!!error} required={field.required}>
            <Controller
                name={field.name}
                control={control}
                rules={{
                    required: field.required ? `${field.label} is required` : false,
                }}
                render={({ field: { onChange, value, ...rest } }) => (
                    <Autocomplete<string, true>
                        multiple
                        disablePortal
                        options={options}
                        value={value || []}
                        onChange={(_, newValue) => onChange(newValue)}
                        renderInput={(params) => (
                            <MuiTextField
                                {...params}
                                label={field.label}
                                error={!!error}
                                helperText={error?.message?.toString() || field.description}
                                required={field.required}
                                color="info"
                                {...register(field.name, {
                                    pattern: {
                                        message: translate(
                                            "Group names must contain only alphanumeric characters, hyphens (-), underscores (_), and commas (,) with a maximum length of 100 characters.",
                                        ),
                                        value: REGEX.GROUP,
                                    },
                                    required: field.required ? `${field.label} is required` : false,
                                })}
                            />
                        )}
                        {...rest}
                    />
                )}
            />
        </FormControl>
    );
};

const CheckboxField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
}: FieldComponentProps<T>) => (
    <FormControl error={!!error} required={field.required}>
        <Box sx={{ alignItems: "center", display: "flex" }}>
            <Controller
                name={field.name}
                control={control}
                rules={{
                    required: field.required ? `${field.label} is required` : false,
                }}
                render={({ field: { onChange, value, ...rest } }) => (
                    <Checkbox color="info" checked={!!value} onChange={(e) => onChange(e.target.checked)} {...rest} />
                )}
            />
            <Box component="label" sx={{ cursor: "pointer" }}>
                {field.label}
            </Box>
        </Box>
        {(error || field.description) && (
            <FormHelperText>{error?.message?.toString() || field.description}</FormHelperText>
        )}
    </FormControl>
);

const EmailField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");

    return (
        <MuiTextField
            fullWidth
            type="email"
            color="info"
            label={field.label}
            helperText={error?.message?.toString() || field.description}
            error={!!error}
            required={field.required}
            {...register(field.name, {
                pattern: {
                    message: translate("Invalid email address"),
                    value: REGEX.EMAIL,
                },
                required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
            })}
        />
    );
};

const UsernameField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    return (
        <MuiTextField
            fullWidth
            type="text"
            color="info"
            label={field.label}
            helperText={error?.message?.toString() || field.description}
            error={!!error}
            required={field.required}
            {...register(field.name, {
                pattern: {
                    message: translate(
                        "Usernames must contain only alphanumeric characters, hyphens (-), underscores (_), and commas (,) with a maximum length of 100 characters.",
                    ),
                    value: REGEX.USERNAME,
                },
                required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
            })}
        />
    );
};

const NumberField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");

    return (
        <MuiTextField
            fullWidth
            type="number"
            color="info"
            label={field.label}
            helperText={error?.message?.toString() || field.description}
            error={!!error}
            required={field.required}
            {...register(field.name, {
                required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                valueAsNumber: true,
            })}
        />
    );
};

const PhoneField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");

    return (
        <MuiTextField
            fullWidth
            type="tel"
            color="info"
            label={field.label}
            helperText={error?.message?.toString() || field.description}
            error={!!error}
            required={field.required}
            {...register(field.name, {
                pattern: {
                    message: translate("Invalid phone number"),
                    value: REGEX.TELEPHONE_NUMBER,
                },
                required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
            })}
        />
    );
};

const UrlField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");

    return (
        <MuiTextField
            fullWidth
            type="url"
            color="info"
            label={field.label}
            helperText={error?.message?.toString() || field.description}
            error={!!error}
            required={field.required}
            {...register(field.name, {
                required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                validate: (value: string) => {
                    if (!value) return true;
                    try {
                        new URL(value);
                        return true;
                    } catch {
                        return "Invalid URL";
                    }
                },
            })}
        />
    );
};

const DateField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");

    return (
        <MuiTextField
            fullWidth
            type="date"
            color="info"
            label={field.label}
            helperText={error?.message?.toString() || field.description}
            error={!!error}
            required={field.required}
            slotProps={{
                inputLabel: {
                    shrink: true,
                },
            }}
            {...register(field.name, {
                required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
            })}
        />
    );
};

const MultiEmailField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");

    //TODO: implement actual component
    return (
        <MuiTextField
            fullWidth
            color="info"
            type="email"
            label={field.label + " (comma-separated)"}
            helperText={error?.message?.toString() || field.description}
            error={!!error}
            required={field.required}
            placeholder="email1@example.com, email2@example.com"
            {...register(field.name, {
                required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                validate: (value: string) => {
                    if (!value) return true;
                    const emails = value.split(",").map((e) => e.trim());
                    const emailRegex = REGEX.EMAIL;
                    const allValid = emails.every((email) => emailRegex.test(email));
                    return allValid || "One or more invalid email addresses";
                },
            })}
        />
    );
};

export default UserFormField;
