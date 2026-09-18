import { useState } from "react";

import { EyeIcon, EyeOffIcon } from "lucide-react";
import { Control, Controller, FieldErrors, Path, UseFormRegister, UseFormSetValue } from "react-hook-form";
import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Checkbox } from "@components/UI/Checkbox";
import { Combobox } from "@components/UI/Combobox";
import { Field, FieldDescription, FieldError, FieldLabel } from "@components/UI/Field";
import { Input } from "@components/UI/Input";
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "@components/UI/InputGroup";
import { Label } from "@components/UI/Label";
import { REGEX } from "@constants/Regex";
import { CreateUserRequest, UserDetailsExtended, ValidateUsername } from "@models/UserManagement";
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
    idPrefix?: string;
}

const fieldId = (idPrefix: string | undefined, name: string) =>
    idPrefix ? `${idPrefix}-${name.replaceAll(".", "-")}` : undefined;

const UserFormField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    errors,
    field,
    idPrefix,
    onGeneratePassword,
    options,
    register,
}: UserFormFieldProps<T>) => {
    const error = errors[field.name as keyof FieldErrors<T>];

    switch (field.name as string) {
        case "username":
            return (
                <UsernameField field={field} idPrefix={idPrefix} register={register} error={error} control={control} />
            );
        case "password":
            return (
                <PasswordField
                    field={field}
                    idPrefix={idPrefix}
                    register={register}
                    error={error}
                    onGeneratePassword={onGeneratePassword}
                    control={control}
                />
            );

        case "groups":
            return renderByType(field, register, control, error, options, onGeneratePassword, idPrefix);

        case "mail":
            if (field.meta.multiple) {
                return (
                    <MultiEmailField
                        field={field}
                        idPrefix={idPrefix}
                        register={register}
                        control={control}
                        error={error}
                    />
                );
            }
            return <EmailField field={field} idPrefix={idPrefix} register={register} error={error} />;

        case "birthdate":
            return <DateField field={field} idPrefix={idPrefix} register={register} control={control} error={error} />;

        default:
            return renderByType(field, register, control, error, options, onGeneratePassword, idPrefix);
    }
};

const renderByType = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>(
    field: UserFormFieldProps<T>["field"],
    register: any,
    control: any,
    error: any,
    options?: string[],
    onGeneratePassword?: () => void,
    idPrefix?: string,
) => {
    switch (field.meta.type) {
        case "email":
            if (field.meta.multiple) {
                return (
                    <MultiEmailField
                        field={field}
                        idPrefix={idPrefix}
                        register={register}
                        control={control}
                        error={error}
                    />
                );
            }
            return <EmailField field={field} idPrefix={idPrefix} register={register} error={error} />;

        case "tel":
            return <PhoneField field={field} idPrefix={idPrefix} register={register} error={error} />;

        case "url":
            return <UrlField field={field} idPrefix={idPrefix} register={register} error={error} />;

        case "date":
            return <DateField field={field} idPrefix={idPrefix} register={register} control={control} error={error} />;

        case "checkbox":
            return (
                <CheckboxField field={field} idPrefix={idPrefix} register={register} control={control} error={error} />
            );

        case "number":
            return <NumberField field={field} idPrefix={idPrefix} register={register} error={error} />;

        case "groups":
            return (
                <GroupsField
                    field={field}
                    idPrefix={idPrefix}
                    control={control}
                    error={error}
                    options={options}
                    register={register}
                />
            );

        case "password":
            return (
                <PasswordField
                    field={field}
                    idPrefix={idPrefix}
                    register={register}
                    error={error}
                    onGeneratePassword={onGeneratePassword}
                    control={control}
                />
            );

        case "text":
        default:
            if (field.meta.multiple) {
                return (
                    <MultiValuedTextField
                        field={field}
                        idPrefix={idPrefix}
                        register={register}
                        error={error}
                        control={control}
                    />
                );
            }
            return <TextField field={field} idPrefix={idPrefix} register={register} error={error} />;
    }
};

interface FieldComponentProps<T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest> {
    field: UserFormFieldProps<T>["field"];
    register: any;
    error: any;
    control?: any;
    onGeneratePassword?: () => void;
    idPrefix?: string;
}

const TextField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    idPrefix,
    register,
}: FieldComponentProps<T>) => {
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Input
                id={id}
                type="text"
                error={!!error}
                required={field.required}
                {...register(field.name, {
                    required: field.required ? `${field.label} is required` : false,
                })}
            />
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const MultiValuedTextField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
    idPrefix,
}: FieldComponentProps<T>) => {
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{`${field.label} (Press Enter to add multiple)`}</FieldLabel>
            <Controller
                name={field.name}
                control={control}
                defaultValue={[] as any}
                rules={{
                    required: field.required ? `${field.label} is required` : false,
                }}
                render={({ field: { onChange, value } }) => (
                    <Combobox
                        id={id}
                        freeSolo
                        options={[]}
                        value={Array.isArray(value) ? value : []}
                        onChange={onChange}
                        error={!!error}
                        placeholder={field.label}
                    />
                )}
            />
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const PasswordField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
    idPrefix,
    onGeneratePassword,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const [visible, setVisible] = useState(false);
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Controller
                name={field.name}
                control={control}
                rules={{
                    required: field.required ? `${field.label} is required` : false,
                }}
                render={({ field: controllerField }) => (
                    <InputGroup>
                        <InputGroupInput
                            {...controllerField}
                            id={id}
                            type={visible ? "text" : "password"}
                            error={!!error}
                            required={field.required}
                        />
                        <InputGroupAddon align="inline-end">
                            <InputGroupButton
                                aria-label={visible ? translate("Hide Password") : translate("Show Password")}
                                onClick={() => setVisible((prev) => !prev)}
                                type="button"
                            >
                                {visible ? <EyeOffIcon /> : <EyeIcon />}
                            </InputGroupButton>
                        </InputGroupAddon>
                    </InputGroup>
                )}
            />
            {onGeneratePassword && (
                <Button
                    id={idPrefix ? `${idPrefix}-generate-password` : undefined}
                    onClick={onGeneratePassword}
                    variant="ghost"
                    size="sm"
                    type="button"
                >
                    {translate("Generate Password")}
                </Button>
            )}
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const GroupsField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
    idPrefix,
    options = [],
}: FieldComponentProps<T> & { options?: string[] }) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Controller
                name={field.name}
                control={control}
                defaultValue={[] as any}
                rules={{
                    required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                }}
                render={({ field: { onChange, value } }) => (
                    <Combobox
                        id={id}
                        options={options}
                        value={Array.isArray(value) ? value : []}
                        onChange={onChange}
                        error={!!error}
                        placeholder={field.label}
                    />
                )}
            />
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const CheckboxField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
    idPrefix,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <div className="flex flex-col gap-1.5">
            <div className="flex items-center gap-2">
                <Controller
                    name={field.name}
                    control={control}
                    rules={{
                        required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                    }}
                    render={({ field: { onChange, value, ...rest } }) => (
                        <Checkbox
                            id={id}
                            checked={!!value}
                            onCheckedChange={(checked) => onChange(!!checked)}
                            {...rest}
                        />
                    )}
                />
                <Label htmlFor={id}>{field.label}</Label>
            </div>
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : field.description ? (
                <FieldDescription>{field.description}</FieldDescription>
            ) : null}
        </div>
    );
};

const EmailField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    idPrefix,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Input
                id={id}
                type="email"
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
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const UsernameField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    idPrefix,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Input
                id={id}
                type="text"
                error={!!error}
                required={field.required}
                {...register(field.name, {
                    required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                    validate: (value: string) => {
                        if (!value) return true;

                        return (
                            ValidateUsername(value) ||
                            translate(
                                "Usernames must contain only alphanumeric characters, hyphens (-), underscores (_), and commas (,) with a maximum length of 100 characters.",
                            )
                        );
                    },
                })}
            />
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const NumberField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    idPrefix,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Input
                id={id}
                type="number"
                error={!!error}
                required={field.required}
                {...register(field.name, {
                    required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                    valueAsNumber: true,
                })}
            />
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const PhoneField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    idPrefix,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Input
                id={id}
                type="tel"
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
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const UrlField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    idPrefix,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Input
                id={id}
                type="url"
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
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const DateField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    error,
    field,
    idPrefix,
    register,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{field.label}</FieldLabel>
            <Input
                id={id}
                type="date"
                error={!!error}
                required={field.required}
                {...register(field.name, {
                    required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                })}
            />
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

const MultiEmailField = <T extends CreateUserRequest | UserDetailsExtended = CreateUserRequest>({
    control,
    error,
    field,
    idPrefix,
}: FieldComponentProps<T>) => {
    const { t: translate } = useTranslation("settings");
    const id = fieldId(idPrefix, field.name);

    return (
        <Field data-invalid={!!error}>
            <FieldLabel htmlFor={id}>{`${field.label} (comma-separated)`}</FieldLabel>
            <Controller
                name={field.name}
                control={control}
                defaultValue={[] as any}
                rules={{
                    required: field.required ? translate("{{item}} is required", { item: field.label }) : false,
                    validate: (value: string[]) => {
                        if (!value || value.length === 0) return true;

                        const emails = Array.isArray(value) ? value : [value];
                        const allValid = emails.every((email) => typeof email === "string" && REGEX.EMAIL.test(email));

                        return allValid || translate("One or more invalid email addresses");
                    },
                }}
                render={({ field: { onChange, value } }) => (
                    <Combobox
                        id={id}
                        freeSolo
                        options={[]}
                        value={Array.isArray(value) ? value : []}
                        onChange={onChange}
                        error={!!error}
                        placeholder="email1@example.com, email2@example.com"
                    />
                )}
            />
            {error ? (
                <FieldError>{error?.message?.toString()}</FieldError>
            ) : (
                <FieldDescription>{field.description}</FieldDescription>
            )}
        </Field>
    );
};

export default UserFormField;
