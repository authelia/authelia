import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  container: {
    textAlign: 'center',
  },
  messageContainer: {
    fontSize: theme.typography.fontSize,
    marginTop: theme.spacing.unit * 2,
    marginBottom: theme.spacing.unit * 2,
    color: 'green',
    display: 'inline-block',
    marginLeft: theme.spacing.unit * 2,
    textAlign: 'left',

  },
  successContainer: {
    verticalAlign: 'middle',
    paddingTop: theme.spacing.unit * 2,
    paddingBottom: theme.spacing.unit * 2,
    marginTop: theme.spacing.unit * 2,
    marginBottom: theme.spacing.unit * 3,
    border: '1px solid #8ae48a',
    borderRadius: '100px',
  },
  successLogoContainer: {
    display: 'inline-block',
  },
  logoutButtonContainer: {
    marginTop: theme.spacing.unit * 2,
  },
}));

export default styles;