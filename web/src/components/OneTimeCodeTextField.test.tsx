import { render } from "@testing-library/react";

import OneTimeCodeTextField from "@components/OneTimeCodeTextField";

it("renders without crashing", () => {
    render(<OneTimeCodeTextField />);
});
