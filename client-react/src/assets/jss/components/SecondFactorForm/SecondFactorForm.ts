
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
    paddingLeft: theme.spacing.unit * 2,
    paddingRight: theme.spacing.unit * 2,
    border: '1px solid #e0e0e0',
    borderRadius: '2px',
  },
  methodName: {
    fontSize: theme.typography.fontSize * 1.2,
    fontWeight: 'bold',
    marginBottom: theme.spacing.unit,
  },
  methodU2f: {
    borderBottom: '1px solid #e0e0e0',
    padding: theme.spacing.unit,
  },
  methodTotp: {
    padding: theme.spacing.unit,
    paddingTop: theme.spacing.unit * 2,
  },
  image: {
    width: '120px',
  },
  imageContainer: {
    textAlign: 'center',
    marginTop: theme.spacing.unit * 2,
    marginBottom: theme.spacing.unit * 2,
  },
  registerDeviceContainer: {
    textAlign: 'right',
    fontSize: theme.typography.fontSize * 0.8,
  },
  registerDevice: {},
  totpField: {
    marginTop: theme.spacing.unit * 2,
    width: '100%',
  },
  totpButton: {
    marginTop: theme.spacing.unit * 2,
    width: '100%',
  }
}));

export default styles;