import { render } from "@testing-library/react";

import PushNotificationIcon from "@components/PushNotificationIcon";

it("renders without crashing", () => {
    render(<PushNotificationIcon width={32} height={32} />);
});

it("renders with animation", () => {
    render(<PushNotificationIcon width={32} height={32} animated />);
});
