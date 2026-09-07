import { act, render } from "@testing-library/react";
import axios from "axios";

import App from "@root/App";
import "@i18n/index";

it("renders without crashing", async () => {
    vi.spyOn(axios, "get").mockResolvedValue({ data: { data: {}, status: "OK" }, status: 200 });

    await act(async () => {
        render(<App />);
    });
});
