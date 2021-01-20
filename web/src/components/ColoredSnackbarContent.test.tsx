import React from "react";

import { SnackbarContent } from "@material-ui/core";
import { expect } from "chai";
import { mount, shallow } from "enzyme";
import ReactDOM from "react-dom";

import ColoredSnackbarContent from "./ColoredSnackbarContent";

it("renders without crashing", () => {
    const div = document.createElement("div");
    ReactDOM.render(<ColoredSnackbarContent level="success" message="this is a success" />, div);
    ReactDOM.unmountComponentAtNode(div);
});

it("should contain the message", () => {
    const el = mount(<ColoredSnackbarContent level="success" message="this is a success" />);
    expect(el.text()).to.contain("this is a success");
});

/* eslint-disable @typescript-eslint/no-unused-expressions */
it("should have correct color", () => {
    let el = shallow(<ColoredSnackbarContent level="success" message="this is a success" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("success") > -1).to.be.true;

    el = shallow(<ColoredSnackbarContent level="info" message="this is an info" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("info") > -1).to.be.true;

    el = shallow(<ColoredSnackbarContent level="error" message="this is an error" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("error") > -1).to.be.true;

    el = shallow(<ColoredSnackbarContent level="warning" message="this is an warning" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("warning") > -1).to.be.true;
});
/* eslint-enable @typescript-eslint/no-unused-expressions */
