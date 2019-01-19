import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  form: {
    paddingTop: theme.spacing.unit * 2,
  },
  field: {
    width: '100%',
  },
  buttonsContainer: {
    marginTop: theme.spacing.unit * 2,
    width: '100%',
  },
  buttonContainer: {
    width: '50%',
    display: 'inline-block',
    boxSizing: 'border-box',
  },
  buttonConfirmContainer: {
    paddingRight: theme.spacing.unit / 2,
  },
  buttonConfirm: {
    width: '100%',
  },
  buttonCancelContainer: {
    paddingLeft: theme.spacing.unit / 2,
  },
  buttonCancel: {
    width: '100%',
  },
}));

export default styles;