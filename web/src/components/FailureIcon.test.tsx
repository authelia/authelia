import { render } from "@testing-library/react";

import FailureIcon from "@components/FailureIcon";

it("renders without crashing", () => {
    render(<FailureIcon />);
});
