import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  fields: {
    marginBottom: theme.spacing.unit,
  },
  field: {
    paddingBottom: theme.spacing.unit * 2,
  },
  input: {
    width: '100%',
  },
  buttons: {
    '& button': {
      width: '100%',
    },
  },
  controls: {
    display: 'inline-block',
    width: '100%',
    fontSize: '0.875rem',
  },
  rememberMe: {
    float: 'left',
  },
  resetPassword: {
    padding: '12px 0px',
    float: 'right',
    '& a': {
      color: 'black',
    },
  },  
}));

export default styles;