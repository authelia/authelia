import { render } from "@testing-library/react";

import PieChartIcon from "@components/PieChartIcon";

it("renders without crashing", () => {
    render(<PieChartIcon progress={40} />);
});

it("renders maxProgress without crashing", () => {
    render(<PieChartIcon progress={40} maxProgress={100} />);
});

it("renders width without crashing", () => {
    render(<PieChartIcon progress={40} width={20} />);
});

it("renders height without crashing", () => {
    render(<PieChartIcon progress={40} height={20} />);
});

it("renders color without crashing", () => {
    render(<PieChartIcon progress={40} color="black" />);
});

it("renders backgroundColor without crashing", () => {
    render(<PieChartIcon progress={40} backgroundColor="white" />);
});
