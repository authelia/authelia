import { render, screen } from "@testing-library/react";

import SettingsView from "@views/Settings/SettingsView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders the settings overview", () => {
    render(<SettingsView />);
    expect(screen.getByText("User Settings")).toBeInTheDocument();
});
