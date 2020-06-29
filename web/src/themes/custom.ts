import { createMuiTheme } from '@material-ui/core/styles';
import { usePrimaryColor, useSecondaryColor } from '../hooks/Theme';

const custom = createMuiTheme({
  palette: {
    primary: {
      main: usePrimaryColor(), //dark grey
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
          color: usePrimaryColor(),
        },
      }
    },
    MuiOutlinedInput: {
      root: {
        "& $notchedOutline": {
          borderColor: usePrimaryColor()
        },
        "&:hover:not($disabled):not($focused):not($error) $notchedOutline": {
          borderColor: usePrimaryColor(),
          borderWidth: 2
        },
        "&$focused $notchedOutline": {
          borderColor: usePrimaryColor()
        },
      },
      notchedOutline: {}
    },
    MuiCheckbox: {
      root: {
        color: usePrimaryColor()
      },
    },
    MuiInputBase: {
      input: {
        color: usePrimaryColor()
      }
    },
    MuiInputLabel: {
      root: {
        color: usePrimaryColor()
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
