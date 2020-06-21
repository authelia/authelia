import { createMuiTheme } from '@material-ui/core/styles';

const dark = createMuiTheme({
  palette: {
    primary: {
      main: '#fff', //black
    },
    background: {
      default: "#000",
      paper: "#000"
    },
  },
  overrides: {
    MuiCssBaseline: {
      "@global": {
        body: {
          backgroundColor: "#000",
          color: "#fff",
        },
      }
    },
    MuiOutlinedInput: {
      root: {

      },
    },
    MuiCheckbox: {
      root: {
        color: "#fff"
      },
    },
    MuiInputBase: {
      input: {
        color: "#fff"
      }
    },
    MuiInputLabel: {
      root: {
        color: "#fff"
      }
    }
  }
});

export default dark;
