import React from 'react';
import ReactDOM from 'react-dom';
import { mount, shallow } from "enzyme";
import { expect } from "chai";
import ColoredSnackbarContent from "./ColoredSnackbarContent";
import { SnackbarContent } from '@material-ui/core';

it('renders without crashing', () => {
    const div = document.createElement('div');
    ReactDOM.render(<ColoredSnackbarContent variant="success" message="this is a success" />, div);
    ReactDOM.unmountComponentAtNode(div);
});

it('should contain the message', () => {
    const el = mount(<ColoredSnackbarContent variant="success" message="this is a success" />);
    expect(el.text()).to.contain("this is a success");
});

it('should have correct color', () => {
    let el = shallow(<ColoredSnackbarContent variant="success" message="this is a success" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("success") > -1).to.be.true;

    el = shallow(<ColoredSnackbarContent variant="info" message="this is an info" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("info") > -1).to.be.true;

    el = shallow(<ColoredSnackbarContent variant="error" message="this is an error" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("error") > -1).to.be.true;

    el = shallow(<ColoredSnackbarContent variant="warning" message="this is an warning" />);
    expect(el.find(SnackbarContent).props().className!.indexOf("warning") > -1).to.be.true;
});