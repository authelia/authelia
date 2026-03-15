import { render } from "@testing-library/react";

import LinearProgressBar from "@components/LinearProgressBar";

it("renders without crashing", () => {
    render(<LinearProgressBar value={40} />);
});

it("renders with adjusted height", () => {
    render(<LinearProgressBar value={40} height={2} />);
});
