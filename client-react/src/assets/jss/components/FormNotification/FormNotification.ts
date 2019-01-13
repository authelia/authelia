import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  messageOuter: {
    position: 'relative',
    paddingTop: theme.spacing.unit * 2,
    paddingBottom: theme.spacing.unit,
  },
  messageInner: {
    width: '100%',
  },
  messageContainer: {
    color: 'white',
    fontSize: theme.typography.fontSize * 0.9,
    padding: theme.spacing.unit * 2,
    border: '1px solid red',
    borderRadius: '5px',
    backgroundColor: '#ff8d8d',
  },
}));

export default styles;