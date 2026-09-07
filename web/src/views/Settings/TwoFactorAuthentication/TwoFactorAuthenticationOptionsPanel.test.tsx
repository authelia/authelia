import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { SecondFactorMethod } from "@models/Methods";
import { setPreferred2FAMethod } from "@services/UserInfo";
import TwoFactorAuthenticationOptionsPanel from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsPanel";

const mocks = vi.hoisted(() => ({
    createErrorNotification: vi.fn(),
    localStorageMethod: undefined as SecondFactorMethod | undefined,
    localStorageMethodAvailable: true,
    setLocalStorageMethod: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/LocalStorageMethodContext", () => ({
    useLocalStorageMethodContext: () => ({
        localStorageMethod: mocks.localStorageMethod,
        localStorageMethodAvailable: mocks.localStorageMethodAvailable,
        setLocalStorageMethod: mocks.setLocalStorageMethod,
    }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@services/UserInfo", async () => {
    const actual = await vi.importActual<typeof import("@services/UserInfo")>("@services/UserInfo");
    return { ...actual, setPreferred2FAMethod: vi.fn() };
});

vi.mock("@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsMethodsRadioGroup", () => ({
    default: (props: any) => (
        <div
            data-testid={`radio-group-${props.id}`}
            data-name={props.name}
            data-method={props.method}
            data-methods={props.methods.join(",")}
        >
            <button data-testid={`${props.id}-select-totp`} onClick={() => props.handleMethodChanged("totp")} />
            <button data-testid={`${props.id}-select-webauthn`} onClick={() => props.handleMethodChanged("webauthn")} />
            <button data-testid={`${props.id}-select-bogus`} onClick={() => props.handleMethodChanged("bogus")} />
        </div>
    ),
}));

const setPreferredMock = vi.mocked(setPreferred2FAMethod);

const config = {
    available_methods: new Set([SecondFactorMethod.TOTP, SecondFactorMethod.WebAuthn, SecondFactorMethod.MobilePush]),
} as any;

const info = {
    has_duo: false,
    has_totp: true,
    has_webauthn: true,
    method: SecondFactorMethod.TOTP,
} as any;

function renderPanel(props: Partial<{ config: any; info: any; refresh: () => void }> = {}) {
    const refresh = props.refresh ?? vi.fn();
    const merged = { config, info, ...props, refresh };

    return { ...render(<TwoFactorAuthenticationOptionsPanel {...merged} />), refresh };
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
    mocks.localStorageMethod = SecondFactorMethod.TOTP;
    mocks.localStorageMethodAvailable = true;
    setPreferredMock.mockResolvedValue(undefined as any);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("rendering", () => {
    it("renders options panel with radio groups", () => {
        renderPanel();
        expect(screen.getByText("Options")).toBeInTheDocument();
        expect(screen.getByTestId("radio-group-account")).toBeInTheDocument();
        expect(screen.getByTestId("radio-group-local")).toBeInTheDocument();
    });

    it("renders nothing when user has no methods", () => {
        const { container } = renderPanel({
            info: { ...info, has_duo: false, has_totp: false, has_webauthn: false },
        });
        expect(container.textContent).toBe("");
    });

    it("hides the browser group when local storage is unavailable", () => {
        mocks.localStorageMethodAvailable = false;

        renderPanel();

        expect(screen.queryByTestId("radio-group-local")).not.toBeInTheDocument();
        expect(screen.getByTestId("radio-group-account")).toBeInTheDocument();
    });

    it("hides the browser group when no browser method is stored", () => {
        mocks.localStorageMethod = undefined;

        renderPanel();

        expect(screen.queryByTestId("radio-group-local")).not.toBeInTheDocument();
    });

    it("names the two groups", () => {
        renderPanel();
        expect(screen.getByTestId("radio-group-account")).toHaveAttribute("data-name", "Default Method");
        expect(screen.getByTestId("radio-group-local")).toHaveAttribute("data-name", "Default Method (Browser)");
    });

    it("shows the current account method", () => {
        renderPanel({ info: { ...info, method: SecondFactorMethod.WebAuthn } });
        expect(screen.getByTestId("radio-group-account")).toHaveAttribute(
            "data-method",
            String(SecondFactorMethod.WebAuthn),
        );
    });

    it("shows the current browser method", () => {
        mocks.localStorageMethod = SecondFactorMethod.WebAuthn;

        renderPanel();

        expect(screen.getByTestId("radio-group-local")).toHaveAttribute(
            "data-method",
            String(SecondFactorMethod.WebAuthn),
        );
    });
});

describe("available methods", () => {
    it("only offers methods the user has registered", () => {
        renderPanel();

        expect(screen.getByTestId("radio-group-account")).toHaveAttribute(
            "data-methods",
            `${SecondFactorMethod.TOTP},${SecondFactorMethod.WebAuthn}`,
        );
    });

    it("offers mobile push when Duo is registered", () => {
        renderPanel({ info: { ...info, has_duo: true, has_webauthn: false } });

        expect(screen.getByTestId("radio-group-account")).toHaveAttribute(
            "data-methods",
            `${SecondFactorMethod.TOTP},${SecondFactorMethod.MobilePush}`,
        );
    });

    it("ignores methods that are not configured on the server", () => {
        renderPanel({ config: { available_methods: new Set([SecondFactorMethod.TOTP]) } });

        expect(screen.getByTestId("radio-group-account")).toHaveAttribute(
            "data-methods",
            String(SecondFactorMethod.TOTP),
        );
    });

    it("ignores unknown methods offered by the server", () => {
        renderPanel({ config: { available_methods: new Set([SecondFactorMethod.TOTP, 99]) } });

        expect(screen.getByTestId("radio-group-account")).toHaveAttribute(
            "data-methods",
            String(SecondFactorMethod.TOTP),
        );
    });
});

describe("changing the account method", () => {
    it("persists the new method and refreshes", async () => {
        const { refresh } = renderPanel();

        fireEvent.click(screen.getByTestId("account-select-webauthn"));

        await waitFor(() => expect(setPreferredMock).toHaveBeenCalledWith(SecondFactorMethod.WebAuthn));
        await waitFor(() => expect(refresh).toHaveBeenCalled());
        await waitFor(() =>
            expect(screen.getByTestId("radio-group-account")).toHaveAttribute(
                "data-method",
                String(SecondFactorMethod.WebAuthn),
            ),
        );
    });

    it("notifies and still refreshes when the update fails", async () => {
        setPreferredMock.mockRejectedValue(new Error("boom"));

        const { refresh } = renderPanel();

        fireEvent.click(screen.getByTestId("account-select-webauthn"));

        await waitFor(() =>
            expect(mocks.createErrorNotification).toHaveBeenCalledWith(
                "There was an issue updating preferred second factor method",
            ),
        );
        await waitFor(() => expect(refresh).toHaveBeenCalled());
    });

    it("ignores an unrecognized method", async () => {
        const { refresh } = renderPanel();

        fireEvent.click(screen.getByTestId("account-select-bogus"));

        await waitFor(() => expect(setPreferredMock).not.toHaveBeenCalled());
        expect(refresh).not.toHaveBeenCalled();
    });
});

describe("changing the browser method", () => {
    it("stores the new method locally", () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("local-select-webauthn"));

        expect(mocks.setLocalStorageMethod).toHaveBeenCalledWith(SecondFactorMethod.WebAuthn);
        expect(setPreferredMock).not.toHaveBeenCalled();
    });

    it("ignores an unrecognized method", () => {
        renderPanel();

        fireEvent.click(screen.getByTestId("local-select-bogus"));

        expect(mocks.setLocalStorageMethod).not.toHaveBeenCalled();
    });
});
