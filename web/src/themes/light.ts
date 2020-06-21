import { createMuiTheme } from '@material-ui/core/styles';

const light = createMuiTheme({
  palette: {
    primary: {
      main: '#1976d2', //default
    },
    background: {
      default: "#fff",
      paper: "#fff"
    },
  },
});

export default light;
