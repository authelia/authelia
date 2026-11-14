// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { signInWithRecoveryCode } from "@services/UserRecoveryCodes";
import RecoveryCodeMethod from "@views/LoginPortal/SecondFactor/RecoveryCodeMethod";

const mocks = vi.hoisted(() => ({
    queryParams: {} as Record<string, null | string>,
    userCode: undefined as string | undefined,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: (name: string) => mocks.queryParams[name] ?? null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: "flow", id: "flow-id", subflow: "subflow" }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => mocks.userCode,
}));

vi.mock("@services/UserRecoveryCodes", () => ({
    signInWithRecoveryCode: vi.fn(),
}));

const signInMock = vi.mocked(signInWithRecoveryCode);

function renderMethod() {
    const props = {
        id: "recovery-code-method",
        onCancel: vi.fn(),
        onSignInError: vi.fn(),
        onSignInSuccess: vi.fn(),
    };

    return { ...render(<RecoveryCodeMethod {...props} />), props };
}

function getInput() {
    return screen.getByLabelText("Recovery code") as HTMLInputElement;
}

function getSubmit() {
    return document.getElementById("recovery-code-submit") as HTMLButtonElement;
}

function getCancel() {
    return document.getElementById("recovery-code-cancel") as HTMLButtonElement;
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.queryParams = {};
    mocks.userCode = undefined;
    signInMock.mockResolvedValue({ redirect: "https://example.com" } as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders the input with the provided id", () => {
    renderMethod();

    expect(document.getElementById("recovery-code-method")).toBeInTheDocument();
    expect(getInput()).toHaveValue("");
    expect(getInput()).toHaveFocus();
});

it("disables submission until a code is entered", () => {
    renderMethod();

    expect(getSubmit()).toBeDisabled();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });

    expect(getSubmit()).not.toBeDisabled();
});

it("signs in with the code and forwards the redirect", async () => {
    mocks.queryParams = { rd: "https://target.example.com" };
    mocks.userCode = "user-code";

    const { props } = renderMethod();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });
    fireEvent.click(getSubmit());

    await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalledWith("https://example.com"));
    expect(signInMock).toHaveBeenCalledWith({
        code: "AAAA-BBBB",
        flow: "flow",
        flowID: "flow-id",
        subflow: "subflow",
        targetURL: "https://target.example.com",
        userCode: "user-code",
    });
    expect(props.onSignInError).not.toHaveBeenCalled();
});

it("omits the target URL when no redirection is requested", async () => {
    const { props } = renderMethod();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });
    fireEvent.click(getSubmit());

    await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalled());
    expect(signInMock).toHaveBeenCalledWith(expect.objectContaining({ targetURL: undefined }));
});

it("forwards an undefined redirect when the response is empty", async () => {
    signInMock.mockResolvedValue(undefined);

    const { props } = renderMethod();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });
    fireEvent.click(getSubmit());

    await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalledWith(undefined));
});

it("reports a failed sign in", async () => {
    signInMock.mockRejectedValue(new Error("boom"));

    const { props } = renderMethod();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });
    fireEvent.click(getSubmit());

    await waitFor(() => expect(props.onSignInError).toHaveBeenCalled());
    expect(props.onSignInError.mock.calls[0][0]).toBeInstanceOf(Error);
    expect(props.onSignInError.mock.calls[0][0].message).toBe("The recovery code might be wrong");
    expect(props.onSignInSuccess).not.toHaveBeenCalled();
    await waitFor(() => expect(getSubmit()).not.toBeDisabled());
});

it("submits when Enter is pressed", async () => {
    const { props } = renderMethod();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });
    fireEvent.keyDown(getInput(), { key: "Enter" });

    await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalledWith("https://example.com"));
    expect(signInMock).toHaveBeenCalledTimes(1);
});

it("does not submit when Enter is pressed without a code", () => {
    renderMethod();

    fireEvent.keyDown(getInput(), { key: "Enter" });

    expect(signInMock).not.toHaveBeenCalled();
});

it("ignores keys other than Enter", () => {
    renderMethod();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });
    fireEvent.keyDown(getInput(), { key: "a" });

    expect(signInMock).not.toHaveBeenCalled();
});

it("disables the buttons and ignores Enter while submitting", async () => {
    let resolve: (_value: any) => void = () => {};
    signInMock.mockImplementation(() => new Promise((r) => (resolve = r)));

    const { props } = renderMethod();

    fireEvent.change(getInput(), { target: { value: "AAAA-BBBB" } });
    fireEvent.click(getSubmit());

    await waitFor(() => expect(getSubmit()).toBeDisabled());
    expect(getCancel()).toBeDisabled();

    fireEvent.keyDown(getInput(), { key: "Enter" });

    expect(signInMock).toHaveBeenCalledTimes(1);

    resolve({ redirect: "https://example.com" });

    await waitFor(() => expect(props.onSignInSuccess).toHaveBeenCalledWith("https://example.com"));
    await waitFor(() => expect(getCancel()).not.toBeDisabled());
});

it("cancels back to the second factor", () => {
    const { props } = renderMethod();

    fireEvent.click(getCancel());

    expect(props.onCancel).toHaveBeenCalledTimes(1);
    expect(signInMock).not.toHaveBeenCalled();
});
