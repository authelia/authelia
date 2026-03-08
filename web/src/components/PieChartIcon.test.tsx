import { render } from "@testing-library/react";

import PieChartIcon from "@components/PieChartIcon";

it("renders without crashing", () => {
    render(<PieChartIcon progress={40} />);
});

it("renders with maxProgress", () => {
    render(<PieChartIcon progress={40} maxProgress={100} />);
});

it("renders with width", () => {
    render(<PieChartIcon progress={40} width={20} />);
});

it("renders with height", () => {
    render(<PieChartIcon progress={40} height={20} />);
});

it("renders with color", () => {
    render(<PieChartIcon progress={40} color="black" />);
});

it("renders with backgroundColor", () => {
    render(<PieChartIcon progress={40} backgroundColor="white" />);
});
