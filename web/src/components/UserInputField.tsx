import {
    Autocomplete,
    Box,
    Button,
    Checkbox,
    FormControl,
    FormHelperText,
    TextField as MuiTextField,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { AttributeMetadata } from "@services/UserManagement";
import { Control, Controller, FieldErrors, Path, UseFormRegister, UseFormSetValue } from "react-hook-form";
import { CreateUserRequest, UserDetailsExtended } from "@root/models/UserManagement";

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
    field,
    register,
    control,
    errors,
    options,
    onGeneratePassword
}: UserFormFieldProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const error = errors[field.name as keyof FieldErrors<T>];

    switch (field.name as string) {
        case "password":
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

        case "groups":
            return (
                <GroupsField
                    field={field}
                    control={control}
                    error={error}
                    options={options}
                    register={register}
                />
            );

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
            return <CheckboxField field={field} register={register} control={control} error={error} />

        case "groups":
            return <GroupsField field={field} control={control} error={error} options={options} register={register} />;

        case "text":
        case "password":
        default:
            return <TextField field={field} register={register} error={error} />;
    }
};


interface FieldComponentProps<T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest> {
    field: UserFormFieldProps<T>["field"];
    register: any;
    control?: any;
    error: any;
}

const TextField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    field,
    register,
    error
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

interface FieldComponentPropsWithOptions<T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest> extends FieldComponentProps<T> {
    options?: string[] | { label: string; value: string }[];
}

const GroupsField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
                                                                                                field,
                                                                                                control,
                                                                                                error,
                                                                                                options = []
                                                                                            }: FieldComponentProps<T> & { options?: string[] }) => (
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
                        />
                    )}
                    {...rest}
                />
            )}
        />
    </FormControl>
);

const CheckboxField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
                                                                                                  field,
                                                                                                  control,
                                                                                                  error
                                                                                              }: FieldComponentProps<T>) => (
    <FormControl error={!!error} required={field.required}>
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Controller
                name={field.name}
                control={control}
                rules={{
                    required: field.required ? `${field.label} is required` : false,
                }}
                render={({ field: { onChange, value, ...rest } }) => (
                    <Checkbox
                        color="info"
                        checked={!!value}
                        onChange={(e) => onChange(e.target.checked)}
                        {...rest}
                    />
                )}
            />
            <Box component="label" sx={{ cursor: 'pointer' }}>
                {field.label}
            </Box>
        </Box>
        {(error || field.description) && (
            <FormHelperText>{error?.message?.toString() || field.description}</FormHelperText>
        )}
    </FormControl>
);

const EmailField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    field,
    register,
    error
}: FieldComponentProps<T>) => (
    <MuiTextField
        fullWidth
        type="email"
        color="info"
        label={field.label}
        helperText={error?.message?.toString() || field.description}
        error={!!error}
        required={field.required}
        {...register(field.name, {
            required: field.required ? `${field.label} is required` : false,
            pattern: {
                value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                message: "Invalid email address",
            },
        })}
    />
);

const PhoneField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    field,
    register,
    error
}: FieldComponentProps<T>) => (
    <MuiTextField
        fullWidth
        type="tel"
        color="info"
        label={field.label}
        helperText={error?.message?.toString() || field.description}
        error={!!error}
        required={field.required}
        {...register(field.name, {
            required: field.required ? `${field.label} is required` : false,
            pattern: {
                value: /^[+]?[(]?[0-9]{1,4}[)]?[-\s.]?[(]?[0-9]{1,4}[)]?[-\s.]?[0-9]{1,9}$/,
                message: "Invalid phone number",
            },
        })}
    />
);

const UrlField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    field,
    register,
    error
}: FieldComponentProps<T>) => (
    <MuiTextField
        fullWidth
        type="url"
        color="info"
        label={field.label}
        helperText={error?.message?.toString() || field.description}
        error={!!error}
        required={field.required}
        {...register(field.name, {
            required: field.required ? `${field.label} is required` : false,
            validate: (value: string) => {
                if (!value) return true;
                try {
                    new URL(value);
                    return true;
                } catch {
                    return "Invalid URL format";
                }
            },
        })}
    />
);

const DateField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    field,
    register,
    control,
    error
}: FieldComponentProps<T>) => {
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
                }
            }}
            {...register(field.name, {
                required: field.required ? `${field.label} is required` : false,
            })}
        />
    );
};

const MultiEmailField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    field,
    register,
    control,
    error
}: FieldComponentProps<T>) => {
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
                required: field.required ? `${field.label} is required` : false,
                validate: (value: string) => {
                    if (!value) return true;
                    const emails = value.split(",").map(e => e.trim());
                    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                    const allValid = emails.every(email => emailRegex.test(email));
                    return allValid || "One or more invalid email addresses";
                },
            })}
        />
    );
};

export default UserFormField;
