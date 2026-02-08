import { render } from "@testing-library/react";

import App from "@root/App";
import "@i18n/index";

it("renders without crashing", () => {
    render(<App />);
});
