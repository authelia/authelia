import { createMuiTheme } from '@material-ui/core/styles';
import { useMainColor, useSecondaryColor } from '../hooks/Theme';

const custom = createMuiTheme({
  palette: {
    primary: {
      main: useMainColor(), //dark grey
    },
    background: {
      default: useSecondaryColor(),
      paper: useSecondaryColor()
    },
  },
  overrides: {
    MuiCssBaseline: {
      "@global": {
        body: {
          backgroundColor: useSecondaryColor(),
          color: useMainColor(),
        },
      }
    },
    MuiOutlinedInput: {
      root: {
        "& $notchedOutline": {
          borderColor: useMainColor()
        },
        "&:hover:not($disabled):not($focused):not($error) $notchedOutline": {
          borderColor: useMainColor(),
          borderWidth: 2
        },
        "&$focused $notchedOutline": {
          borderColor: useMainColor()
        },
      },
      notchedOutline: {}
    },
    MuiCheckbox: {
      root: {
        color: useMainColor()
      },
    },
    MuiInputBase: {
      input: {
        color: useMainColor()
      }
    },
    MuiInputLabel: {
      root: {
        color: useMainColor()
      }
    },
    MuiButton: {
      label: {
        color: useSecondaryColor()
      }
    }
  }
});

export default custom;
