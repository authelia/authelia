import { render, screen } from "@testing-library/react";

import Authenticated from "@views/LoginPortal/Authenticated";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { iconContainer: "iconContainer" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

it("renders the authenticated stage with success icon", () => {
    render(<Authenticated />);
    expect(screen.getByText("Authenticated")).toBeInTheDocument();
    expect(screen.getByTestId("success-icon")).toBeInTheDocument();
});
