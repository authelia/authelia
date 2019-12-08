import React from 'react';
import { mount } from "enzyme";
import Tracker from "./Tracker";

import { MemoryRouter as Router } from 'react-router-dom';

const mountWithRouter = node => mount(<Router>{node}</Router>);

it('renders without crashing', () => {
    mountWithRouter(<Tracker trackingIDs={[]} />);
});