import { configure } from 'enzyme';
import Adapter from 'enzyme-adapter-react-16';
document.body.setAttribute("data-basepath", "");
document.body.setAttribute("data-rememberme", "false");
document.body.setAttribute("data-disable-resetpassword", "false");
configure({ adapter: new Adapter() });
