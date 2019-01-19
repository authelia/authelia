import { createStyles, Theme } from "@material-ui/core";

const borderColor = '#e0e0e0';

const styles = createStyles((theme: Theme) => ({
  secretContainer: {
    width: '100%',
    border: '1px solid ' + borderColor,
    marginTop: theme.spacing.unit * 2,
    marginBottom: theme.spacing.unit * 2,
  },
  qrcodeContainer: {
    textAlign: 'center',
    padding: theme.spacing.unit * 2,
  },
  base32Container: {
    textAlign: 'center',
    borderTop: '1px solid ' + borderColor,
    padding: theme.spacing.unit,
    wordWrap: 'break-word',
  },
  text: {
    textAlign: 'center',
  },
  needGoogleAuthenticator: {
    textAlign: 'center',
    marginTop: theme.spacing.unit * 2,
  },
  needGoogleAuthenticatorText: {
    fontSize: theme.typography.fontSize * 0.8,
  },
  store: {
    width: '100px',
    marginTop: theme.spacing.unit * 0.5,
    marginLeft: theme.spacing.unit * 0.5,
    marginRight: theme.spacing.unit * 0.5,
  },
  buttonContainer: {
    textAlign: 'center',
    paddingTop: theme.spacing.unit * 2,
  },
  progressContainer: {
    textAlign: 'center',
    paddingTop: theme.spacing.unit * 2,
  },
  button: {
    marginLeft: theme.spacing.unit,
    marginRight: theme.spacing.unit,
  },
  loginButtonContainer: {
    textAlign: 'center',
  },
}));

export default styles;