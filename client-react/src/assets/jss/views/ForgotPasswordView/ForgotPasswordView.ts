import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  form: {
    paddingTop: theme.spacing.unit * 2,
  },
  field: {
    width: '100%',
  },
  button: {
    marginTop: theme.spacing.unit * 2,
    width: '100%',
  }
}));

export default styles;