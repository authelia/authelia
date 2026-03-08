import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import ConsentPortal from "@views/ConsentPortal/OpenIDConnect/ConsentPortal";

vi.mock("@constants/Routes", () => ({
    ConsentDecisionSubRoute: "/decision",
    ConsentOpenIDDeviceAuthorizationSubRoute: "/device-authorization",
}));

it("renders without crashing", () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    render(
        <MemoryRouter>
            <ConsentPortal state={{ authentication_level: 1 } as any} />
        </MemoryRouter>,
    );
});
