import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  form: {
    marginTop: theme.spacing.unit * 2,
  },
  field: {
    width: '100%',
    marginBottom: theme.spacing.unit * 2,
  },
  button: {
    width: '100%',
  }
}));

export default styles;