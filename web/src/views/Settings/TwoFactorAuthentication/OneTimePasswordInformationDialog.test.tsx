import { fireEvent, render, screen } from "@testing-library/react";

import OneTimePasswordInformationDialog from "@views/Settings/TwoFactorAuthentication/OneTimePasswordInformationDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@models/TOTPConfiguration", () => ({
    toAlgorithmString: (algo: number) => `algo-${algo}`,
}));

vi.mock("@i18n/formats", () => ({
    FormatDateHumanReadable: {},
}));

const config = {
    algorithm: 1,
    created_at: new Date("2024-01-01"),
    digits: 6,
    issuer: "Authelia",
    last_used_at: new Date("2024-06-01"),
    period: 30,
} as any;

it("renders config information when config is provided", () => {
    render(<OneTimePasswordInformationDialog config={config} open={true} handleClose={vi.fn()} />);
    expect(screen.getByText("One-Time Password Information")).toBeInTheDocument();
    expect(screen.getByText("Authelia")).toBeInTheDocument();
});

it("renders not loaded message when config is null", () => {
    render(<OneTimePasswordInformationDialog config={null} open={true} handleClose={vi.fn()} />);
    expect(screen.getByText("The One-Time Password information is not loaded")).toBeInTheDocument();
});

it("renders never used when last_used_at is undefined", () => {
    render(
        <OneTimePasswordInformationDialog
            config={{ ...config, last_used_at: undefined }}
            open={true}
            handleClose={vi.fn()}
        />,
    );
    expect(screen.getByText("Never")).toBeInTheDocument();
});

it("calls handleClose when close button is clicked", () => {
    const handleClose = vi.fn();
    render(<OneTimePasswordInformationDialog config={config} open={true} handleClose={handleClose} />);
    fireEvent.click(screen.getByText("Close"));
    expect(handleClose).toHaveBeenCalledOnce();
});
