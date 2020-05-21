import { configure } from 'enzyme';
import Adapter from 'enzyme-adapter-react-16';
document.body.setAttribute("data-basepath", "");
configure({ adapter: new Adapter() });
