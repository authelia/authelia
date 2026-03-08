import { render, screen } from "@testing-library/react";

import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { gridInner: "gridInner", gridOuter: "gridOuter" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("react-spinners/ScaleLoader", () => ({
    default: () => <div data-testid="scale-loader" />,
}));

vi.mock("@mui/material", async () => {
    const actual = await vi.importActual("@mui/material");
    return {
        ...actual,
        useTheme: () => ({
            custom: { loadingBar: "#000" },
            spacing: (n: number) => `${(n || 1) * 8}px`,
        }),
    };
});

it("renders the loading message", () => {
    render(<BaseLoadingPage message="Please wait" />);
    expect(screen.getByText("Please wait...")).toBeInTheDocument();
});

it("renders the scale loader", () => {
    render(<BaseLoadingPage message="Loading" />);
    expect(screen.getByTestId("scale-loader")).toBeInTheDocument();
});
