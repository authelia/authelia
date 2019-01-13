
import { createStyles, Theme } from "@material-ui/core";
import { isAbsolute } from "path";

const styles = createStyles((theme: Theme) => ({
  container: {
    position: 'relative',
  },
  hello: {},
  logout: {},
  header: {
    fontSize: theme.typography.fontSize * 1.5,
    marginBottom: theme.spacing.unit,
    position: 'relative',
    '& $hello': {
      display: 'inline-block',
    },
    '& $logout': {
      position: 'absolute',
      bottom: '0px',
      right: '0px',
      fontSize: theme.typography.fontSize * 0.9,
    },
  },
  body: {
    paddingTop: theme.spacing.unit * 2,
    paddingBottom: theme.spacing.unit * 2,
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
    border: '1px solid #e0e0e0',
    borderRadius: '2px',
    textAlign: 'justify',
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
    paddingTop: theme.spacing.unit,
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