import { configure } from "enzyme";
import Adapter from "enzyme-adapter-react-16";
document.body.setAttribute("data-basepath", "");
document.body.setAttribute("data-rememberme", "true");
document.body.setAttribute("data-resetpassword", "true");
document.body.setAttribute("data-theme-name", "light");
document.body.setAttribute("data-theme-primarycolor", "#1976d2");
document.body.setAttribute("data-theme-secondarycolor", "#ffffff");
configure({ adapter: new Adapter() });
