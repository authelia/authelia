
import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  container: {
    position: 'relative',
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
  },
  body: {
    paddingTop: theme.spacing.unit * 2,
    paddingBottom: theme.spacing.unit * 2,
  },
  image: {
    width: '120px',
  },
  imageContainer: {
    textAlign: 'center',
    marginTop: theme.spacing.unit * 2,
    marginBottom: theme.spacing.unit * 2,
  },
  footer: {
    fontSize: theme.typography.fontSize * 0.9,
  },
  registerDevice: {
    float: 'right',
  },
  totpField: {
    marginTop: theme.spacing.unit * 2,
    marginBottom: theme.spacing.unit * 2,
    width: '100%',
  },
  totpButton: {
    width: '100%',
  }
}));

export default styles;