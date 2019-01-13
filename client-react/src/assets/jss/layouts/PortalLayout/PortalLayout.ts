import { createStyles, Theme } from "@material-ui/core";

const styles = createStyles((theme: Theme) => ({
  mainContent: {
    width: '440px',
    margin: '0 auto',
    padding: '50px 0px',
  },
  frame: {
    boxShadow: 'rgba(0,0,0,0.14902) 0px 1px 1px 0px,rgba(0,0,0,0.09804) 0px 1px 2px 0px',
    backgroundColor: 'white',
    borderRadius: '5px',
    padding: '30px 40px',
  },
  innerFrame: {
    width: '100%',
  },
  title: {
    fontSize: '1.4em',
    fontWeight: 'bold',
    borderBottom: '1px solid #c7c7c7',
    display: 'inline-block',
    paddingRight: '10px',
    paddingBottom: '5px',
  },
  content: {
    paddingTop: theme.spacing.unit * 2,
    paddingBottom: theme.spacing.unit,
  },
  footer: {
    marginTop: '10px',
    textAlign: 'center',
    fontSize: '0.65em',
    color: 'grey',
    '& a': {
      color: 'grey',
    }
  },
}));

export default styles;