import { fireEvent, render, screen } from "@testing-library/react";

import WebAuthnCredentialItem from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialItem";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/CredentialItem", () => ({
    default: (props: any) => (
        <div
            data-testid="credential-item"
            data-id={props.id}
            data-description={props.description}
            data-qualifier={props.qualifier}
            data-problem={props.problem}
            data-last-used={props.last_used_at ? "yes" : "no"}
        >
            <button data-testid="information" onClick={() => props.handleInformation()} />
            <button data-testid="edit" onClick={() => props.handleEdit()} />
            <button data-testid="delete" onClick={() => props.handleDelete()} />
        </div>
    ),
}));

const credential = {
    attestation_format: "fido-u2f",
    attestation_type: "none",
    created_at: "2024-01-01T00:00:00Z",
    description: "Security Key",
    id: "abc123",
    legacy: false,
} as any;

it("renders credential item with correct description", () => {
    render(
        <WebAuthnCredentialItem
            index={0}
            credential={credential}
            handleInformation={vi.fn()}
            handleEdit={vi.fn()}
            handleDelete={vi.fn()}
        />,
    );
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-description", "Security Key");
});

it("passes legacy flag as problem prop", () => {
    render(
        <WebAuthnCredentialItem
            index={0}
            credential={{ ...credential, legacy: true }}
            handleInformation={vi.fn()}
            handleEdit={vi.fn()}
            handleDelete={vi.fn()}
        />,
    );
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-problem", "true");
});

function renderItem(index = 2, overrides: Record<string, unknown> = {}) {
    const handleInformation = vi.fn();
    const handleEdit = vi.fn();
    const handleDelete = vi.fn();

    render(
        <WebAuthnCredentialItem
            index={index}
            credential={{ ...credential, ...overrides }}
            handleInformation={handleInformation}
            handleEdit={handleEdit}
            handleDelete={handleDelete}
        />,
    );

    return { handleDelete, handleEdit, handleInformation };
}

it("builds an id from the index", () => {
    renderItem(3);
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-id", "webauthn-credential-3");
});

it("uppercases the attestation format as a qualifier", () => {
    renderItem();
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-qualifier", " (FIDO-U2F)");
});

it("omits the last used date when the credential has never been used", () => {
    renderItem(0, { last_used_at: undefined });
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-last-used", "no");
});

it("passes the last used date when the credential has been used", () => {
    renderItem(0, { last_used_at: "2024-06-01T00:00:00Z" });
    expect(screen.getByTestId("credential-item")).toHaveAttribute("data-last-used", "yes");
});

it("forwards the information request with the index", () => {
    const { handleInformation } = renderItem(4);

    fireEvent.click(screen.getByTestId("information"));

    expect(handleInformation).toHaveBeenCalledWith(4);
});

it("forwards the edit request with the index", () => {
    const { handleEdit } = renderItem(5);

    fireEvent.click(screen.getByTestId("edit"));

    expect(handleEdit).toHaveBeenCalledWith(5);
});

it("forwards the delete request with the index", () => {
    const { handleDelete } = renderItem(6);

    fireEvent.click(screen.getByTestId("delete"));

    expect(handleDelete).toHaveBeenCalledWith(6);
});
