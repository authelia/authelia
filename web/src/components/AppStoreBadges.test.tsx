import React from "react";

import ReactDOM from "react-dom";

import AppStoreBadges from "./AppStoreBadges";

it("renders without crashing", () => {
    const div = document.createElement("div");
    ReactDOM.render(<AppStoreBadges iconSize={32} appleStoreLink="http://apple" googlePlayLink="http://google" />, div);
    ReactDOM.unmountComponentAtNode(div);
});
