import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  form: {
    marginTop: theme.spacing.unit * 2,
  },
  field: {
    width: '100%',
    marginBottom: theme.spacing.unit * 2,
  },
  buttonsContainer: {
    width: '100%',
  },
  buttonContainer: {
    width: '50%',
    boxSizing: 'border-box',
    display: 'inline-block',
  },
  buttonResetContainer: {
    paddingRight: theme.spacing.unit / 2,
  },
  buttonCancelContainer: {
    paddingLeft: theme.spacing.unit / 2,
  },
  button: {
    width: '100%',
  }
}));

export default styles;