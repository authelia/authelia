import React from "react";

import { render } from "@testing-library/react";

import TimerIcon from "@components/TimerIcon";

it("renders without crashing", () => {
    render(<TimerIcon width={32} height={32} period={30} />);
});
