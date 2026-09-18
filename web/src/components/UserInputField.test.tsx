import { act, fireEvent, render, screen } from "@testing-library/react";
import { useForm } from "react-hook-form";

import UserFormField from "@components/UserInputField";
import { AttributeMetadata } from "@services/UserManagement";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string, opts?: any) => (opts?.item ? `${key}:${opts.item}` : key) }),
}));

interface HarnessProps {
    name: string;
    meta: AttributeMetadata;
    label?: string;
    description?: string;
    required?: boolean;
    options?: string[];
    onGeneratePassword?: () => void;
    onSubmit?: (data: any) => void;
    defaultValues?: Record<string, any>;
}

const Harness = (props: HarnessProps) => {
    const {
        control,
        formState: { errors },
        handleSubmit,
        register,
        setValue,
    } = useForm<any>({ defaultValues: props.defaultValues ?? {} });

    return (
        <form onSubmit={handleSubmit(props.onSubmit ?? (() => {}))}>
            <UserFormField
                idPrefix="t"
                field={{
                    description: props.description ?? "",
                    label: props.label ?? props.name,
                    meta: props.meta,
                    name: props.name,
                    required: props.required ?? false,
                }}
                register={register}
                control={control}
                errors={errors}
                setValue={setValue}
                options={props.options}
                onGeneratePassword={props.onGeneratePassword}
            />
            <button type="submit">submit</button>
        </form>
    );
};

const byId = (id: string) => document.getElementById(id) as HTMLInputElement;

// Dispatches submit directly so the browser's native constraint validation (required, type=email/url)
// does not swallow the event; this exercises the react-hook-form rules themselves.
const submitForm = async () => {
    await act(async () => {
        fireEvent.submit(document.querySelector("form")!);
    });
};

// Clicks the submit button so native constraint validation applies, as it would for a real user.
const clickSubmit = async () => {
    await act(async () => {
        fireEvent.click(screen.getByText("submit"));
    });
};

it("renders a text input for a plain text attribute", () => {
    render(<Harness name="nickname" meta={{ type: "text" }} label="Nickname" description="Shown publicly" />);

    expect(byId("t-nickname")).toHaveAttribute("type", "text");
    expect(screen.getByText("Shown publicly")).toBeInTheDocument();
});

it("renders the username field with username validation regardless of type", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="username" meta={{ type: "text" }} onSubmit={onSubmit} />);

    fireEvent.change(byId("t-username"), { target: { value: "bad user!" } });
    await submitForm();

    expect(screen.getByText(/Usernames must contain only alphanumeric characters/)).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();

    fireEvent.change(byId("t-username"), { target: { value: "good_user-1" } });
    await submitForm();

    expect(onSubmit).toHaveBeenCalledWith({ username: "good_user-1" }, expect.anything());
});

it("renders a password input and only shows the generate button when a callback is supplied", () => {
    const { unmount } = render(<Harness name="password" meta={{ type: "password" }} />);

    expect(byId("t-password")).toHaveAttribute("type", "password");
    expect(screen.queryByText("Generate Password")).not.toBeInTheDocument();

    unmount();

    const onGeneratePassword = vi.fn();
    render(<Harness name="password" meta={{ type: "password" }} onGeneratePassword={onGeneratePassword} />);

    fireEvent.click(screen.getByText("Generate Password"));

    expect(onGeneratePassword).toHaveBeenCalledOnce();
    expect(byId("t-generate-password")).toBeInTheDocument();
});

it("renders the groups field as a multi-select over the supplied options", () => {
    render(<Harness name="groups" meta={{ multiple: true, type: "groups" }} options={["admins", "dev"]} />);

    const combobox = screen.getByRole("combobox");
    expect(combobox).toBeInTheDocument();

    fireEvent.mouseDown(combobox);

    expect(screen.getByText("admins")).toBeInTheDocument();
    expect(screen.getByText("dev")).toBeInTheDocument();
});

it("validates a single email attribute", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="mail" meta={{ type: "email" }} onSubmit={onSubmit} />);

    expect(byId("t-mail")).toHaveAttribute("type", "email");

    fireEvent.change(byId("t-mail"), { target: { value: "nope" } });
    await submitForm();

    expect(screen.getByText("Invalid email address")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
});

it("validates a comma separated multi email attribute", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="mail" meta={{ multiple: true, type: "email" }} label="Email" onSubmit={onSubmit} />);

    expect(screen.getByLabelText(/Email \(comma-separated\)/)).toBeInTheDocument();

    const input = byId("t-mail");

    fireEvent.change(input, { target: { value: "a@b.co" } });
    fireEvent.keyDown(input, { key: "," });
    fireEvent.change(input, { target: { value: "nope" } });
    fireEvent.keyDown(input, { key: "Enter" });
    await submitForm();

    expect(screen.getByText("One or more invalid email addresses")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();

    fireEvent.click(screen.getByLabelText("Remove nope"));
    fireEvent.change(input, { target: { value: "c@d.co" } });
    fireEvent.keyDown(input, { key: "Enter" });
    await submitForm();

    expect(onSubmit).toHaveBeenCalledWith({ mail: ["a@b.co", "c@d.co"] }, expect.anything());
});

it("renders a date input for the birthdate attribute", () => {
    render(<Harness name="birthdate" meta={{ type: "text" }} />);

    expect(byId("t-birthdate")).toHaveAttribute("type", "date");
});

it("validates telephone attributes", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="phone_number" meta={{ type: "tel" }} onSubmit={onSubmit} />);

    expect(byId("t-phone_number")).toHaveAttribute("type", "tel");

    fireEvent.change(byId("t-phone_number"), { target: { value: "abc" } });
    await submitForm();

    expect(screen.getByText("Invalid phone number")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
});

it("validates url attributes", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="website" meta={{ type: "url" }} onSubmit={onSubmit} />);

    fireEvent.change(byId("t-website"), { target: { value: "not a url" } });
    await submitForm();

    expect(screen.getByText("Invalid URL")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();

    fireEvent.change(byId("t-website"), { target: { value: "https://example.com" } });
    await submitForm();

    expect(onSubmit).toHaveBeenCalledOnce();
});

it("renders a number input that submits a number", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="age" meta={{ type: "number" }} onSubmit={onSubmit} />);

    fireEvent.change(byId("t-age"), { target: { value: "42" } });
    await submitForm();

    expect(onSubmit).toHaveBeenCalledWith({ age: 42 }, expect.anything());
});

it("renders a checkbox that submits a boolean", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="active" meta={{ type: "checkbox" }} label="Active" onSubmit={onSubmit} />);

    expect(screen.getByText("Active")).toBeInTheDocument();

    fireEvent.click(byId("t-active"));
    await submitForm();

    expect(onSubmit).toHaveBeenCalledWith({ active: true }, expect.anything());
});

it("renders a free text multi-value input for multiple text attributes", () => {
    render(<Harness name="aliases" meta={{ multiple: true, type: "text" }} label="Aliases" />);

    expect(screen.getByRole("combobox")).toBeInTheDocument();
    expect(screen.getByLabelText(/Aliases \(Press Enter to add multiple\)/)).toBeInTheDocument();
});

it("reports required fields on submit", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="given_name" meta={{ type: "text" }} label="Given Name" required onSubmit={onSubmit} />);

    await submitForm();

    expect(screen.getByText("Given Name is required")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
});

it("reports required translated fields on submit", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="mail" meta={{ type: "email" }} label="Email" required onSubmit={onSubmit} />);

    await submitForm();

    expect(screen.getByText("{{item}} is required:Email")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
});

it("lets native constraint validation block submission of an empty required field", async () => {
    const onSubmit = vi.fn();
    render(<Harness name="given_name" meta={{ type: "text" }} label="Given Name" required onSubmit={onSubmit} />);

    expect(byId("t-given_name")).toBeRequired();

    await clickSubmit();

    expect(onSubmit).not.toHaveBeenCalled();
});

it("accepts an array default value for a multi email attribute", async () => {
    const onSubmit = vi.fn();
    render(
        <Harness
            name="mail"
            meta={{ multiple: true, type: "email" }}
            defaultValues={{ mail: ["a@b.co", "c@d.co"] }}
            onSubmit={onSubmit}
        />,
    );

    await submitForm();

    expect(onSubmit).toHaveBeenCalledOnce();
});
