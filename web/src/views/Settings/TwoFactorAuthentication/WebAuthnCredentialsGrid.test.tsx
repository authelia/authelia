import { render, screen } from "@testing-library/react";

import WebAuthnCredentialsGrid from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsGrid";

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialItem", () => ({
    default: (props: any) => <div data-testid={`credential-${props.index}`}>{props.credential.description}</div>,
}));

it("renders credential items for each credential", () => {
    const credentials = [
        { description: "Key 1", id: "1" },
        { description: "Key 2", id: "2" },
    ] as any[];

    render(
        <WebAuthnCredentialsGrid
            credentials={credentials}
            handleInformation={vi.fn()}
            handleEdit={vi.fn()}
            handleDelete={vi.fn()}
        />,
    );

    expect(screen.getByTestId("credential-0")).toHaveTextContent("Key 1");
    expect(screen.getByTestId("credential-1")).toHaveTextContent("Key 2");
});

it("renders empty grid when no credentials", () => {
    const { container } = render(
        <WebAuthnCredentialsGrid
            credentials={[]}
            handleInformation={vi.fn()}
            handleEdit={vi.fn()}
            handleDelete={vi.fn()}
        />,
    );
    expect(container.querySelector("[data-testid^='credential-']")).toBeNull();
});
