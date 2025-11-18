import { LinearProgress } from "@mui/material";

export interface Props {
    value: number;
    height?: number | string;
}

const LinearProgressBar = function (props: Props) {
    return (
        <LinearProgress
            variant="determinate"
            value={props.value}
            sx={{
                "& .MuiLinearProgress-determinate": {
                    transition: "transform .2s linear",
                },
                height: props.height ? props.height : (theme) => theme.spacing(),
                marginTop: (theme) => theme.spacing(),
            }}
        />
    );
};

export default LinearProgressBar;
