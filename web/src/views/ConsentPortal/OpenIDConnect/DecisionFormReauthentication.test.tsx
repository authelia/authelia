import { createRef } from "react";

import { fireEvent, render, screen } from "@testing-library/react";

import { IsCapsLockModified } from "@services/CapsLock";
import DecisionFormReauthentication from "@views/ConsentPortal/OpenIDConnect/DecisionFormReauthentication";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: vi.fn(() => null),
}));

beforeEach(() => {
    vi.mocked(IsCapsLockModified).mockReturnValue(null);
});

it("renders a password input", () => {
    render(<DecisionFormReauthentication value={""} error={false} disabled={false} onChange={vi.fn()} />);

    const input = screen.getByLabelText("Password");

    expect(input).toHaveAttribute("type", "password");
    expect(input).toHaveAttribute("name", "password");
});

it("forwards the ref to the password input", () => {
    const ref = createRef<HTMLInputElement>();

    render(<DecisionFormReauthentication ref={ref} value={""} error={false} disabled={false} onChange={vi.fn()} />);

    expect(ref.current).toBe(screen.getByLabelText("Password"));
});

it("reports typed characters to the change handler", () => {
    const onChange = vi.fn();

    render(<DecisionFormReauthentication value={""} error={false} disabled={false} onChange={onChange} />);

    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "a" } });

    expect(onChange).toHaveBeenCalledWith("a");
});

it("disables the password input while submitting", () => {
    render(<DecisionFormReauthentication value={"secret"} error={false} disabled onChange={vi.fn()} />);

    expect(screen.getByLabelText("Password")).toBeDisabled();
});

it("reveals the password while the toggle is held", () => {
    render(<DecisionFormReauthentication value={"secret"} error={false} disabled={false} onChange={vi.fn()} />);

    const toggle = screen.getByRole("button", { name: "Toggle password visibility" });

    fireEvent.mouseDown(toggle);
    expect(screen.getByLabelText("Password")).toHaveAttribute("type", "text");

    fireEvent.mouseUp(toggle);
    expect(screen.getByLabelText("Password")).toHaveAttribute("type", "password");
});

it("hides the password again when the pointer leaves the toggle", () => {
    render(<DecisionFormReauthentication value={"secret"} error={false} disabled={false} onChange={vi.fn()} />);

    const toggle = screen.getByRole("button", { name: "Toggle password visibility" });

    fireEvent.mouseDown(toggle);
    fireEvent.mouseLeave(toggle);

    expect(screen.getByLabelText("Password")).toHaveAttribute("type", "password");
});

it("reveals the password while the toggle is held with the keyboard", () => {
    render(<DecisionFormReauthentication value={"secret"} error={false} disabled={false} onChange={vi.fn()} />);

    const toggle = screen.getByRole("button", { name: "Toggle password visibility" });

    fireEvent.keyDown(toggle, { key: " " });
    expect(screen.getByLabelText("Password")).toHaveAttribute("type", "text");

    fireEvent.keyUp(toggle, { key: " " });
    expect(screen.getByLabelText("Password")).toHaveAttribute("type", "password");
});

it("reveals the password while the toggle is held on touch devices", () => {
    render(<DecisionFormReauthentication value={"secret"} error={false} disabled={false} onChange={vi.fn()} />);

    const toggle = screen.getByRole("button", { name: "Toggle password visibility" });

    fireEvent.touchStart(toggle);
    expect(screen.getByLabelText("Password")).toHaveAttribute("type", "text");

    fireEvent.touchEnd(toggle);
    expect(screen.getByLabelText("Password")).toHaveAttribute("type", "password");
});

it("warns when the password was entered with caps lock", () => {
    vi.mocked(IsCapsLockModified).mockReturnValue(true);

    render(<DecisionFormReauthentication value={"SECRET"} error={false} disabled={false} onChange={vi.fn()} />);

    fireEvent.keyUp(screen.getByLabelText("Password"), { key: "T" });

    expect(screen.getByText("The password was entered with Caps Lock")).toBeInTheDocument();
});

it("warns when the password was partially entered with caps lock", () => {
    vi.mocked(IsCapsLockModified).mockReturnValue(true);

    const { rerender } = render(
        <DecisionFormReauthentication value={"SECRET"} error={false} disabled={false} onChange={vi.fn()} />,
    );

    fireEvent.keyUp(screen.getByLabelText("Password"), { key: "T" });

    vi.mocked(IsCapsLockModified).mockReturnValue(false);

    rerender(<DecisionFormReauthentication value={"SECRETs"} error={false} disabled={false} onChange={vi.fn()} />);
    fireEvent.keyUp(screen.getByLabelText("Password"), { key: "s" });

    expect(screen.getByText("The password was partially entered with Caps Lock")).toBeInTheDocument();
});

it("does not warn when the caps lock state cannot be determined", () => {
    vi.mocked(IsCapsLockModified).mockReturnValue(null);

    render(<DecisionFormReauthentication value={"secret"} error={false} disabled={false} onChange={vi.fn()} />);

    fireEvent.keyUp(screen.getByLabelText("Password"), { key: "t" });

    expect(screen.queryByText("The password was entered with Caps Lock")).not.toBeInTheDocument();
});

it("clears the caps lock warning once the password is emptied", () => {
    vi.mocked(IsCapsLockModified).mockReturnValue(true);

    const { rerender } = render(
        <DecisionFormReauthentication value={"SECRET"} error={false} disabled={false} onChange={vi.fn()} />,
    );

    fireEvent.keyUp(screen.getByLabelText("Password"), { key: "T" });
    expect(screen.getByText("The password was entered with Caps Lock")).toBeInTheDocument();

    rerender(<DecisionFormReauthentication value={""} error={false} disabled={false} onChange={vi.fn()} />);
    fireEvent.keyUp(screen.getByLabelText("Password"), { key: "Backspace" });

    expect(screen.queryByText("The password was entered with Caps Lock")).not.toBeInTheDocument();
});

it("clears the caps lock warning when the password is reset programmatically", () => {
    vi.mocked(IsCapsLockModified).mockReturnValue(true);

    const { rerender } = render(
        <DecisionFormReauthentication value={"SECRET"} error={false} disabled={false} onChange={vi.fn()} />,
    );

    fireEvent.keyUp(screen.getByLabelText("Password"), { key: "T" });
    expect(screen.getByText("The password was entered with Caps Lock")).toBeInTheDocument();

    rerender(<DecisionFormReauthentication value={""} error={false} disabled={false} onChange={vi.fn()} />);

    expect(screen.queryByText("The password was entered with Caps Lock")).not.toBeInTheDocument();
});

it("does not render a failure alert by default", () => {
    render(<DecisionFormReauthentication value={""} error={false} disabled={false} onChange={vi.fn()} />);

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
});

it("renders the failure alert when an attempt failed", () => {
    render(
        <DecisionFormReauthentication
            value={""}
            error={false}
            disabled={false}
            failure={"Incorrect password"}
            onChange={vi.fn()}
        />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent("Incorrect password");
});

it("renders the failure alert above the password field", () => {
    const { container } = render(
        <DecisionFormReauthentication
            value={""}
            error={false}
            disabled={false}
            failure={"Incorrect password"}
            onChange={vi.fn()}
        />,
    );

    const nodes = Array.from(container.querySelectorAll("[role='alert'], input"));

    expect(nodes[0]).toHaveAttribute("role", "alert");
});

it("marks the password field as invalid when an attempt failed", () => {
    render(
        <DecisionFormReauthentication
            value={""}
            error={false}
            disabled={false}
            failure={"Incorrect password"}
            onChange={vi.fn()}
        />,
    );

    expect(screen.getByLabelText("Password")).toHaveAttribute("aria-invalid", "true");
});

it("associates the failure message with the password field", () => {
    render(
        <DecisionFormReauthentication
            value={""}
            error={false}
            disabled={false}
            failure={"Incorrect password"}
            onChange={vi.fn()}
        />,
    );

    expect(screen.getByLabelText("Password")).toHaveAccessibleDescription("Incorrect password");
});

it("leaves the password field undescribed when there is no failure", () => {
    render(<DecisionFormReauthentication value={""} error={false} disabled={false} onChange={vi.fn()} />);

    expect(screen.getByLabelText("Password")).not.toHaveAccessibleDescription();
});

it("places the reveal toggle inside the input group", () => {
    const { container } = render(
        <DecisionFormReauthentication value={"secret"} error={false} disabled={false} onChange={vi.fn()} />,
    );

    const group = container.querySelector("[data-slot='input-group']");

    expect(group).toBeInTheDocument();
    expect(group).toContainElement(screen.getByRole("button", { name: "Toggle password visibility" }));
    expect(group).toContainElement(screen.getByLabelText("Password"));
});

it("marks the surrounding field invalid when an attempt failed", () => {
    const { container } = render(
        <DecisionFormReauthentication
            value={""}
            error={false}
            disabled={false}
            failure={"Incorrect password"}
            onChange={vi.fn()}
        />,
    );

    expect(container.querySelector("[data-slot='field']")).toHaveAttribute("data-invalid", "true");
});
