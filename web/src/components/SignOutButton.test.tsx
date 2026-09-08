import { fireEvent, render, screen } from "@testing-library/react";
import { createInstance } from "i18next";
import { I18nextProvider } from "react-i18next";

import SignOutButton from "@components/SignOutButton";
import { TooltipProvider } from "@components/UI/Tooltip";

const mockDoSignOut = vi.fn();

vi.mock("@hooks/SignOut", () => ({
    useSignOut: vi.fn(() => mockDoSignOut),
}));

it("renders without crashing", () => {
    render(<SignOutButton id="test" text="Sign Out" />);
});

it("renders button with translated text", () => {
    render(<SignOutButton id="test" text="Sign Out" />);
    expect(screen.getByText("Sign Out")).toBeInTheDocument();
});

it("renders the translated tooltip when opened", async () => {
    const i18n = createInstance();

    await i18n.init({
        defaultNS: "portal",
        lng: "en",
        resources: { en: { portal: { "Sign out": "Sign out (translated)" } } },
    });

    render(
        <I18nextProvider i18n={i18n}>
            <TooltipProvider>
                <SignOutButton id="test" text="Sign Out" tooltip="Sign out" />
            </TooltipProvider>
        </I18nextProvider>,
    );

    const trigger = screen.getByRole("button");

    fireEvent.focus(trigger);
    fireEvent.pointerEnter(trigger);
    fireEvent.mouseEnter(trigger);

    expect(await screen.findByText("Sign out (translated)")).toBeInTheDocument();
});

it("calls sign out on click", () => {
    render(<SignOutButton id="test" text="Sign Out" />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockDoSignOut).toHaveBeenCalledWith(false);
});

it("calls sign out with preserve", () => {
    render(<SignOutButton id="test" text="Sign Out" preserve={true} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockDoSignOut).toHaveBeenCalledWith(true);
});
