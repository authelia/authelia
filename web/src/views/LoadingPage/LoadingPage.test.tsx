import { render, screen } from "@testing-library/react";

import LoadingPage from "@views/LoadingPage/LoadingPage";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@views/LoadingPage/BaseLoadingPage", () => ({
    default: (props: any) => <div data-testid="base-loading">{props.message}</div>,
}));

it("renders with translated loading message", () => {
    render(<LoadingPage />);
    expect(screen.getByTestId("base-loading")).toHaveTextContent("Loading");
});
